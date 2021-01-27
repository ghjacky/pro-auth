package consts

const (
	AUTH_KEY               = "authinfo"
	AUTH_BY_SECRET_USER_Id = "client"
)
const (
	NO_AUTH = iota
	AUTH_BY_TOKEN
	AUTH_BY_SECRET
	AUTH_BY_SESSION
)
