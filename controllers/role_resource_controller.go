package controllers

import (
	"encoding/json"
	"strconv"

	"code.xxxxx.cn/platform/auth/models/db"
)

type RoleResourceController struct {
	BaseController
}

/**
查询角色关联的资源
*/
func (c *RoleResourceController) GetRoleResource() {
	if !c.checkScope("role-resource:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	roleId, _ := strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64)
	if roleId > 0 {
		if pass, _ := c.checkRoleType(clientId, roleId, db.RoleSuper, true); !pass {
			return
		}
	} else {
		if !c.checkRootRoleSuper(clientId) {
			return
		}
	}
	if roleResources, err := db.GetRoleResources(clientId, roleId); err != nil {
		c.Failed(err)
	} else {
		c.Ok(roleResources)

	}
}

/**
增加角色关联的资源（批量）
*/
func (c *RoleResourceController) AddRoleResource() {
	if !c.checkScope("role-resource:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	var err error
	var roleId int64
	var resourceIds []int64
	if roleId, err = strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64); err != nil {
		c.Failed(err)
		return
	}
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &resourceIds); err != nil {
		c.Failed(err)
		return
	}
	pass, isRootRoleSuper := c.checkRoleType(clientId, roleId, db.RoleSuper, true)
	if !pass {
		return
	}
	if num, err := db.AddRoleResource(resourceIds, clientId, roleId, isRootRoleSuper); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}
}

/**
更新角色关联的权限（批量）
*/
func (c *RoleResourceController) UpdateRoleResource() {
	if !c.checkScope("role-resource:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	var err error
	var roleId int64
	var resourceIds []int64
	if roleId, err = strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64); err != nil {
		c.Failed(err)
		return
	}
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &resourceIds); err != nil {
		c.Failed(err)
		return
	}
	pass, isRootRoleSuper := c.checkRoleType(clientId, roleId, db.RoleSuper, true)
	if !pass {
		return
	}
	if num, err := db.UpdateRoleResource(resourceIds, clientId, roleId, isRootRoleSuper); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}
}

/**
删除角色关联的资源
*/
func (c *RoleResourceController) DeleteRoleResource() {
	if !c.checkScope("role-resource:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	var err error
	var roleId int64
	var resourceIds []int64
	if roleId, err = strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64); err != nil {
		c.Failed(err)
		return
	}
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &resourceIds); err != nil {
		c.Failed(err)
		return
	}
	if pass, _ := c.checkRoleType(clientId, roleId, db.RoleSuper, true); !pass {
		return
	}
	if num, err := db.DeleteRoleResource(resourceIds, clientId, roleId); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}

}
