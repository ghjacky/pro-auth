package errors

import "errors"

type ErrValue struct {
	StatusCode int
	ErrMsg     string
}

var (
	// 验证登录身份
	ErrNoAuth    = errors.New("no_auth")
	ErrWrongAuth = errors.New("wrong_auth")

	// 通用
	ErrInvalidRequest      = errors.New("invalid_request")
	ErrUnauthorizedRequest = errors.New("unauthorized_request")
	ErrServerError         = errors.New("server_error")
	ErrNoAuthorityAPI      = errors.New("this account has no authority to visit the api")

	// client
	ErrClientNotFound        = errors.New("client_not_found")
	ErrClientFullNameExisted = errors.New("client_fullname_existed")
	ErrClientFullNameInvalid = errors.New("client_fullname_invalid")

	// role
	ErrRoleNotFound               = errors.New("role_not_found")
	ErrParentRoleNotFound         = errors.New("parent_role_not_found")
	ErrParentRoleInvalid          = errors.New("parent_role_is_invalid")
	ErrLeafRoleHasNoSubRole       = errors.New("leaf_role_have_no_subRole")
	ErrRootRoleCannotModifyParent = errors.New("root_role_cannot_modify_parent")
	ErrRootRoleCannotBeDeleted    = errors.New("root_role_cannot_be_deleted")

	// resource
	ErrResourceNotFound = errors.New("resource_not_found")

	ErrUserNotFound  = errors.New("user not found")
	ErrUserExisted   = errors.New("user_id_existed")
	ErrUserNotActive = errors.New("账户已冻结或注销，请联系管理员")

	// role-resource
	ErrRoleResourceNotFound     = errors.New("role_resource_not_found")
	ErrRoleResourceCannotModify = errors.New("role_resource_cannot_modify")

	// role-user
	ErrUserNotFoundInRole            = errors.New("user_not_found_in_role")
	ErrNoUnauthorizedToRoleTypeSuper = errors.New("unauthorized_to_role_type_super")
	ErrRoleUserExisted               = errors.New("role_user_existed")
)

var ErrValueMap = map[error]ErrValue{
	ErrNoAuth:    {401, "The request session is not login,  not includes a valid Authorization in request header"},
	ErrWrongAuth: {403, "The request includes a wrong type Authorization in request header"},

	ErrInvalidRequest:      {400, "The request is missing a required parameter or includes an invalid parameter value"},
	ErrUnauthorizedRequest: {403, "The server denied the request, request has insufficient permissions"},
	ErrServerError:         {500, "The unexpected error occurred"},

	ErrClientNotFound:        {400, "The client not found , maybe request with a wrong client_id "},
	ErrClientFullNameExisted: {400, "The client fullname existed"},
	ErrClientFullNameInvalid: {400, "The client fullname invalid"},

	ErrRoleNotFound:               {400, "role not found"},
	ErrParentRoleNotFound:         {400, "parent role not found or parent role not in this client"},
	ErrParentRoleInvalid:          {400, "parent role is invalid"},
	ErrLeafRoleHasNoSubRole:       {400, "leaf role (the role has relation with resource) can't have subRole"},
	ErrRootRoleCannotModifyParent: {400, "root role can't modify parent"},
	ErrRootRoleCannotBeDeleted:    {400, "root role can't be deleted"},

	ErrResourceNotFound:         {400, "resource not found or resource not in the client"},
	ErrRoleResourceCannotModify: {400, "root role cannot modify resource"},

	ErrUserNotFound: {400, "user not found"},

	ErrRoleResourceNotFound: {400, "one or more relation in role and resource not found "},

	ErrUserNotFoundInRole:            {400, "one or more user not found in this role"},
	ErrNoUnauthorizedToRoleTypeSuper: {403, "no unauthorized to add or delete or update super user"},
	ErrRoleUserExisted:               {400, "one or more user have relation to the role"},
}
