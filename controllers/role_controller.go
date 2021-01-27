package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"code.xxxxx.cn/platform/auth/models/db"
)

type RoleController struct {
	BaseController
}

/**
查询client下全部角色 （rootRoleAdmin）,可选是否关联权限，可选是否关联用户
*/
func (c *RoleController) GetRoleByClient() {
	if !c.checkScope("role:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkIsRoleAdminOrSuper(clientId) {
		return
	}
	relateUser, _ := c.GetBool("relate_user", false)
	relateResource, _ := c.GetBool("relate_resource", false)
	isTree, _ := c.GetBool("is_tree", false)
	if roles, err := db.GetRoleByClient(clientId, relateUser, relateResource, isTree); err != nil {
		c.Failed(err)
	} else {
		c.Ok(roles)
	}
}

// SearchRoleByNameInClient 搜索client下全部角色, 可选是否关联权限，可选是否关联用户
func (c *RoleController) SearchRoleByNameInClient() {
	if !c.checkScope("role:read") {
		return
	}
	clientID := c.checkClientId()
	if clientID == 0 {
		return
	}
	if !c.checkIsRoleAdminOrSuper(clientID) {
		c.Failed(fmt.Errorf("No Authority"))
		return
	}
	relateUser, _ := c.GetBool("relate_user", false)
	relateResource, _ := c.GetBool("relate_resource", false)
	roleName := c.GetString("role_name")
	if roleName == "" {
		c.Failed(fmt.Errorf("role name should not be empty %v", roleName))
		return
	}

	roles, err := db.SearchRoleByName(clientID, relateUser, relateResource, roleName)
	if err != nil {
		c.Failed(err)
		return
	}
	c.Ok(roles)
}

// GetRoleTreeByID 查询指定角色, 可选是否关联权限，可选是否关联用户
func (c *RoleController) GetRoleTreeByID() {
	if !c.checkScope("role:read") {
		return
	}
	clientID := c.checkClientId()
	if clientID == 0 {
		return
	}
	if !c.checkIsRoleAdminOrSuper(clientID) {
		return
	}
	roleID, err := c.GetInt64("role_id", 0)
	if err != nil {
		c.Failed(err)
		return
	}

	relateUser, _ := c.GetBool("relate_user", false)
	relateChildren, _ := c.GetBool("relate_children", false)
	relateResource, _ := c.GetBool("relate_resource", false)
	isTree, _ := c.GetBool("is_tree", false)

	rootRole, err := db.GetRole(roleID)
	if err != nil {
		c.Failed(err)
		return
	}
	roles := []*db.Role{rootRole}

	if relateChildren {
		parentRoleIDs := []int64{roleID}
		for len(parentRoleIDs) > 0 {
			childrenRoles, err := db.GetRolesByParentIDs(parentRoleIDs)
			if err != nil {
				c.Failed(err)
				return
			}
			roles = append(roles, childrenRoles...)
			parentRoleIDs = make([]int64, len(childrenRoles))
			for index, role := range childrenRoles {
				parentRoleIDs[index] = role.Id
			}
		}
	}

	if _, err = db.RelateUserResourceForRole(roles, relateUser, relateResource); err != nil {
		c.Failed(err)
		return
	}

	if isTree {
		roles = db.BuildRoleTree(roles)
	}

	c.Ok(roles)
}

