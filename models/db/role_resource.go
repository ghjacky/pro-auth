package db

import (
	"fmt"
	"strconv"

	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/models/lock"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type RoleResource struct {
	Id       int64     `orm:"pk;auto" json:"-"`
	RoleId   int64     `orm:"index" json:"role_id"`
	Resource *Resource `orm:"rel(fk)" json:"resource"`
}

type RoleResourceJSON struct {
	RoleId     int64 `json:"role_id"`
	ResourceId int64 `json:"resource_id"`
}

func GetRoleResources(clientId, roleId int64) ([]*RoleResourceJSON, error) {
	var roleResource []*RoleResourceJSON
	o := orm.NewOrm()
	sql := "select * from role_resource where role_id in (select role_id from role where client_id = ? )"
	if roleId > 0 {
		sql += fmt.Sprintf(" and role_id = %d", roleId)
	}
	if _, err := o.Raw(sql, clientId).QueryRows(&roleResource); err != nil {
		return nil, err
	} else if len(roleResource) == 0 {
		return []*RoleResourceJSON{}, nil
	} else {
		return roleResource, nil
	}
}

func UpdateRoleResource(ids []int64, clientId, roleId int64, isAll bool) (int64, error) {
	lock := lock.NewAppLock(clientId)
	lock.Lock()
	defer lock.UnLock()
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return 0, err
	}
	defer o.Commit()
	res := map[string]int64{
		"del_role_resource_num": 0,
		"add_role_resource_num": 0,
	}
	if isAll {
		/**
		允许添加client下全部资源
		1.查询client下全部资源，判断ids是否合法
		2.查询role已有的资源，计算删除关联和新增关联
		3.查询全局role，判断role是否存在，计算祖先集合和后代集合
		4.祖先集合-新增关联    后代集合-删除关联
		*/
		validIdMap := make(orm.Params)
		var roles []*Role
		if _, err := o.Raw(fmt.Sprintf("select id, name from resource  where client_id = %d", clientId)).
			RowsToMap(&validIdMap, "id", "name"); err != nil {
			logs.Error(err)
			return 0, err
		}
		logs.Info(ids)
		logs.Info(validIdMap)
		for _, v := range ids {
			if _, ok := validIdMap[strconv.FormatInt(v, 10)]; !ok {
				return 0, errors.ErrResourceNotFound
			}
		}
		if _, err := o.QueryTable("role").Filter("client_id", clientId).All(&roles); err != nil {
			if err == orm.ErrNoRows {
				return 0, errors.ErrRoleNotFound
			} else {
				logs.Error(err)
				return 0, err
			}
		}
		roleMap := make(map[int64]*Role)
		for i := range roles {
			roleMap[roles[i].Id] = roles[i]
		}
		role, ok := roleMap[roleId]
		if !ok {
			return 0, errors.ErrRoleNotFound
		} else if role.ParentId == -1 {
			return 0, errors.ErrRoleResourceCannotModify
		}
		ancestorIds := []int64{roleId}
		offspringIds := []int64{roleId}
		BuildRoleTree(roles)
		parentId := role.ParentId
		for {
			if r, ok := roleMap[parentId]; ok && r.ParentId != -1 {
				ancestorIds = append(ancestorIds, r.Id)
				parentId = r.ParentId
			} else {
				break
			}
		}
		queue := role.Children
		for {
			if len(queue) == 0 {
				break
			}
			r := queue[0]
			queue = queue[1:]
			offspringIds = append(offspringIds, r.Id)
			if len(r.Children) > 0 {
				queue = append(queue, r.Children...)
			}
		}
		logs.Info("祖先：%+v", ancestorIds)
		logs.Info("后代：%+v", offspringIds)
		var roleResources []*RoleResource
		var addResourceId []int64
		var delResourceId []int64
		existedRelMap := make(map[int64]map[int64]int64)
		if _, err := o.QueryTable("role_resource").Filter("role_id__in", ancestorIds).All(&roleResources); err != nil && err != orm.ErrNoRows {
			logs.Error(err)
			return 0, err
		}
		for _, v := range roleResources {
			if r, ok := existedRelMap[v.RoleId]; !ok {
				r = make(map[int64]int64)
				r[v.Resource.Id] = v.Resource.Id
				existedRelMap[v.RoleId] = r
				logs.Info(existedRelMap)
			} else {
				r[v.Resource.Id] = v.Resource.Id
			}
		}
		if updateRoleMap, ok := existedRelMap[roleId]; ok {
			for _, v := range ids {
				if _, ok := updateRoleMap[v]; !ok {
					addResourceId = append(addResourceId, v)
				} else {
					delete(updateRoleMap, v)
				}
			}
			for k := range updateRoleMap {
				delResourceId = append(delResourceId, k)
			}
		} else {
			addResourceId = ids
		}
		logs.Info("新增资源：%+v", addResourceId)
		logs.Info("删除资源：%+v", delResourceId)
		if len(delResourceId) > 0 {
			if num, err := o.QueryTable("role_resource").Filter("role_id__in", offspringIds).Filter("resource_id__in", delResourceId).Delete(); err != nil && err != orm.ErrNoRows {
				o.Rollback()
				return 0, err
			} else {
				res["del_role_resource_num"] = num
			}
		}
		sql := "insert into role_resource (role_id, resource_id) values "
		count := 0
		for _, roleId := range ancestorIds {
			if rMap, ok := existedRelMap[roleId]; ok {
				for _, resourceId := range addResourceId {
					if _, ok := rMap[resourceId]; !ok {
						count++
						sql += fmt.Sprintf("(%d,%d),", roleId, resourceId)
					}
				}
			} else {
				for _, resourceId := range addResourceId {
					count++
					sql += fmt.Sprintf("(%d,%d),", roleId, resourceId)
				}
			}
		}
		if count == 0 {
			return 0, nil
		} else {
			if r, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
				o.Rollback()
				return 0, err
			} else {
				num, _ := r.RowsAffected()
				res["add_role_resource_num"] = num
				return num, nil
			}
		}

	} else {
		/**
		默认只能添加父角色中有的资源
		1.判断role是否存在，生成后代集合
		2.查询parent资源范围，判断ids是否合法，拆分为新增关联和删除关联
		3.该角色添加关联，后代集合删除关联
		*/
		var roles []*Role
		var role Role
		var offspringIds []int64

		if _, err := o.QueryTable("role").Filter("client_id", clientId).All(&roles); err != nil {
			if err == orm.ErrNoRows {
				return 0, errors.ErrRoleNotFound
			} else {
				logs.Error(err)
				return 0, err
			}
		}
		queue := BuildRoleTree(roles)
		find := false
		for {
			if len(queue) == 0 {
				break
			}
			logs.Info(queue)
			r := queue[0]
			queue = queue[1:]
			if find {
				offspringIds = append(offspringIds, r.Id)
				if len(r.Children) > 0 {
					queue = append(queue, r.Children...)
				}
			} else if r.Id == roleId {
				if r.ParentId == -1 {
					return 0, errors.ErrRoleResourceCannotModify
				}
				find = true
				role = *r
				offspringIds = append(offspringIds, r.Id)
				queue = r.Children
			} else if len(r.Children) > 0 {
				queue = append(queue, r.Children...)
			}
		}
		logs.Info("后代：%+v", offspringIds)
		if !find {
			return 0, errors.ErrRoleNotFound
		}
		validIdMap := make(orm.Params)
		if _, err := o.Raw(fmt.Sprintf("select resource_id as id, resource_id as name from role_resource  where role_id = %d ", role.ParentId)).
			RowsToMap(&validIdMap, "id", "name"); err != nil {
			logs.Error(err)
			return 0, err
		}
		for _, v := range ids {
			if _, ok := validIdMap[strconv.FormatInt(v, 10)]; !ok {
				return 0, errors.ErrResourceNotFound
			}
		}
		var addResourceId []int64
		var delResourceId []int64
		existedRelMap := make(orm.Params)
		if _, err := o.Raw(fmt.Sprintf("select resource_id as id, resource_id as name from role_resource where role_id = %d", roleId)).RowsToMap(&existedRelMap, "id", "name"); err != nil && err != orm.ErrNoRows {
			logs.Error(err)
			return 0, err
		}
		for _, v := range ids {
			k := strconv.FormatInt(v, 10)
			if _, ok := existedRelMap[k]; ok {
				delete(existedRelMap, k)
			} else {
				addResourceId = append(addResourceId, v)
			}
		}
		for k := range existedRelMap {
			id, _ := strconv.ParseInt(k, 10, 64)
			delResourceId = append(delResourceId, id)
		}
		if len(addResourceId) > 0 {
			sql := "insert into role_resource (role_id, resource_id) values "
			for _, resourceId := range addResourceId {
				sql += fmt.Sprintf("(%d,%d),", roleId, resourceId)
			}
			if r, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
				o.Rollback()
				return 0, err
			} else {
				num, _ := r.RowsAffected()
				res["add_role_resource_num"] = num
			}
		}

		if len(delResourceId) > 0 {
			if num, err := o.QueryTable("role_resource").Filter("role_id__in", offspringIds).Filter("resource_id__in", delResourceId).Delete(); err != nil && err != orm.ErrNoRows {
				o.Rollback()
				return 0, err
			} else {
				res["del_role_resource_num"] = num
			}
		}
		return 0, nil

	}

}

