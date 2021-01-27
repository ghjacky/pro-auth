# GO auth sdk document

## Notice

**This document needs to be completed!.**

### Visit API

**More exmaple is in [here](sdk/golang/example/example.go)**

To use the sdk, you need import the sdk package in your code:

```go
import (
    authsdk code.xxxxx.cn/platform/auth/sdk/golang
)
```

suggest use ssh rather than https to use dep likes:   
`git config --global url."git@code.xxxxx.cn:".insteadOf "https://code.xxxxx.cn"`

Then use it as below:

```go
func main() {
    // create auth config
    // please do not hardcode the secret
    config := authsdk.APIConfig{
        ClientID:       0, // your id
        ClientSecret:   "123456", // yoursecret
        APIHost:        "https://ac.xxxxx.cn", // ac platform address
    }
    // create auth client
    authClient := authsdk.NewApiAuth(config)
    
    // use the client
    client, err := authClient.GetClient()
    if err == nil {
        // Bingo! do what you want to do.
    }
}
```

### Authority Management

Currently, you need to design your roles tree by yourself, and using the roles tree to manage your user authority.  
The API includes:(in sdk)

**Client**

- GetClient
- UpdateClient
- GetClientByUser

**Role**

- AddRole
- GetAllRole
- UpdateRole
- DeleteRole

**User and Role**

- AddUserToRole
- GetUsersOfRole
- GetUserRoles
- UpdateUserOfRole
- DeleteUserFromRole

**Resource**

- AddResource
- GetAllResources
- UpdateResource

**Role,Resource and User**

- AddRoleResourceRelations
- GetResourceByRole
- GetUserResources
- GetAllRoleResourceRelatedInfo
- GetRoleResourceRelatedInfo
- UpdateRoleResourceRelations
- DeleteRoleResourceRelations

- CheckThirdToken

Besides, you could orgnize your user by group.
The API includes:(in sdk)

**Group**

- CreateGroup
- GetGroup
- GetAllGroup
- UpdateGroup
- DeleteGroup

**Group and User**

- CreateGroupUser
- GetGroupsByUserID
- GetUsersByGroup
- DeleteGroupUser

### User Login

Same as before, we will use go as display language.
we provide a simple package for you to do oauth steps.

Firstly, import the sdk

```go
import (
    authcustome code.xxxxx.cn/platform/auth/sdk/golang/custome
)
```

Secondly, route the redirect uri which filled when you create your application to your controller and pass the code in param to Login function.  
If you dont know how to get the code, you could generate a link by `authcustome.LoginURL("state string")` and let broswer to redirect this link when user want to login.  
Then you will get the code in the redirect uri api.

```go
config := authcustome.OauthConfig{
	ClientId:     *clientID,
	ClientSecret: *clientSecret,
	RedirectUri:  "",
	Host:         *authHost,
	Scope:        *authScope,
}
service := authcustome.NewAuthService(config)
user, err := service.Login(code)
```

Or you could login by user secret without oauth2 code as below. the faq will help you to find your user secret.

```go
config := authcustome.OauthConfig{
	ClientId:     *clientID,
	ClientSecret: *clientSecret,
	RedirectUri:  "",
	Host:         *authHost,
	Scope:        *authScope,
}
service := authcustome.NewAuthService(config)
user, err := service.LoginBySecret("username", "user secret")
```

The user you could get only includes some basic infomation, you need do some extend operations to get more info like `LoadUserResource`

```go
res, err := service.LoadUserResource(user)
```

---------------
More exmaple is in [here](sdk/golang/example/example.go)

Suppose your app config is this:
```json
{
    "id": 2,
    "secret": "anpfEJX5ZtrbI792_ecZvw",
}
```
And the ac address is `http://172.16.244.6:30099`

Then example image usage: `docker run --rm registry.xxxxx.cn/platform/auth_example:latest ./example -id=2 -secret=anpfEJX5ZtrbI792_ecZvw -host=http://172.16.244.6:30099`

## FAQ

see [faq](faq.md)
