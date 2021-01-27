package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego/orm"

	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego/logs"
)

type SyncController struct {
	BaseController
}

type BodyData struct {
	ClientRows []*ClientRow `json:"clients"`
	Tokens     []*db.Token  `json:"tokens"`
	Users      []*db.User   `json:"users"`
	Groups     []*db.Group  `json:"groups"`
}

type ClientRow struct {
	Client       *db.Client     `json:"client"`
	RoleTree     []*db.Role     `json:"role_tree"`
	ResourceRows []*ResourceRow `json:"resources"`
}

type RoleRow struct {
	Role      *db.Role       `json:"role"`
	Resources []*db.Resource `json:"resources"`
	RoleUsers []*db.RoleUser `json:"role_users"`
}

type ResourceRow struct {
	Resource *db.Resource `json:"resource"`
	Roles    []*db.Role   `json:"roles"`
}

// Sync upload users and groups
func (c *SyncController) Sync() {
	bodyData := BodyData{}
	logs.Info("body:%s", string(c.Ctx.Input.RequestBody))
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &bodyData); err != nil {
		c.Failed(err)
		return
	}

	// Sync users
	for _, user := range bodyData.Users {
		if user.Id == "" {
			continue
		}
		// fill default field if empty
		if len(user.Status) == 0 {
			user.Status = db.UserStatusActive
		}
		if len(user.Type) == 0 {
			if user.Dn == "special user" {
				user.Type = db.UserTypeSpecial
			} else {
				user.Type = db.UserTypeSys
			}
		}
		//if len(user.Email) == 0 {
		//	var mailPostfix = ""
		//	if service.Ldap.Enabled() {
		//		mailPostfix = beego.AppConfig.String("ldap_mailPostfix")
		//	} else {
		//		mailPostfix = "@" + beego.AppConfig.String("org_domain")
		//	}
		//	user.Email = user.Id + mailPostfix
		//}
		if _, err := db.SaveUser(*user); err != nil {
			logs.Warning("save user %s failed: %v", user.Id, err)
			continue
		}
		logs.Info("save user %s succeed", user.Id)
	}
	logs.Info("Sync %d users succeed", len(bodyData.Users))

	// Sync groups
	for _, group := range bodyData.Groups {
		if _, err := db.SaveGroup(group); err != nil {
			logs.Info("save group %s failed: ", group.Name, err)
			c.Failed(err)
			continue
		}
		logs.Info("save group %s succeed", group.Name)

		for _, member := range group.Members {
			gu := &db.GroupUser{
				Group: group.Name,
				User:  member,
			}
			if err := db.SaveGroupUser(gu); err != nil {
				logs.Info("save group user %s:%s failed:", gu.Group, gu.User, err)
				c.Failed(err)
				continue
			}
			logs.Info("save group user %s:%s succeed", gu.Group, gu.User)
		}
	}
	logs.Info("Sync %d groups succeed", len(bodyData.Groups))

	o := orm.NewOrm()
	// Sync clients
	for _, clientRow := range bodyData.ClientRows {
		oldClient := &db.Client{}
		if exist := o.QueryTable("client").Filter("fullname", clientRow.Client.Fullname).Exist(); !exist {
			clientId, err := o.Insert(clientRow.Client)
			if err != nil {
				logs.Error("Insert Client %d - %s failed:%v\n", clientRow.Client.Id, clientRow.Client.Fullname, err)
				continue
			}
			logs.Info("Insert Client %d - %s succeed", clientId, clientRow.Client.Fullname)
			oldClient = clientRow.Client
		} else {
			err := o.QueryTable("client").Filter("id", clientRow.Client.Id).One(oldClient)
			if err != nil {
				logs.Error("Get client %s failed: %v", clientRow.Client.Id, err)
				continue
			}
			logs.Warning("Client %d already exist", clientRow.Client.Id)
		}

		syncClientRow(clientRow, oldClient.Id, o)
	}
	logs.Info("Sync %d clients succeed", len(bodyData.ClientRows))

	logs.Info("Sync All succeed")
	c.Ok("ok")
}