func AddRoleResource(ids []int64, clientId, roleId int64, isAll bool) (int64, error) {

	lock := lock.NewAppLock(clientId)
	lock.Lock()
	defer lock.UnLock()
	o := orm.NewOrm()
	if isAll {
		/**
		允许添加client下全部资源
		1.查询client下全部资源，判断ids是否合法
		2.查询全局role，判断role是否存在，并找出role的parents
		3.查询parents已经关联的资源，将ids中没有关联的与parents关联，将ids与role关联
		4. 不严格限制关联关系不存在，允许添加成功0个关系
		*/
		var roles []*Role
		var parentIds []int64
		validIdMap := make(orm.Params)
		roleMap := make(map[int64]*Role)
		if _, err := o.Raw(fmt.Sprintf("select id, name from resource  where client_id = %d", clientId)).
			RowsToMap(&validIdMap, "id", "name"); err != nil {
			logs.Error(err)
			return 0, err
		}
		for _, v := range ids {
			if _, ok := validIdMap[strconv.FormatInt(v, 10)]; !ok {
				return 0, errors.ErrResourceNotFound
			}
		}
		if _, err := o.QueryTable("role").Filter("client_id", clientId).All(&roles); err != nil {
			if err == orm.ErrNoRows {
				return 0, errors.ErrRoleNotFound
			} else {
				logs.Error(err)
				return 0, err
			}
		}
		for i := range roles {
			roleMap[roles[i].Id] = roles[i]
		}
		role, ok := roleMap[roleId]
		if !ok {
			return 0, errors.ErrRoleNotFound
		} else if role.ParentId == -1 {
			return 0, errors.ErrRoleResourceCannotModify
		}
		parentId := role.ParentId
		parentIds = append(parentIds, role.Id)
		for {
			if r, ok := roleMap[parentId]; ok && r.ParentId != -1 {
				parentIds = append(parentIds, r.Id)
				parentId = r.ParentId
			} else {
				break
			}
		}
		if _, ok := roleMap[roleId]; !ok {
			return 0, errors.ErrRoleNotFound
		}
		sql := "insert into role_resource (role_id, resource_id) values "
		if len(parentIds) > 0 {
			var roleResources []*RoleResource
			existedMap := make(map[int64]map[int64]int64)
			if _, err := o.QueryTable("role_resource").Filter("role_id__in", parentIds).All(&roleResources); err != nil && err != orm.ErrNoRows {
				logs.Error(err)
				return 0, err
			}
			for _, v := range roleResources {
				if r, ok := existedMap[v.RoleId]; !ok {
					r = make(map[int64]int64)
					r[v.Resource.Id] = v.Resource.Id
					existedMap[v.RoleId] = r
					logs.Info(existedMap)
				} else {
					r[v.Resource.Id] = v.Resource.Id
				}
			}
			count := 0
			for _, roleId := range parentIds {
				if rMap, ok := existedMap[roleId]; ok {
					for _, resourceId := range ids {
						if _, ok := rMap[resourceId]; !ok {
							count++
							sql += fmt.Sprintf("(%d,%d),", roleId, resourceId)
						}
					}
				} else {
					for _, resourceId := range ids {
						count++
						sql += fmt.Sprintf("(%d,%d),", roleId, resourceId)
					}
				}

			}
			if count == 0 {
				return 0, nil
			}
		}
		if res, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
			return 0, err
		} else {
			num, _ := res.RowsAffected()
			return num, nil
		}

	} else {
		/**
		默认只能添加父角色中有的资源
		1.判断role是否存在
		2.查询parent资源范围，判断ids是否合法
		3.添加关联
		4.严格限制关联关系不存在，必须添加成功len(ids)个关系
		*/
		var role Role
		if err := o.QueryTable("role").Filter("id", roleId).Filter("client_id", clientId).One(&role); err != nil {
			if err == orm.ErrNoRows {
				return 0, errors.ErrRoleNotFound
			} else {
				logs.Error(err)
				return 0, err
			}
		}
		validIdMap := make(orm.Params)
		if _, err := o.Raw(fmt.Sprintf("select resource_id, role_id from role_resource where role_id = %d", role.ParentId)).
			RowsToMap(&validIdMap, "resource_id", "role_id"); err != nil {
			logs.Error(err)
			return 0, err
		}
		for _, v := range ids {
			if _, ok := validIdMap[strconv.FormatInt(v, 10)]; !ok {
				return 0, errors.ErrResourceNotFound
			}
		}
		sql := "insert into role_resource (role_id, resource_id) values "
		for _, v := range ids {
			sql += fmt.Sprintf("(%d,%d),", roleId, v)
		}
		if res, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
			return 0, err
		} else {
			num, _ := res.RowsAffected()
			return num, nil
		}
	}

}

