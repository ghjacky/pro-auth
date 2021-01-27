package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/models/lock"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type Role struct {
	Id          int64           `orm:"pk;auto" json:"id"`
	Name        string          `orm:"size(255)"json:"name"`
	Description string          `orm:"size(255)" json:"description"`
	ParentId    int64           `orm:"index" json:"parent_id"`
	ClientId    int64           `orm:"index" json:"client_id"`
	CreatedBy   string          `orm:"size(255)" json:"created_by"`
	UpdatedBy   string          `orm:"size(255)" json:"updated_by"`
	Created     time.Time       `orm:"auto_now_add;type(datetime)"json:"created"`
	Updated     time.Time       `orm:"auto_now;type(datetime)" json:"updated"`
	RoleType    string          `orm:"-" json:"role_type,omitempty"`
	Resources   []*RoleResource `orm:"-" json:"resources,omitempty"`
	Users       []*RoleUser     `orm:"-" json:"users,omitempty"`
	Children    []*Role         `orm:"-" json:"children,omitempty"`
}

type RoleJson struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentId    int64     `json:"parent_id"`
	ClientId    int64     `json:"client_id"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	RoleType    string    `json:"role_type"`
}

func AddSubRole(role Role) (int64, error) {
	/*
		1、加锁
		2、判断父角色是否存在
		3、判断父角色和该角色appId是否相同
		4、解锁
	*/
	lock := lock.NewAppLock(role.ClientId)
	lock.Lock()
	defer lock.UnLock()
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return 0, err
	}
	defer o.Commit()
	var parentRole Role
	if err := o.QueryTable("role").
		Filter("id", role.ParentId).Filter("client_id", role.ClientId).One(&parentRole); err != nil {
		return 0, errors.ErrParentRoleNotFound
	}
	if id, err := o.Insert(&role); err != nil {
		o.Rollback()
		logs.Error(err)
		return 0, err
	} else {
		return id, nil
	}

}

func GetRole(roleId int64) (*Role, error) {
	var role = Role{}
	o := orm.NewOrm()
	if err := o.QueryTable("role").Filter("id", roleId).One(&role); err != nil {
		return nil, errors.ErrRoleNotFound
	}
	return &role, nil
}

func GetRolesByParentIDs(roleIDs []int64) ([]*Role, error) {
	if len(roleIDs) == 0 {
		return []*Role{}, nil
	}
	var roles = []*Role{}
	o := orm.NewOrm()
	if _, err := o.QueryTable("role").Filter("parent_id__in", roleIDs).RelatedSel().All(&roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func GetRoleUserByRoleIDs(roleIDs []int64) ([]*RoleUser, error) {
	if len(roleIDs) == 0 {
		return []*RoleUser{}, nil
	}

	o := orm.NewOrm()
	var roleUsers []*RoleUser
	if _, err := o.QueryTable("role_user").Filter("role_id__in", roleIDs).RelatedSel().All(&roleUsers); err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	return roleUsers, nil
}

func GetRoleUserByUserIDs(userIDs []string) ([]*SimpleRoleUser, error) {
	if len(userIDs) == 0 {
		return []*SimpleRoleUser{}, nil
	}

	o := orm.NewOrm()
	var roleUsers []*RoleUser
	if _, err := o.QueryTable("role_user").Filter("user_id__in", userIDs).All(&roleUsers); err != nil && err != orm.ErrNoRows {
		return nil, err
	}

	simpleRoleUsers := make([]*SimpleRoleUser, len(roleUsers))
	for index, ru := range roleUsers {
		simpleRoleUsers[index] = &SimpleRoleUser{
			RoleId:   ru.RoleId,
			User:     ru.User.Id,
			RoleType: ru.RoleType,
		}
	}
	return simpleRoleUsers, nil
}

func GetRoleByIDs(roleIDs []int64) ([]*Role, error) {
	if len(roleIDs) == 0 {
		return []*Role{}, nil
	}
	var roles = []*Role{}
	o := orm.NewOrm()
	if _, err := o.QueryTable("role").Filter("id__in", roleIDs).RelatedSel().All(&roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func GetRoleResourceByRoleIDs(roleIDs []int64) ([]*RoleResource, error) {
	if len(roleIDs) == 0 {
		return []*RoleResource{}, nil
	}

	o := orm.NewOrm()
	var roleResources []*RoleResource
	if _, err := o.QueryTable("role_resource").Filter("role_id__in", roleIDs).RelatedSel().All(&roleResources); err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	return roleResources, nil
}

func RelateUserResourceForRole(roles []*Role, relateUser, relateResource bool) ([]*Role, error) {
	if !relateUser && !relateResource {
		return roles, nil
	}

	loadCount := 0
	ch := make(chan int, 2)
	var roleIds []int64
	var roleMap = make(map[int64]*Role)

	for i := range roles {
		roleIds = append(roleIds, roles[i].Id)
		roleMap[roles[i].Id] = roles[i]
	}

	loadUserFunc := func(roles []*Role, ch chan<- int) {
		if roleUsers, err := GetRoleUserByRoleIDs(roleIds); err != nil {
			logs.Error(err)
		} else if len(roleUsers) > 0 {
			for i := range roleUsers {
				if role, ok := roleMap[roleUsers[i].RoleId]; ok {
					role.Users = append(role.Users, roleUsers[i])
				}
			}
		}
		ch <- 1
	}
	loadResourceFunc := func(roles []*Role, ch chan<- int) {
		if roleResources, err := GetRoleResourceByRoleIDs(roleIds); err != nil {
			logs.Error(err)
		} else if len(roleResources) > 0 {
			for i := range roleResources {
				if role, ok := roleMap[roleResources[i].RoleId]; ok {
					role.Resources = append(role.Resources, roleResources[i])
				}
			}
		}
		ch <- 1
	}
	if relateUser {
		loadCount++
		go loadUserFunc(roles, ch)
	}
	if relateResource {
		loadCount++
		go loadResourceFunc(roles, ch)
	}
	for finishCount := 0; finishCount < loadCount; {
		select {
		case <-ch:
		case <-time.NewTimer(time.Duration(10) * time.Second).C:
		}
		finishCount++
	}

	return roles, nil
}

func SearchRoleByName(clientID int64, relateUser, relateResource bool, roleName string) ([]*Role, error) {
	var res []*Role
	o := orm.NewOrm()
	if _, err := o.QueryTable("role").Filter("client_id", clientID).Filter("name__icontains", roleName).All(&res); err != nil {
		if err == orm.ErrNoRows {
			return []*Role{}, nil
		}
		logs.Error(err)
		return nil, err
	}

	if _, err := RelateUserResourceForRole(res, relateUser, relateResource); err != nil {
		return nil, err
	}

	return res, nil
}

func GetRoleByClient(clientId int64, relateUser, relateResource, isTree bool) ([]*Role, error) {
	var res []*Role
	o := orm.NewOrm()
	if _, err := o.QueryTable("role").Filter("client_id", clientId).All(&res); err != nil {
		if err == orm.ErrNoRows {
			return []*Role{}, nil
		}
		logs.Error(err)
		return nil, err
	}

	if _, err := RelateUserResourceForRole(res, relateUser, relateResource); err != nil {
		return nil, err
	}
	if isTree {
		res = BuildRoleTree(res)
	}
	return res, nil
}

func GetUserRoleByClient(clientId int64, userId string, roleType int, isAll, relateUser, relateResource, isTree, isRoute bool) ([]*Role, error) {
	var allRoles []*Role
	var userRoles []*RoleJson
	var res []*Role
	loadCount := 0
	ch := make(chan int, 2)
	o := orm.NewOrm()
	sql := fmt.Sprintf("select role.*, case role_user.role_type when %d then '%s' when %d then '%s' else '%s' end as role_type from role inner join role_user on role.id = role_user.role_id where role.client_id = ? and role_user.user_id = ? and  role_user.role_type >= ? ", RoleSuper, RoleSuperStr, RoleAdmin, RoleAdminStr, RoleNormalStr)
	// get user direct role
	directRoleFunc := func(ch chan<- int) {
		if _, err := o.Raw(sql, clientId, userId, roleType).
			QueryRows(&userRoles); err != nil {
			logs.Error(err)
		}
		ch <- 1
	}
	// get all role in client
	allRoleFunc := func(ch chan<- int) {
		if _, err := o.QueryTable("role").Filter("client_id", clientId).All(&allRoles); err != nil && err != orm.ErrNoRows {
			logs.Error(err)
		}
		ch <- 1
	}
	loadCount++
	go directRoleFunc(ch)

	if isAll {
		loadCount++
		go allRoleFunc(ch)
	}

	for finishCount := 0; finishCount < loadCount; {
		select {
		case <-ch:
		case <-time.NewTimer(time.Duration(10) * time.Second).C:
		}
		finishCount++
	}
	if len(userRoles) == 0 {
		return []*Role{}, nil
	}
	restRoleMap := make(map[int64]*Role)
	if isAll {
		if len(allRoles) == 0 {
			return []*Role{}, nil
		}
		userRoleIdMap := map[int64]*RoleJson{}
		for _, v := range userRoles {
			userRoleIdMap[v.Id] = v
		}
		roleTree := BuildRoleTree(allRoles)
		queue := roleTree
		for {
			if len(queue) == 0 {
				break
			}
			role := queue[0]
			queue = queue[1:]
			if _, ok := userRoleIdMap[role.Id]; ok {
				res = append(res, unbuildRoleTreeWithRoleType([]*Role{role}, userRoleIdMap, RoleNormalStr)...)
			} else {
				restRoleMap[role.Id] = role
				if len(role.Children) > 0 {
					queue = append(queue, role.Children...)
				}
			}
		}
	} else {
		roleJson, _ := json.Marshal(userRoles)
		json.Unmarshal(roleJson, &res)
	}

	if _, err := RelateUserResourceForRole(res, relateUser, relateResource); err != nil {
		return []*Role{}, err
	}

	if isAll && isRoute {
		var queue []int64
		routeRole := make(map[int64]*Role)
		resTree := BuildRoleTree(res)
		for i := range resTree {
			queue = append(queue, resTree[i].ParentId)
		}
		for {
			if len(queue) == 0 {
				break
			}
			parentId := queue[0]
			queue = queue[1:]
			if parent, ok := restRoleMap[parentId]; ok {
				routeRole[parent.Id] = parent
				queue = append(queue, parent.ParentId)
			}
		}
		res = unbuildRoleTree(resTree)
		for _, v := range routeRole {
			res = append(res, v)
		}

	}

	if isAll && isTree {
		res = BuildRoleTree(res)
	}
	return res, nil
}

func DeleteNonRootRole(role Role) (map[string]int64, error) {
	/*
		1、加锁
		2、判断角色是否存在
		3、判断是否为非根角色
		4、删除角色（防止有其他请求向这个角色添加user户或者关联resource）
		5、将该角色的子角色的父角色设置为该角色的父角色（ parent <- role <- children  修改后=> parent <- children）
		6、解除角色与权限关联关系
		7、解除角色与用户的关联关系
		8、解锁
	*/
	res := make(map[string]int64)
	lock := lock.NewAppLock(role.ClientId)
	lock.Lock()
	defer lock.UnLock()
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		logs.Error(err)
		return nil, err
	}
	if err := o.QueryTable("role").Filter("id", role.Id).Filter("client_id", role.ClientId).One(&role); err != nil {
		o.Commit()
		if err == orm.ErrNoRows {
			return nil, errors.ErrRoleNotFound
		} else {
			logs.Error(err)
			return nil, err
		}
	} else if role.ParentId == -1 {
		o.Commit()
		return nil, errors.ErrRootRoleCannotBeDeleted
	}

	if deletedRoleNum, err := o.Delete(&Role{Id: role.Id}); err != nil {
		logs.Error(err)
		o.Rollback()
		return nil, err
	} else {
		res["del_role_num"] = deletedRoleNum
	}
	if _, err := o.QueryTable("role").Filter("parent_id", role.Id).Update(orm.Params{"parent_id": role.ParentId}); err != nil && err != orm.ErrNoRows {
		logs.Error(err)
		o.Rollback()
		return nil, err
	}
	if deletedRelResourceNum, err := o.QueryTable("role_resource").Filter("role_id", role.Id).Delete(); err != nil && err != orm.ErrNoRows {
		logs.Error(err)
		o.Rollback()
		return nil, err
	} else {
		res["del_role_resource_num"] = deletedRelResourceNum
	}
	if deletedRelUserNum, err := o.QueryTable("role_user").Filter("role_id", role.Id).Delete(); err != nil && err != orm.ErrNoRows {
		logs.Error(err)
		o.Rollback()
		return nil, err
	} else {
		res["del_role_user_num"] = deletedRelUserNum
	}
	o.Commit()
	return res, nil
}

func UpdateRole(role Role) (*Role, error) {
	o := orm.NewOrm()
	var old Role
	if err := o.QueryTable("role").
		Filter("id", role.Id).Filter("client_id", role.ClientId).One(&old); err != nil {
		if err == orm.ErrNoRows {
			return nil, errors.ErrRoleNotFound
		} else {
			return nil, err
		}
	} else {
		cols := []string{
			"name",
			"description",
			"updated",
			"updatedBy",
		}
		if _, err := o.Update(&role, cols...); err != nil {
			return nil, err
		} else {
			role.Created = old.Created
			role.CreatedBy = old.CreatedBy
			return &role, nil
		}
	}
}

func InsertRole(childrenIds []int64, role Role) (int64, error) {
	lock := lock.NewAppLock(role.ClientId)
	lock.Lock()
	defer lock.UnLock()
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		logs.Error(err)
		return 0, err
	}
	defer o.Commit()
	var parentRole Role
	if err := o.QueryTable("role").Filter("client_id", role.ClientId).Filter("id", role.ParentId).One(&parentRole); err != nil {
		return 0, err
	}
	if num, err := o.QueryTable("role").Filter("client_id", role.ClientId).Filter("parent_id", role.ParentId).Filter("id__in", childrenIds).Count(); err != nil {
		return 0, err
	} else if num != int64(len(childrenIds)) {
		return 0, errors.ErrRoleNotFound
	}
	if id, err := o.Insert(&role); err != nil {
		o.Rollback()
		logs.Error(err)
		return 0, err
	} else {
		var resources []*Resource
		var roleIds string
		for _, v := range childrenIds {
			roleIds += strconv.FormatInt(v, 10) + ","
		}
		roleIds = roleIds[0 : len(roleIds)-1]
		if _, err := o.Raw(fmt.Sprintf("select distinct(resource.id), resource.* from resource inner join role_resource on resource.id = role_resource.resource_id where role_resource.role_id in (%s)", roleIds)).
			QueryRows(&resources); err != nil {
			o.Rollback()
			return 0, nil
		}
		if len(resources) > 0 {
			sql := "insert into role_resource (role_id, resource_id) values "
			for _, v := range resources {
				sql += fmt.Sprintf("(%d,%d),", id, v.Id)
			}
			if _, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
				o.Rollback()
				return 0, err
			}
		}
		params := orm.Params{}
		params["parent_id"] = id
		if _, err := o.QueryTable("role").Filter("id__in", childrenIds).Update(params); err != nil {
			o.Rollback()
			logs.Error(err)
			return 0, err
		} else {
			return id, nil
		}
	}
}