func SyncRoleTree(clientId int64, root *db.Role) {
	if root == nil {
		return
	}
	// save role
	role, err := SaveRole(clientId, root)
	if err != nil {
		return
	}
	for _, child := range root.Children {
		child.ParentId = role.Id
		child.ClientId = clientId
		SyncRoleTree(clientId, child)
	}
}

func SaveRole(clientId int64, role *db.Role) (*db.Role, error) {
	o := orm.NewOrm()
	var oldRole = &db.Role{}
	if exist := o.QueryTable("role").Filter("name", role.Name).Filter("client_id", clientId).Exist(); !exist {
		oldRole.Name = role.Name
		oldRole.Description = role.Description
		oldRole.ParentId = role.ParentId
		oldRole.ClientId = role.ClientId
		oldRole.CreatedBy = role.CreatedBy
		oldRole.Created = role.Created
		oldRole.UpdatedBy = role.UpdatedBy
		oldRole.Updated = role.Updated
		roleId, err := o.Insert(oldRole)
		if err != nil {
			logs.Error("Insert role %s failed: %v", oldRole.Name, err)
			return nil, err
		}
		role.Id = roleId
		logs.Info("Insert role %s id=%d, pid=%d succeed", oldRole.Name, oldRole.Id, oldRole.ParentId)
	} else {
		err := o.QueryTable("role").Filter("name", role.Name).Filter("client_id", clientId).One(oldRole)
		if err != nil {
			logs.Error("Get role %s failed: %v", role.Name, err)
			return nil, err
		}
		logs.Warning("Role %s already exist", role.Name)
	}
	for _, roleResource := range role.Resources {
		resource, err := saveResource(roleResource.Resource, clientId)
		if err != nil {
			continue
		}
		saveRoleResource(role, resource)
	}

	for _, roleUser := range role.Users {
		saveRoleUser(role, roleUser)
	}
	return oldRole, nil
}

func saveResource(resource *db.Resource, clientID int64) (*db.Resource, error) {
	o := orm.NewOrm()
	var oldResource = &db.Resource{}
	if exist := o.QueryTable("resource").Filter("name", resource.Name).Filter("data", resource.Data).Filter("client_id", clientID).Exist(); !exist {
		resource.Id = 0
		resource.ClientId = clientID
		resId, err := o.Insert(resource)
		if err != nil {
			logs.Error("Insert resource %s failed: %v", resource.Name, err)
			return nil, err
		}
		logs.Info("Insert resource %s succeed", resource.Name)
		resource.Id = resId
		oldResource = resource
	} else {
		err := o.QueryTable("resource").Filter("name", resource.Name).Filter("data", resource.Data).Filter("client_id", clientID).One(oldResource)
		if err != nil {
			logs.Error("Get resource %s failed: %v", resource.Name, err)
			return nil, err
		}
		logs.Warning("Resource %s already exist", resource.Name)
	}
	return oldResource, nil
}

func saveRoleResource(oldRole *db.Role, oldResource *db.Resource) {
	o := orm.NewOrm()
	// sync role_resource
	if exist := o.QueryTable("role_resource").Filter("role_id", oldRole.Id).Filter("resource_id", oldResource.Id).Exist(); !exist {
		roleResource := &db.RoleResource{RoleId: oldRole.Id, Resource: oldResource}
		_, err := o.Insert(roleResource)
		if err != nil {
			logs.Error("Insert role_resource<%d,%d> failed: %v", oldRole.Id, oldResource.Id, err)
			return
		}
		logs.Info("Insert role_resource <%d,%d> succeed", oldRole.Id, oldResource.Id)
	} else {
		logs.Warning("role_resource <%d,%d> already exist", oldRole.Id, oldResource.Id)
	}
}