/**
查询某用户在某client下的角色 (全部或直接，normal或admin)，可选是否关联权限，可选是否关联用户
*/
func (c *RoleController) GetUserRoleByClient() {
	if !c.checkScope("role:read") {
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
	roleType := db.RoleNormal
	if c.GetString("role_type") == "admin" {
		roleType = db.RoleAdmin
	} else if c.GetString("role_type") == "super" {
		roleType = db.RoleSuper
	}
	isAll, _ := c.GetBool("is_all", false)
	relateUser, _ := c.GetBool("relate_user", false)
	relateResource, _ := c.GetBool("relate_resource", false)
	isTree, _ := c.GetBool("is_tree", false)
	isRoute, _ := c.GetBool("is_route", false)
	if roles, err := db.GetUserRoleByClient(clientId, userId, roleType, isAll, relateUser, relateResource, isTree, isRoute); err != nil {
		c.Failed(err)
	} else {
		c.Ok(roles)
	}
}

// GetUserRoleByUserIDs filter user role by user ids
func (c *RoleController) GetUserRoleByUserIDs() {
	if !c.checkScope("role:read") {
		return
	}
	clientID := c.checkClientId()
	if clientID == 0 {
		return
	}
	userIDstrs := strings.Split(c.Ctx.Request.FormValue("user_ids"), ",")
	userRoles, err := db.GetRoleUserByUserIDs(userIDstrs)
	if err != nil {
		c.Failed(err)
		return
	}

	c.Ok(userRoles)
}

// GetRoleByIDs filter role by ids
func (c *RoleController) GetRoleByIDs() {
	if !c.checkScope("role:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	roleIDstrs := strings.Split(c.Ctx.Request.FormValue("role_ids"), ",")
	roleIDs := make([]int64, len(roleIDstrs))
	for index, IDstr := range roleIDstrs {
		if IDstr == "" {
			continue
		}
		id, err := strconv.Atoi(IDstr)
		if err != nil {
			c.Failed(fmt.Errorf("Invalid user ids"))
			return
		}
		roleIDs[index] = int64(id)
	}

	roles, err := db.GetRoleByIDs(roleIDs)
	if err != nil {
		c.Failed(err)
		return
	}

	c.Ok(roles)
}

/**
增加子角色（rootRoleAdmin）
*/

func (c *RoleController) AddRole() {
	if !c.checkScope("role:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	authInfo := c.getAuthInfo()
	role := db.Role{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &role)
	if err != nil {
		c.Failed(err)
	}
	if pass, _ := c.checkRoleType(clientId, role.ParentId, db.RoleSuper, false); !pass {
		return
	}
	role.CreatedBy = authInfo.UserId
	role.UpdatedBy = authInfo.UserId
	role.ClientId = clientId
	if id, err := db.AddSubRole(role); err != nil {
		c.Failed(err)
	} else {
		c.Ok(id)
	}
}

/**
删除角色（rootRoleAdmin）
*/
func (c *RoleController) DeleteRole() {
	if !c.checkScope("role:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	roleId, err := strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64)
	if err != nil {
		c.Failed(err)
	}
	if pass, _ := c.checkRoleType(clientId, roleId, db.RoleSuper, true); !pass {
		return
	}
	role := db.Role{ClientId: clientId, Id: roleId}
	if res, err := db.DeleteNonRootRole(role); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

/**
更新角色
*/
func (c *RoleController) UpdateRole() {

	if !c.checkScope("role:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	role := db.Role{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &role)
	if err != nil {
		c.Failed(err)
		return
	}
	if pass, _ := c.checkRoleType(clientId, role.Id, db.RoleSuper, true); !pass {
		return
	}
	authInfo := c.getAuthInfo()
	role.UpdatedBy = authInfo.UserId
	role.ClientId = clientId
	if r, err := db.UpdateRole(role); err != nil {
		c.Failed(err)
	} else {
		c.Ok(r)
	}
}

func (c *RoleController) InsertRole() {
	if !c.checkScope("role:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	var err error
	var childrenIds []int64
	param := strings.Split(c.Ctx.Input.Param(":children_ids"), ",")
	for _, v := range param {
		if vFormat, err := strconv.ParseInt(v, 10, 64); err != nil {
			c.Failed(err)
			return
		} else {
			childrenIds = append(childrenIds, vFormat)
		}
	}
	authInfo := c.getAuthInfo()
	role := db.Role{}
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &role)
	if err != nil {
		c.Failed(err)
		return
	}
	if pass, _ := c.checkRoleType(clientId, role.ParentId, db.RoleSuper, false); !pass {
		return
	}
	role.CreatedBy = authInfo.UserId
	role.UpdatedBy = authInfo.UserId
	role.ClientId = clientId
	if id, err := db.InsertRole(childrenIds, role); err != nil {
		c.Failed(err)
	} else {
		c.Ok(id)
	}
}
