package routers

import (
	"code.xxxxx.cn/platform/auth/controllers"
	"code.xxxxx.cn/platform/auth/filter"
	"github.com/astaxie/beego"
)

func init() {

	//oauth2
	beego.Router("/oauth2/authorize", &controllers.Oauth2Controller{}, "GET,POST:Authorize")
	beego.Router("/oauth2/token", &controllers.Oauth2Controller{}, "POST:Token")
	beego.Router("/oauth2/expire", &controllers.Oauth2Controller{}, "GET:IsTokenExpired")
	beego.Router("/oauth2/token", &controllers.Oauth2Controller{}, "DELETE:DestroyToken")

	beego.InsertFilter("/api/*", beego.BeforeRouter, filter.CheckAuthFilter)
	beego.InsertFilter("/api2/*", beego.BeforeRouter, filter.CheckAuthFilter)

	//auth
	beego.Router("/api:v([2]{0,1})/checkResource", &controllers.AuthController{}, "GET:CheckResource")
	beego.Router("/api:v([2]{0,1})/user/auth", &controllers.AuthController{}, "POST:AuthUser")
	beego.Router("/api:v([2]{0,1})/user/check", &controllers.AuthController{}, "GET:CheckThirdParty")
	beego.Router("/api:v([2]{0,1})/user/secret/generate", &controllers.AuthController{}, "POST:GenerateUserSecret")

	// role
	beego.Router("/api:v([2]{0,1})/roles", &controllers.RoleController{}, "GET:GetRoleByClient")
	beego.Router("/api:v([2]{0,1})/roles/search", &controllers.RoleController{}, "GET:SearchRoleByNameInClient")
	beego.Router("/api:v([2]{0,1})/userRoles", &controllers.RoleController{}, "GET:GetUserRoleByClient")
	beego.Router("/api:v([2]{0,1})/userRoles/users", &controllers.RoleController{}, "GET:GetUserRoleByUserIDs")
	beego.Router("/api:v([2]{0,1})/roles/batch", &controllers.RoleController{}, "GET:GetRoleByIDs")
	beego.Router("/api:v([2]{0,1})/roles", &controllers.RoleController{}, "POST:AddRole")
	beego.Router("/api:v([2]{0,1})/roles/:role_id", &controllers.RoleController{}, "DELETE:DeleteRole")
	beego.Router("/api:v([2]{0,1})/roles", &controllers.RoleController{}, "PUT:UpdateRole")
	beego.Router("/api:v([2]{0,1})/roles/:children_ids", &controllers.RoleController{}, "POST:InsertRole")
	beego.Router("/api:v([2]{0,1}/roleTree", &controllers.RoleController{}, "GET:GetRoleTreeByID")

	// resource
	beego.Router("/api:v([2]{0,1})/resources", &controllers.ResourceController{}, "GET:GetResourcesByClient")
	beego.Router("/api:v([2]{0,1})/resources", &controllers.ResourceController{}, "POST:AddResources")
	beego.Router("/api:v([2]{0,1})/resources", &controllers.ResourceController{}, "PUT:UpdateResources")
	beego.Router("/api:v([2]{0,1})/resources/list", &controllers.ResourceController{}, "GET:GetResourcesByIDs")
	beego.Router("/api:v([2]{0,1})/resources/:ids", &controllers.ResourceController{}, "DELETE:DeleteResources")
	beego.Router("/api:v([2]{0,1})/userResources", &controllers.ResourceController{}, "GET:GetResourcesByUserAndClient")

	// role_resource
	beego.Router("/api:v([2]{0,1})/roleResources/?:role_id", &controllers.RoleResourceController{}, "GET:GetRoleResource")
	beego.Router("/api:v([2]{0,1})/roleResources/:role_id", &controllers.RoleResourceController{}, "POST:AddRoleResource")
	beego.Router("/api:v([2]{0,1})/roleResources/:role_id", &controllers.RoleResourceController{}, "PUT:UpdateRoleResource")
	beego.Router("/api:v([2]{0,1})/roleResources/:role_id", &controllers.RoleResourceController{}, "Delete:DeleteRoleResource")

	// role_user
	beego.Router("/api:v([2]{0,1})/roleUsers/?:role_id", &controllers.RoleUserController{}, "GET:GetRoleUser")
	beego.Router("/api:v([2]{0,1})/roleUsers/:role_id", &controllers.RoleUserController{}, "POST:AddRoleUser")
	beego.Router("/api:v([2]{0,1})/roleUsers", &controllers.RoleUserController{}, "POST:AddRoleUserBatchImpl")
	beego.Router("/api:v([2]{0,1})/roleUsers", &controllers.RoleUserController{}, "DELETE:DeleteRoleUserBatchImpl")
	beego.Router("/api:v([2]{0,1})/roleUsers/:role_id", &controllers.RoleUserController{}, "PUT:UpdateRoleUser")
	beego.Router("/api:v([2]{0,1})/roleUsers/:role_id", &controllers.RoleUserController{}, "DELETE:DeleteRoleUser")

	// client
	beego.Router("/api:v([2]{0,1})/allClients", &controllers.ClientController{}, "GET:GetAllClients")
	beego.Router("/api:v([2]{0,1})/userClients/?:user_id", &controllers.ClientController{}, "GET:GetClientsByUser")
	beego.Router("/api:v([2]{0,1})/client", &controllers.ClientController{}, "POST:CreateClient")
	beego.Router("/api:v([2]{0,1})/client", &controllers.ClientController{}, "PUT:UpdateClient")
	beego.Router("/api:v([2]{0,1})/client", &controllers.ClientController{}, "GET:GetClient")
	beego.Router("/api:v([2]{0,1})/client/clone", &controllers.ClientController{}, "POST:ClientClone")

	// user
	beego.Router("/api:v([2]{0,1})/user/?:user_id", &controllers.UserController{}, "GET:GetUser")
	beego.Router("/api:v([2]{0,1})/users/members", &controllers.UserController{}, "GET:GetMemberUsers")
	beego.Router("/api:v([2]{0,1})/users", &controllers.UserController{}, "GET:GetAllUsers")
	beego.Router("/api:v([2]{0,1})/users", &controllers.UserController{}, "POST:CreateUser")
	beego.Router("/api:v([2]{0,1})/users/?:user_id", &controllers.UserController{}, "PUT:UpdateUser")
	beego.Router("/api:v([2]{0,1})/users/:user_id/status/:status", &controllers.UserController{}, "PUT:UpdateUserStatus")
	beego.Router("/api:v([2]{0,1})/users/pwds", &controllers.UserController{}, "POST:GetUserEncryptedPwds")

	// group
	beego.Router("/api:v([2]{0,1})/group", &controllers.GroupController{}, "GET:GetGroup")
	beego.Router("/api:v([2]{0,1})/group", &controllers.GroupController{}, "POST:CreateGroup")
	beego.Router("/api:v([2]{0,1})/group", &controllers.GroupController{}, "PUT:UpdateGroup")
	beego.Router("/api:v([2]{0,1})/group", &controllers.GroupController{}, "DELETE:DeleteGroup")

	// group user
	beego.Router("/api:v([2]{0,1})/groupuser/group", &controllers.GroupUserController{}, "GET:GetGroupByUser")
	beego.Router("/api:v([2]{0,1})/groupuser/user", &controllers.GroupUserController{}, "GET:GetUserByGroup")
	beego.Router("/api:v([2]{0,1})/groupuser", &controllers.GroupUserController{}, "POST:AddGroupUser")
	beego.Router("/api:v([2]{0,1})/groupuser", &controllers.GroupUserController{}, "DELETE:RemoveGroupUser")

	// default
	beego.Router("/auth/login", &controllers.BaseController{}, "GET,POST:Login")
	beego.Router("/", &controllers.BaseController{}, "GET:Login")
	beego.Router("/auth/logout", &controllers.BaseController{}, "GET:Logout")
	beego.Router("/frontend/*", &controllers.BaseController{}, "GET:Frontend")
	beego.Router("/captcha", &controllers.BaseController{}, "GET:Captcha")

	//lain health check
	beego.Router("/echo", &controllers.BaseController{}, "GET:Echo")

	// System setting
	beego.Router("/system/setting", &controllers.SystemController{}, "GET:SystemSetting")

	// Sync
	beego.Router("/sync", &controllers.SyncController{}, "POST:Sync")
	beego.Router("/syncClients", &controllers.SyncController{}, "POST:SyncInClients")

}