func DeleteRoleResource(ids []int64, clientId, roleId int64) (int64, error) {
	/**
	1、判断role存在
	2、删除role_resource
	*/
	o := orm.NewOrm()
	var roles []*Role
	var childIds []int64
	if _, err := o.QueryTable("role").Filter("client_id", clientId).All(&roles); err != nil {
		if err == orm.ErrNoRows {
			return 0, errors.ErrRoleNotFound
		} else {
			logs.Error(err)
			return 0, err
		}
	}
	queue := BuildRoleTree(roles)
	find := false
	for {
		if len(queue) == 0 {
			break
		}
		logs.Info(queue)
		r := queue[0]
		queue = queue[1:]
		if find {
			childIds = append(childIds, r.Id)
			if len(r.Children) > 0 {
				queue = append(queue, r.Children...)
			}
		} else if r.Id == roleId {
			if r.ParentId == -1 {
				return 0, errors.ErrRoleResourceCannotModify
			}
			find = true
			childIds = append(childIds, r.Id)
			queue = r.Children
		} else if len(r.Children) > 0 {
			queue = append(queue, r.Children...)
		}
		logs.Info(childIds)
	}
	if !find {
		return 0, errors.ErrRoleNotFound
	}
	if num, err := o.QueryTable("role_resource").Filter("role_id__in", childIds).Filter("resource_id__in", ids).Delete(); err != nil && err != orm.ErrNoRows {
		return 0, err
	} else {
		return num, nil
	}

}