func saveRoleUser(oldRole *db.Role, roleUser *db.RoleUser) {
	o := orm.NewOrm()
	if exist := o.QueryTable("user").Filter("id", roleUser.User.Id).Exist(); !exist {
		logs.Warning("User %s not found", roleUser.User.Id)
	}
	if exist := o.QueryTable("role_user").Filter("role_id", oldRole.Id).Filter("user_id", roleUser.User.Id).Exist(); !exist {
		roleUser.RoleId = oldRole.Id
		roleUser.Id = 0
		_, err := o.Insert(roleUser)
		if err != nil {
			logs.Error("Insert role_user <%d, %s> failed: %v", roleUser.RoleId, roleUser.User.Id, err)
			return
		}
		logs.Info("Insert role_user <%d, %s, %d> succeed", roleUser.RoleId, roleUser.User.Id, roleUser.RoleType)
	} else {
		logs.Warning("role_user <%d,%s> already exist", roleUser.RoleId, roleUser.User.Id, roleUser.RoleType)
	}
}

func (c *SyncController) SyncInClients() {
	src, err := c.GetInt64("src_id")
	if err != nil {
		c.Failed(err, 233)
		return
	}

	dst, err := c.GetInt64("dst_id")
	if err != nil {
		c.Failed(err, 234)
		return
	}

	if err := syncClients(src, dst); err != nil {
		c.Failed(err, 235)
		return
	}

	c.Ok("")
}

func syncClients(src, dst int64) error {
	cr, err := getSyncClientResources(src)
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	return syncClientRow(cr, dst, o)
}

func syncClientRow(clientRow *ClientRow, newClientID int64, o orm.Ormer) error {
	// sync resource not in roleTree
	for index, resourceRow := range clientRow.ResourceRows {
		resource := resourceRow.Resource
		resource.Id = 0
		resource.ClientId = newClientID
		clientRow.ResourceRows[index].Resource.ClientId = newClientID

		if exist := o.QueryTable("resource").Filter("name", resource.Name).Filter("data", resource.Data).Filter("client_id", resource.ClientId).Exist(); !exist {
			resId, err := o.Insert(resource)
			if err != nil {
				logs.Error("Insert resource %s failed: %v", resource.Name, err)
				return err
			}
			resource.Id = resId
			logs.Info("Insert resource %s succeed", resource.Name)
		}
	}

	// sync roles
	for _, rootRole := range clientRow.RoleTree {
		SyncRoleTree(newClientID, rootRole)
	}

	return nil
}

func getSyncClientResources(clientID int64) (*ClientRow, error) {
	o := orm.NewOrm()
	client := db.Client{}
	err := o.QueryTable("client").Filter("id", clientID).One(&client)
	if err != nil {
		logs.Error("Get client failed: %v", err)
		return nil, err
	}

	// get roles
	rolesTree, err := db.GetRoleByClient(client.Id, true, true, true)
	if err != nil {
		logs.Error("Get roles tree failed: %v", err)
		return nil, err
	}

	// get resources
	var resources []*db.Resource
	count, err := o.QueryTable("resource").Filter("client_id", client.Id).All(&resources)
	if err != nil {
		logs.Error("Get resources failed: %v", err)
		return nil, err
	}
	logs.Info("Got %d resources", count)

	var resourceRows []*ResourceRow
	for _, resource := range resources {
		var roleResources []*db.RoleResource
		count, err = o.QueryTable("role_resource").Filter("role_id", resource.Id).All(&roleResources)
		var roles []*db.Role
		for _, roleRes := range roleResources {
			var role = &db.Role{}
			o.QueryTable("role").Filter("id", roleRes.RoleId).One(role)
			roles = append(roles, role)
		}
		var resourceRow = &ResourceRow{
			Resource: resource,
			Roles:    roles,
		}
		resourceRows = append(resourceRows, resourceRow)
	}

	clientRow := &ClientRow{
		Client:       &client,
		ResourceRows: resourceRows,
		RoleTree:     rolesTree,
	}

	return clientRow, nil
}
