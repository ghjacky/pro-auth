package controllers

import (
	"encoding/json"
	"errors"
	"regexp"

	"code.xxxxx.cn/platform/auth/consts"
	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type ClientController struct {
	BaseController
}

func (c *ClientController) CreateClient() {
	if !c.checkScope("client:write") {
		return
	}
	authInfo := c.getAuthInfo()
	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		c.Failed(ErrNoAuthorityAPI)
		return
	}
	userId := authInfo.UserId
	client := db.Client{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &client); err != nil {
		c.Failed(err)
	}
	client.UserId = userId
	if client.Fullname == "" {
		c.Failed(errors.New("fullname cannot be empty string "))
		return
	}
	if match, _ := regexp.MatchString(`^((http)|(https))://\w*`, client.RedirectUri); !match {
		c.Failed(errors.New("redirect_uri must start with http:// or https://"))
		return
	}
	if res, err := db.CreateClient(client); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

func (c *ClientController) UpdateClient() {
	if !c.checkScope("client:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	client := db.Client{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &client); err != nil {
		c.Failed(err)
	}
	client.Id = clientId
	if client.Fullname == "" {
		c.Failed(errors.New("fullname cannot be empty string "))
		return
	}
	if match, _ := regexp.MatchString(`^((http)|(https))://\w*`, client.RedirectUri); !match {
		c.Failed(errors.New("redirect_uri must start with http:// or https://"))
		return
	}
	if res, err := db.UpdateClient(client); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

func (c *ClientController) GetClientsByUser() {
	if !c.checkScope("client:read") {
		return
	}
	authInfo := c.getAuthInfo()
	if authInfo.LoginType != consts.AUTH_BY_SESSION {
		c.Failed(ErrNoAuthorityAPI)
		return
	}
	userId, ok := c.checkUserId(0)
	if !ok {
		return
	}
	roleType := db.RoleNormal
	if c.GetString("role_type") == "admin" {
		roleType = db.RoleAdmin
	} else if c.GetString("role_type") == "super" {
		roleType = db.RoleSuper
	}
	if clients, err := db.GetClientsByUser(userId, roleType); err != nil {
		c.Failed(err)
	} else {
		c.Ok(clients)
	}
}

func (c *ClientController) GetAllClients() {
	authInfo := c.getAuthInfo()
	if authInfo.LoginType != consts.AUTH_BY_SESSION {
		c.Failed(ErrNoAuthorityAPI)
		return
	}
	if clients, err := db.GetAllClients(); err != nil {
		c.Failed(err)
	} else {
		c.Ok(clients)
	}
}

func (c *ClientController) GetClient() {
	if !c.checkScope("client:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientId) {
		return
	}
	if client, err := db.GetClient(clientId); err != nil {
		c.Failed(err)
	} else {
		c.Ok(client)
	}
}

// ClientClone clone client to new client
func (c *ClientController) ClientClone() {
	if !c.checkScope("client:write") {
		return
	}

	clientID := c.checkClientId()
	if clientID == 0 {
		return
	}
	if !c.checkRootRoleSuper(clientID) {
		return
	}

	client, err := db.GetClient(clientID)
	if err != nil {
		c.Failed(err)
		logs.Warn("get client err: %v\n", err)
		return
	}

	client.Id = 0
	client.Fullname = client.Fullname + "_copy"
	newClient, err := db.CreateClient(*client)
	if err != nil {
		c.Failed(err)
		logs.Warn("create client err: %v\n", err)
		return
	}

	resources, err := db.GetResources(orm.Params{}, clientID, false)
	if err != nil {
		c.Failed(err)
		logs.Warn("get resource err: %v\n", err)
		return
	}

	oldIDs := make([]int64, len(resources))
	res := make([]db.Resource, len(resources))
	for index, r := range resources {
		oldIDs[index] = r.Id
		r.Id = int64(0)
		res[index] = *r
	}

	resIDs := []int64{}
	if len(resources) > 0 {
		resIDs, err = db.AddResource(res, newClient.Id)
		if err != nil {
			c.Failed(err)
			logs.Warn("add resource err: %v\n", err)
			return
		}
	}

	old2newResMap := make(map[int64]int64, len(resIDs))
	for index, oldID := range oldIDs {
		old2newResMap[oldID] = resIDs[index]
	}

	oldRoleTree, err := db.GetRoleByClient(clientID, true, true, true)
	if err != nil {
		c.Failed(err)
		logs.Warn("get role by client err: %v\n", err)
		return
	}

	if len(oldRoleTree) < 1 {
		c.Failed(errors.New("Root Role DNE"))
		return
	}

	rootRole := oldRoleTree[0]
	rootRole.ClientId = newClient.Id
	if err := cloneRoleUser(rootRole, newClient.RootRoleId); err != nil {
		logs.Error("Clone Root Role Failed!")
		c.Failed(err)
		return
	}

	for _, role := range rootRole.Children {
		role.ParentId = newClient.RootRoleId
		role.ClientId = newClient.Id
		if err = cloneRoleTree(role, old2newResMap); err != nil {
			c.Failed(err)
			logs.Warn("clone role tree err: %v\n", err)
			return
		}
	}

	c.Ok(newClient)
}

func cloneRoleTree(role *db.Role, old2newResMap map[int64]int64) error {
	role.Id = 0
	newID, err := db.AddSubRole(*role)
	if err != nil {
		return err
	}

	if err := cloneRoleResourceAndUser(role, old2newResMap, newID); err != nil {
		return err
	}

	for _, cr := range role.Children {
		cr.ClientId = role.ClientId
		cr.ParentId = newID
		if err = cloneRoleTree(cr, old2newResMap); err != nil {
			logs.Warn("clone sub %v role tree err: %v\n", role.Name, err)
			return err
		}
	}

	return nil
}

func cloneRoleResourceAndUser(role *db.Role, old2newResMap map[int64]int64, newRoleID int64) error {
	if err := cloneRoleResource(role, old2newResMap, newRoleID); err != nil {
		return err
	}

	if err := cloneRoleUser(role, newRoleID); err != nil {
		return err
	}

	return nil
}

func cloneRoleResource(role *db.Role, old2newResMap map[int64]int64, newRoleID int64) error {
	if len(role.Resources) > 0 {
		newResIDs := make([]int64, len(role.Resources))
		for index, rr := range role.Resources {
			newResIDs[index] = old2newResMap[rr.Resource.Id]
		}

		_, err := db.AddRoleResource(newResIDs, role.ClientId, newRoleID, true)
		if err != nil {
			logs.Error("Add Role Resource Error: Res ID: %v, clientID: %v", newResIDs, role.ClientId)
			return err
		}
	}
	return nil
}

func cloneRoleUser(role *db.Role, newRoleID int64) error {
	if len(role.Users) > 0 {
		newRoleUsers := make([]db.RoleUserJSON, len(role.Users))
		for index, ru := range role.Users {
			newRoleUsers[index] = db.RoleUserJSON{
				RoleId:   newRoleID,
				UserId:   ru.User.Id,
				RoleType: db.RoleTypeMap[int(ru.RoleType)],
			}
		}

		_, err := db.AddRoleUserBatch(newRoleUsers)
		if err != nil {
			logs.Error("Add Role User Error: %v", newRoleUsers)
			return err
		}
	}
	return nil
}
