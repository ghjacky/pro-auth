package authsdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	SUCC    = 0
	FAILED  = 1
	UNKNOWN = 2

	R_CLIENT         = "client"
	R_GROUP          = "group"
	R_GROUPUSER      = "groupuser"
	R_RESOURCES      = "resources"
	R_USER_RESOURCES = "userResources"
	R_ROLE           = "roles"
	R_ROLE_RESOURCES = "roleResources"
	R_USER           = "users"
	R_USER_ROLES     = "userRoles"
	R_ROLE_USER      = "roleUsers"
	R_USER_CLIENT    = "userClients"

	RoleTypeNormal = "normal"
	RoleTypeAdmin  = "admin"
	RoleTypeSuper  = "super"
)

type APIAuth struct {
	config *APIConfig `json:"config"`
	client *http.Client
}

type APIConfig struct {
	ClientID     int64  `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	APIHost      string `json:"api_host"`
}

func NewApiAuth(config *APIConfig) *APIAuth {
	if config == nil || config.APIHost == "" {
		panic("sso service init failed: config counld not be nil")
	}

	if config.ClientID == 0 || config.ClientSecret == "" {
		panic("Client config is empty!")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	apiAuth := &APIAuth{
		config,
		client,
	}
	return apiAuth
}

// 统一对接口返回结果进行处理，将有效数据部分序列化后返回
func processResp(response *http.Response) (data []byte, err error) {
	var statusErr error
	if response.StatusCode != http.StatusOK {
		statusErr = errors.New("unexpected status code of " + strconv.Itoa(response.StatusCode))
	}
	rawBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body error when process response: %v, statusErr: %v", err, statusErr)
	}
	var body RespBody
	if err = json.Unmarshal(rawBody, &body); err != nil {
		return nil, fmt.Errorf("Unmarshal body error when process response: %v, statusErr: %v", err, statusErr)
	}
	if body.ResCode != SUCC && body.Data == nil {
		return nil, fmt.Errorf("Unexpected response body, msg: %v, statusErr: %v", body.ResMsg, statusErr)
	}
	if statusErr != nil {
		return nil, statusErr
	}
	if data, err = json.Marshal(body.Data); err != nil {
		return nil, err
	}
	return
}

func (a *APIAuth) doRequest(method, resource string, bodys ...[]byte) ([]byte, error) {
	var body io.Reader
	if bodys != nil && len(bodys) > 0 {
		body = ioutil.NopCloser(bytes.NewBuffer(bodys[0]))
	}

	req, err := http.NewRequest(method, a.config.APIHost+"/api/"+resource, body)
	if err != nil {
		return nil, fmt.Errorf("Create New Request error: %v", err)
	}

	req.Header.Add(
		"Authorization",
		"Client "+GenerateJWTToken(a.config.ClientID, a.config.ClientSecret))

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do request error: %v", err)
	}
	defer resp.Body.Close()
	return processResp(resp)
}

func (a *APIAuth) GET(path string) ([]byte, error) {
	return a.doRequest("GET", path)
}

func (a *APIAuth) PUT(path string, body []byte) ([]byte, error) {
	return a.doRequest("PUT", path, body)
}

func (a *APIAuth) POST(path string, body []byte) ([]byte, error) {
	return a.doRequest("POST", path, body)
}

func (a *APIAuth) DELETE(path string, body ...[]byte) ([]byte, error) {
	return a.doRequest("DELETE", path, body...)
}

// GetClientById 通过ClientId查询Client
func (a *APIAuth) GetClient() (*Client, error) {
	data, err := a.GET(R_CLIENT)
	if err != nil {
		return nil, err
	}

	client := &Client{}
	if err = json.Unmarshal(data, client); err != nil {
		return nil, fmt.Errorf("Unmarshal error: %v", err)
	}
	return client, nil
}

// 更新Client
func (a *APIAuth) UpdateClient(fullname, redirectUri string) (*ClientInfo, error) {
	body := make(map[string]interface{})
	body["fullname"] = fullname
	body["redirect_uri"] = redirectUri
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	data, err := a.PUT(R_CLIENT, b)
	if err != nil {
		return nil, err
	}
	clientInfo := ClientInfo{}
	if err = json.Unmarshal(data, &clientInfo); err != nil {
		return nil, err
	}
	return &clientInfo, nil
}

// 查询某用户在指定类型角色下所在的Client
func (a *APIAuth) GetClientByUser(userId, roleType string) ([]*UserClient, error) {
	data, err := a.GET(R_USER_CLIENT + "?user_id=" + userId + "&role_type=" + roleType)
	if err != nil {
		return nil, err
	}
	var userClients []*UserClient
	if err = json.Unmarshal(data, &userClients); err != nil {
		return nil, err
	}
	return userClients, nil
}

// 新增子角色
func (a *APIAuth) AddRole(name string, description string, parentId int) (int, error) {
	body := map[string]interface{}{
		"name":        name,
		"description": description,
		"parent_id":   parentId,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return -1, err
	}

	data, err := a.POST(R_ROLE, b)
	if err != nil {
		return -1, err
	}
	var id int
	if err = json.Unmarshal(data, &id); err != nil {
		return -1, err
	}
	return id, nil
}

// 查询全部角色
func (a *APIAuth) GetAllRole(relatedResource, relatedUser, isTree bool) ([]*Role, error) {
	path := R_ROLE +
		"?is_tree=" + strconv.FormatBool(isTree) +
		"&relate_user=" + strconv.FormatBool(relatedUser) +
		"&relate_resource=" + strconv.FormatBool(relatedResource)
	data, err := a.GET(path)
	if err != nil {
		return nil, err
	}
	var roles []*Role
	if err = json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, err
}

func (a *APIAuth) GetRoleByIDs(ids []int64) ([]*Role, error) {
	if len(ids) == 0 {
		return []*Role{}, nil
	}

	idsstr := make([]string, len(ids))
	for index, id := range ids {
		idsstr[index] = strconv.Itoa(int(id))
	}
	path := "roles/batch?role_ids=" + strings.Join(idsstr, ",")

	data, err := a.GET(path)
	if err != nil {
		return nil, err
	}
	var roles []*Role
	if err = json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, err
}

// SearchRoles 搜索角色，返回角色列表
func (a *APIAuth) SearchRoles(roleName string, relatedResource, relatedUser bool) ([]*Role, error) {
	path := "roles/search" +
		"?role_name=" + roleName +
		"&relate_user=" + strconv.FormatBool(relatedUser) +
		"&relate_resource=" + strconv.FormatBool(relatedResource)
	data, err := a.GET(path)
	if err != nil {
		return nil, err
	}
	var roles []*Role
	if err = json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, err
}

func (a *APIAuth) GetRoleTreeByID(roleID int, relateChildren, relateUser, relateResource, isTree bool) ([]*Role, error) {
	path := "roleTree" +
		"?is_tree=" + strconv.FormatBool(isTree) +
		"&role_id=" + strconv.Itoa(roleID) +
		"&relate_user=" + strconv.FormatBool(relateUser) +
		"&relate_children=" + strconv.FormatBool(relateChildren) +
		"&relate_resource=" + strconv.FormatBool(relateResource)

	data, err := a.GET(path)
	if err != nil {
		return nil, err
	}
	var roles []*Role
	if err = json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

// 修改角色信息
func (a *APIAuth) UpdateRole(roleId int, name, description string) (*Role, error) {
	body := map[string]interface{}{
		"id":          roleId,
		"name":        name,
		"description": description,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	data, err := a.PUT(R_ROLE, b)
	if err != nil {
		return nil, err
	}
	newRole := Role{}
	if err = json.Unmarshal(data, &newRole); err != nil {
		return nil, err
	}
	return &newRole, nil
}

// 删除单个角色
func (a *APIAuth) DeleteRole(roleId int) (*DeleteRoleInfo, error) {
	data, err := a.DELETE(R_ROLE + "/" + strconv.Itoa(roleId))
	if err != nil {
		return nil, err
	}
	deleteRole := DeleteRoleInfo{}
	if err = json.Unmarshal(data, &deleteRole); err != nil {
		return nil, err
	}
	return &deleteRole, err
}

// 批量新增资源
func (a *APIAuth) AddResource(resources []ResourceInfo) ([]int, error) {
	b, err := json.Marshal(resources)
	if err != nil {
		return nil, err
	}
	data, err := a.POST(R_RESOURCES+"/", b)
	if err != nil {
		return nil, err
	}
	var ids []int
	if err = json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// 查看Client下全部资源
func (a *APIAuth) GetAllResources() ([]*ApiResource, error) {
	data, err := a.GET(R_RESOURCES)
	if err != nil {
		return nil, err
	}
	var resources []*ApiResource
	if err = json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// 修改单个资源内容
func (a *APIAuth) UpdateResource(rId int, rName, rDescription, rData string) (*ApiResource, error) {
	body := map[string]interface{}{
		"id":          rId,
		"name":        rName,
		"description": rDescription,
		"data":        rData,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	data, err := a.PUT(R_RESOURCES+"/", b)
	if err != nil {
		return nil, err
	}
	var resource ApiResource
	if err = json.Unmarshal(data, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// 批量删除资源
func (a *APIAuth) DeleteResources(resourceIds []int) (*DeleteResInfo, error) {
	path := R_RESOURCES + "/"
	for idx, id := range resourceIds {
		if idx != 0 {
			path += ","
		}
		path += strconv.Itoa(id)
	}
	data, err := a.DELETE(path)
	if err != nil {
		return nil, err
	}
	delInfo := DeleteResInfo{}
	if err = json.Unmarshal(data, &delInfo); err != nil {
		return nil, err
	}
	return &delInfo, nil
}

// 向某角色内批量添加用户，返回添加数量
func (a *APIAuth) AddUserToRole(roleId int, infos []UserInfo) (int, error) {
	b, err := json.Marshal(infos)
	if err != nil {
		return -1, err
	}
	data, err := a.POST(R_ROLE_USER+"/"+strconv.Itoa(roleId), b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 向批量添加用户角色关系，返回添加数量
func (a *APIAuth) AddRoleUser(infos []RoleUser) (int, error) {
	b, err := json.Marshal(infos)
	if err != nil {
		return -1, err
	}
	data, err := a.POST(R_ROLE_USER, b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 查询角色中的用户
func (a *APIAuth) GetUsersOfRole(roleId int) ([]*RoleUser, error) {
	data, err := a.GET(R_ROLE_USER + "/" + strconv.Itoa(roleId))
	if err != nil {
		return nil, err
	}
	var users []*RoleUser
	if err = json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, err
}

// 查询指定用户角色（直接关联的或全部）
func (a *APIAuth) GetUserRoles(userId string, isAll, relatedResource, relatedUser, isTree bool) ([]*Role, error) {
	path := R_USER_ROLES +
		"?is_all=" + strconv.FormatBool(isAll) +
		"&is_tree=" + strconv.FormatBool(isTree) +
		"&user_id=" + userId +
		"&relate_user=" + strconv.FormatBool(relatedUser) +
		"&relate_resource=" + strconv.FormatBool(relatedResource)

	data, err := a.GET(path)
	var roles []*Role
	if err = json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, err
}

func (a *APIAuth) GetUserRolesByUserID(userIDs []string) ([]*RoleUser, error) {
	path := R_USER_ROLES + "/users?user_ids=" + strings.Join(userIDs, ",")

	data, err := a.GET(path)
	var roleUser []*RoleUser
	if err = json.Unmarshal(data, &roleUser); err != nil {
		return nil, err
	}

	return roleUser, err
}

// 修改单个用户信息，返回修改后的用户
func (a *APIAuth) UpdateUserOfRole(roleId int, info UserInfo) (*RoleUser, error) {
	b, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	data, err := a.PUT(R_ROLE_USER+"/"+strconv.Itoa(roleId), b)
	if err != nil {
		return nil, err
	}
	user := RoleUser{}
	if err = json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, err
}

// 批量删除用户角色关系
func (a *APIAuth) DeleteRoleUser(infos []RoleUser) (int, error) {
	b, err := json.Marshal(infos)
	if err != nil {
		return -1, err
	}
	data, err := a.DELETE(R_ROLE_USER, b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 批量删除某角色内用户，返回删除人数
func (a *APIAuth) DeleteUserFromRole(roleId int, names []string) (int, error) {
	b, err := json.Marshal(names)
	if err != nil {
		return -1, err
	}
	data, err := a.DELETE(R_ROLE_USER+"/"+strconv.Itoa(roleId), b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 批量添加某角色和资源关联关系，返回新增关联数目
func (a *APIAuth) AddRoleResourceRelations(roleId int, resIds []int) (int, error) {
	b, err := json.Marshal(resIds)
	if err != nil {
		return -1, err
	}
	data, err := a.POST(R_ROLE_RESOURCES+"/"+strconv.Itoa(roleId), b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 查看Client下全部角色资源关联
func (a *APIAuth) GetAllRoleResourceRelatedInfo() ([]*RelatedInfo, error) {
	data, err := a.GET(R_ROLE_RESOURCES)
	if err != nil {
		return nil, err
	}
	var info []*RelatedInfo
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return info, err
}

// 查看指定角色关联的所有资源
func (a *APIAuth) GetRoleResourceRelatedInfo(roleId int) ([]*RelatedInfo, error) {
	data, err := a.GET(R_ROLE_RESOURCES + "/" + strconv.Itoa(roleId))
	if err != nil {
		return nil, err
	}
	var info []*RelatedInfo
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return info, err
}

// 查看指定角色关联的所有资源
func (a *APIAuth) GetResourceByRole(roleId int) ([]*ApiResource, error) {
	relations, err := a.GetRoleResourceRelatedInfo(roleId)
	if err != nil {
		return nil, err
	}

	if len(relations) == 0 {
		return []*ApiResource{}, nil
	}

	resourceIDs := []string{}
	for _, relation := range relations {
		resourceIDs = append(resourceIDs, "id="+strconv.Itoa(relation.ResourceId))
	}

	data, err := a.GET(R_RESOURCES + "/list?" + strings.Join(resourceIDs, "&"))
	if err != nil {
		return nil, err
	}

	var resources []*ApiResource
	if err = json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}
	return resources, err
}

// 查看用户在Client下的全部资源
func (a *APIAuth) GetUserResources(userId string) ([]*ApiResource, error) {
	data, err := a.GET(R_USER_RESOURCES + "?user_id=" + userId)
	if err != nil {
		return nil, err
	}
	var resources []*ApiResource
	if err = json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// 批量修改某角色和资源关联关系，返回当前全部关联数目
func (a *APIAuth) UpdateRoleResourceRelations(roleId int, resIds []int) (int, error) {
	b, err := json.Marshal(resIds)
	if err != nil {
		return -1, err
	}
	data, err := a.PUT(R_ROLE_RESOURCES+"/"+strconv.Itoa(roleId), b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 批量删除某角色和资源关联关系，返回删除的关联数目
func (a *APIAuth) DeleteRoleResourceRelations(roleId int, resIds []int) (int, error) {
	b, err := json.Marshal(resIds)
	if err != nil {
		return -1, err
	}
	data, err := a.DELETE(R_ROLE_RESOURCES+"/"+strconv.Itoa(roleId), b)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

func (a *APIAuth) CheckThirdToken(accessToken, clientID, userID string) (bool, error) {
	data, err := a.GET("/user/check?" + "access_token=" + accessToken + "&client_id=" + clientID + "&user_id=" + userID)
	if err != nil {
		return false, err
	}
	var success bool
	if err = json.Unmarshal(data, &success); err != nil {
		return false, err
	}
	return success, nil
}

func (a *APIAuth) CreateGroup(name, description, email string) error {
	group := Group{
		Name:        name,
		Description: description,
		Email:       email,
	}

	b, err := json.Marshal(group)
	if err != nil {
		return err
	}

	_, err = a.POST(R_GROUP, b)
	return err
}

func (a *APIAuth) GetGroup(name string) ([]Group, error) {
	res, err := a.GET(R_GROUP + "?name=" + name)
	if err != nil {
		return nil, err
	}
	groups := []Group{}
	err = json.Unmarshal(res, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (a *APIAuth) SearchGroups(name string) ([]Group, error) {
	res, err := a.GET(R_GROUP + "?likes=" + name)
	if err != nil {
		return nil, err
	}
	groups := []Group{}
	err = json.Unmarshal(res, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (a *APIAuth) GetAllGroup() ([]Group, error) {
	res, err := a.GET(R_GROUP)
	if err != nil {
		return nil, err
	}
	groups := []Group{}
	err = json.Unmarshal(res, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (a *APIAuth) UpdateGroup(name, description, email string) error {
	group := Group{
		Name:        name,
		Description: description,
		Email:       email,
	}
	b, err := json.Marshal(group)
	if err != nil {
		return err
	}

	_, err = a.PUT(R_GROUP, b)
	return err
}

func (a *APIAuth) DeleteGroup(name string) error {
	_, err := a.DELETE(R_GROUP + "?name=" + name)
	return err
}

func (a *APIAuth) CreateGroupUser(groupName, userID string) error {
	gu := GroupUser{
		Group: groupName,
		User:  userID,
	}

	b, err := json.Marshal(gu)
	if err != nil {
		return err
	}

	_, err = a.POST(R_GROUPUSER, b)
	return err
}

func (a *APIAuth) GetGroupsByUserID(userID string) ([]Group, error) {
	res, err := a.GET(R_GROUPUSER + "/group?user=" + userID)
	if err != nil {
		return nil, err
	}

	groups := []Group{}
	err = json.Unmarshal(res, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (a *APIAuth) GetUsersByGroup(name string) ([]User, error) {
	res, err := a.GET(R_GROUPUSER + "/user?group=" + name)
	if err != nil {
		return nil, err
	}

	users := []User{}
	err = json.Unmarshal(res, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (a *APIAuth) DeleteGroupUser(groupName, userID string) error {
	gu := GroupUser{
		Group: groupName,
		User:  userID,
	}

	b, err := json.Marshal(gu)
	if err != nil {
		return err
	}

	_, err = a.DELETE(R_GROUPUSER, b)
	return err
}

// 接口获取的完整resource
type ApiResource struct {
	Id          int    `json:"id"`          // 资源id
	Name        string `json:"name"`        // 资源名
	Description string `json:"description"` // 资源描述
	ClientId    int    `json:"client_id"`   // ClientId
	Data        string `json:"data"`        // 资源内容
	CreatedBy   string `json:"created_by"`  // 创建者
	UpdatedBy   string `json:"updated_by"`  // 更新者
	Created     string `json:"created"`     // 创建时间
	Updated     string `json:"updated"`     // 更新时间
}

type RoleResource struct {
	Resource ApiResource `json:"resource"`
	RoleID   int         `json:"role_id"`
}

// 添加或修改Resource后，接口返回的新Resource信息
type ResourceInfo struct {
	Name        string `json:"name"`        // 资源名称
	Description string `json:"description"` // 资源描述
	Data        string `json:"data"`        // 资源内容
}

// 删除Resource后，接口返回的删除信息
type DeleteResInfo struct {
	DelResNum     int `json:"del_resource_num"`
	DelRoleResNum int `json:"del_role_resource_num"`
}

// 用户所在的Client
type UserClient struct {
	Id       int     `json:"id"`       // clientId
	Fullname string  `json:"fullname"` // client全名
	Roles    []*Role `json:"roles"`    // 用户在client下的角色
}

// 更新Client后，接口返回的新Client信息
type ClientInfo struct {
	Id          int    `json:"id"`           // clientId
	Fullname    string `json:"fullname"`     // client全名
	RedirectUri string `json:"redirect_uri"` // 重定向uri
}

type RespBody struct {
	ResCode int         `json:"res_code"` // auth接口返回状态码，SUCC-0-成功  FAILED-1-失败  UNKNOWN-2-其他
	ResMsg  string      `json:"res_msg"`  // auth接口返回状态描述
	Data    interface{} `json:"data"`     // auth接口返回数据
}

// 通过ClientId所查询到的Client
type Client struct {
	Id          int    `json:"id"`           // clientId
	Fullname    string `json:"fullname"`     // client全名
	Secret      string `json:"secret"`       // ClientSecret
	RedirectUri string `json:"redirect_uri"` // 重定向uri
	UserId      string `json:"user_id"`      // 用户Id
	Created     string `json:"created_at"`   // 创建时间
	Updated     string `json:"updated_at"`   // 更新时间
}

// 角色
type Role struct {
	Id          int             `json:"id"`          // 角色id
	Name        string          `json:"name"`        // 角色名
	Description string          `json:"description"` // 角色类型
	ParentId    int             `json:"parent_id"`   // 父角色id
	CreatedBy   string          `json:"created_by"`  // 创建者
	UpdatedBy   string          `json:"updated_by"`  // 更新者
	Created     string          `json:"created"`     // 创建时间
	Updated     string          `json:"updated"`     // 更新时间
	RoleType    string          `json:"role_type"`   // 角色类型
	Resources   []*RoleResource `json:"resources"`   // 角色相关资源
	Users       []*UserOfRole   `json:"users"`       // 角色相关用户
	Children    []*Role         `json:"children"`    // 拥有该角色的用户
}

// 角色相关用户
type UserOfRole struct {
	RoleId   int64    `json:"role_id"`
	User     *ApiUser `json:"user"`
	RoleType string   `json:"role_type"`
}

// 删除角色后，接口返回的删除信息
type DeleteRoleInfo struct {
	DelRoleNum         int `json:"del_role_num"`          // 删除的角色数量
	DelRoleResourceNum int `json:"del_role_resource_num"` // 删除的相关资源数量
	DelRoleUserNum     int `json:"del_role_user_num"`     // 删除的相关用户数量
}

// 角色中的用户
type RoleUser struct {
	RoleId   int    `json:"role_id"`   // 角色Id
	UserId   string `json:"user_id"`   // 用户Id
	RoleType string `json:"role_type"` // 角色类型
}

// 用户基本信息
type UserInfo struct {
	UserId   string `json:"user_id"`   // 用户Id
	RoleType string `json:"role_type"` // 角色类型
}

// 角色资源关联信息
type RelatedInfo struct {
	RoleId     int `json:"role_id"`     // 角色id
	ResourceId int `json:"resource_id"` // 角色关联资源id
}

type ApiUser struct {
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email,omitempty"`
	Wechat   string `json:"wechat,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Type     string `json:"type,omitempty"`
}

type User struct {
	// user
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email,omitempty"`
	Wechat   string `json:"wechat,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`

	Dn string `json:"dn"`

	// resource
	Resources   []*Resource          `json:"resource"`
	ResourceMap map[string]*Resource `json:"resourceMap"`
	CacheTime   int64                `json:"cacheTime"`

	Organization string `json:"organization"`

	// token
	Token     Token      `json:"-"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type Resource struct {
	Id          int64  `json:"id"`
	Description string `json:"description"`
	Data        string `json:"data"`
}

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Group struct {
	ID          int64  `orm:"pk;auto" json:"id"`
	Name        string `orm:"size(128);unique" json:"name,omitempty"`
	Email       string `orm:"size(128)" json:"email,omitempty"`
	Description string `orm:"size(256)" json:"description,omitempty"`
}

type GroupUser struct {
	ID    int64  `orm:"pk;auto" json:"id"`
	Group string `orm:"size(128)" json:"group,omitempty"`
	User  string `orm:"size(256)" json:"user,omitempty"`
}
