package controllers

import (
	"encoding/json"
	"strconv"
	"strings"

	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego/orm"
)

type ResourceController struct {
	BaseController
}

// GetResourcesByClient 查询client下全部权限（rootRoleAdmin only）
func (c *ResourceController) GetResourcesByClient() {
	if !c.checkScope("resource:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	relateRole, _ := c.GetBool("relate_role", false)

	if resources, err := db.GetResources(orm.Params{}, clientId, relateRole); err != nil {
		c.Failed(err)
	} else {
		c.Ok(resources)
	}
}

func (c *ResourceController) GetResourcesByIDs() {
	if !c.checkScope("resource:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	roleIDStrings := c.GetStrings("id")
	roleIDs := []int64{}
	for _, idStr := range roleIDStrings {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.Failed(err)
			return
		}
		roleIDs = append(roleIDs, int64(id))
	}

	if resources, err := db.GetResources(orm.Params{
		"id__in": roleIDs,
	}, clientId, false); err != nil {
		c.Failed(err)
	} else {
		c.Ok(resources)
	}
}

// AddResources 增加resource
func (c *ResourceController) AddResources() {
	if !c.checkScope("resource:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	var resources []db.Resource

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &resources); err != nil {
		c.Failed(err)
	} else {
		authInfo := c.getAuthInfo()
		for i := 0; i < len(resources); i++ {
			resources[i].CreatedBy = authInfo.UserId
			resources[i].UpdatedBy = authInfo.UserId
		}
		if ids, err := db.AddResource(resources, clientId); err != nil {
			c.Failed(err)
		} else {
			c.Ok(ids)
		}
	}

}

// DeleteResources 删除resource（rootRoleAdmin only）
func (c *ResourceController) DeleteResources() {
	if !c.checkScope("resource:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	var ids []int64
	param := strings.Split(c.Ctx.Input.Param(":ids"), ",")
	for _, v := range param {
		if vFormat, err := strconv.ParseInt(v, 10, 64); err != nil {
			c.Failed(err)
			return
		} else {
			ids = append(ids, vFormat)
		}
	}
	params := orm.Params{"client_id": clientId}
	if num, err := db.DeleteResources(ids, params); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}
}

// UpdateResources 更新resource(rootRoleAdmin only)
func (c *ResourceController) UpdateResources() {
	if !c.checkScope("resource:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	var resource db.Resource
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &resource); err != nil {
		c.Failed(err)
	} else {
		authInfo := c.getAuthInfo()
		resource.UpdatedBy = authInfo.UserId
		if num, err := db.UpdateResource(resource, orm.Params{"client_id": clientId}); err != nil {
			c.Failed(err)
		} else {
			c.Ok(num)
		}
	}
}

// GetResourcesByUserAndClient 查询user在某client下的resources
func (c *ResourceController) GetResourcesByUserAndClient() {
	if !c.checkScope("resource:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	userId, ok := c.checkUserId(clientId)
	if !ok {
		return
	}
	if resources, err := db.GetUserResources(clientId, userId); err != nil {
		c.Failed(err)
	} else {
		c.Ok(resources)
	}
}
