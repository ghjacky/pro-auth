package db

import (
	"fmt"
	"strconv"
	"time"

	"code.xxxxx.cn/platform/auth/errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type Resource struct {
	Id          int64           `orm:"pk;auto" json:"id"`
	Name        string          `orm:"size(128)"json:"name"`
	Description string          `orm:"size(128)" json:"description"`
	ClientId    int64           `orm:"index" json:"client_id"`
	Data        string          `orm:"size(1024)" json:"data"`
	CreatedBy   string          `orm:"size(255)" json:"created_by"`
	UpdatedBy   string          `orm:"size(255)" json:"updated_by"`
	Created     time.Time       `orm:"auto_now_add;type(datetime)"json:"created"`
	Updated     time.Time       `orm:"auto_now;type(datetime)" json:"updated"`
	Roles       []*ResourceRole `orm:"-" json:"roles,omitempty"`
}

type ResourceRole struct {
	ResourceId  int64  `json:"resource_id"`
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentId    int64  `json:"parent_id"`
}

func GetResources(params orm.Params, clientId int64, relateRole bool) ([]*Resource, error) {
	var resources []*Resource
	o := orm.NewOrm()
	qs := o.QueryTable("resource")
	for k, v := range params {
		qs = qs.Filter(k, v)
	}
	qs = qs.Filter("client_id", clientId)
	if _, err := qs.All(&resources); err != nil {
		if err == orm.ErrNoRows {
			return []*Resource{}, nil
		} else {
			logs.Error(err)
			return nil, err
		}
	} else {
		if relateRole {
			var resourceRoles []*ResourceRole
			sql := fmt.Sprintf("select role_resource.resource_id, role.* from role_resource join role on role.id = role_resource.role_id where role.client_id = %d", clientId)
			if _, err := o.Raw(sql).QueryRows(&resourceRoles); err != nil {
				return nil, err
			}
			resourceMap := make(map[int64]*Resource)
			for i := range resources {
				resourceMap[resources[i].Id] = resources[i]
			}
			for _, v := range resourceRoles {
				if r, ok := resourceMap[v.ResourceId]; ok {
					r.Roles = append(r.Roles, v)
				} else {
					logs.Error("resource_role can't find resource %+v", v)
				}
			}
		}
		return resources, nil
	}

}

func GetUserResources(clientId int64, userId string) ([]*Resource, error) {
	resources := []*Resource{}
	o := orm.NewOrm()
	roleIds := ""
	if roles, err := GetUserRoleByClient(clientId, userId, RoleNormal, true, false, false, false, false); err != nil {
		return nil, err
	} else if len(roles) == 0 {
		return resources, nil
	} else {
		for _, v := range roles {
			roleIds += strconv.FormatInt(v.Id, 10) + ","
		}
		roleIds = roleIds[0 : len(roleIds)-1]
	}
	if _, err := o.Raw(fmt.Sprintf("select distinct(resource.id), resource.* from resource inner join role_resource on resource.id = role_resource.resource_id where role_resource.role_id in (%s)", roleIds)).
		QueryRows(&resources); err != nil {
		if err == orm.ErrNoRows {
			return resources, nil
		} else {
			return nil, err
		}
	}
	return resources, nil
}

func AddResource(resources []Resource, clientId int64) ([]int64, error) {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return nil, err
	}
	var rootRole Role
	if err := o.QueryTable("role").Filter("client_id", clientId).Filter("parent_id", -1).One(&rootRole); err != nil {
		o.Commit()
		if err == orm.ErrNoRows {
			return nil, errors.ErrResourceNotFound
		} else {
			return nil, err
		}
	}

	ids := make([]int64, len(resources))
	inserterR, _ := o.QueryTable("resource").PrepareInsert()

	for i, v := range resources {
		v.ClientId = clientId
		if id, err := inserterR.Insert(&v); err != nil {
			logs.Error(err)
			o.Rollback()
			return nil, err
		} else {
			ids[i] = id
		}
	}
	sql := "insert into role_resource (role_id, resource_id) values "
	for _, v := range ids {
		sql += fmt.Sprintf("(%d,%d),", rootRole.Id, v)
	}
	if _, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
		o.Rollback()
		return nil, err
	} else {
		o.Commit()
		return ids, nil
	}

}

func UpdateResource(resource Resource, params orm.Params) (*Resource, error) {
	o := orm.NewOrm()
	var old Resource
	qs := o.QueryTable("resource")
	for k, v := range params {
		qs = qs.Filter(k, v)
	}
	if err := qs.Filter("id", resource.Id).One(&old); err != nil {
		if err == orm.ErrNoRows {
			return nil, errors.ErrResourceNotFound
		} else {
			return nil, err
		}
	} else {
		cols := []string{
			"Name",
			"Description",
			"Data",
			"UpdatedBy",
			"Updated",
		}
		if _, err := o.Update(&resource, cols...); err != nil {
			logs.Error(err)
			return nil, err
		} else {
			resource.ClientId = old.ClientId
			resource.CreatedBy = old.CreatedBy
			resource.Created = old.Created
			return &resource, nil
		}
	}

}

func DeleteResources(ids []int64, params orm.Params) (map[string]int64, error) {
	res := make(map[string]int64)
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return res, err
	}
	qs := o.QueryTable("resource")
	for k, v := range params {
		qs = qs.Filter(k, v)
	}
	if num, err := qs.Filter("id__in", ids).Delete(); err != nil {
		o.Rollback()
		return res, err
	} else if num != int64(len(ids)) {
		o.Rollback()
		return res, errors.ErrResourceNotFound
	} else if num2, err := o.QueryTable("role_resource").
		Filter("resource_id__in", ids).
		Delete(); err != nil {
		logs.Error(err)
		o.Rollback()
		return res, err
	} else {
		o.Commit()
		res["del_resource_num"] = num
		res["del_role_resource_num"] = num2
		return res, nil
	}

}
