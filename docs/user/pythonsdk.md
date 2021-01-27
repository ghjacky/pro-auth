# Python auth sdk document

## Notice

### Visit API

**More exmaple is in [here](sdk/python/example/main.py)**

To use the sdk, you need install the sdk package by pip:

```bash
$ pip install git+https://code.xxxxx.cn/platform/auth_python_sdk.git
```

Then use it as below:

```python
from authsdk import api

def main():
    # create client
    # please do not hardcode the secret
    authClient = api.API(
        client_id=0, # your id
        secret="123456", # your secret
        host="https://sz-bdy-auth.xxxxx.cn", # ac platform address
        redirecturi="https://your app redirect uri", # your redirect uri
    )

    client = authClient.GetClient() # bingo! you could use this client's methods, the types defined in authsdk.types
    print client.name
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
- GetRoleResourceRelatedInfo
- UpdateRoleResourceRelations
- DeleteRoleResourceRelations

- CheckThirdToken

Besides, you could orgnize your user by group.
The API includes:(in sdk)

**Group**

- CreateGroup
- GetGroup
- UpdateGroup
- DeleteGroup

**Group and User**

- CreateGroupUser
- GetGroupsByUserID
- GetUsersByGroup
- DeleteGroupUser

### User Login

First, route the 'redirect uri' requests which filled when you create your application to your controller and pass the code in param to Login function.  

```python
from authsdk import oauth

oauthService = oauth.Oauth(
    client_id=args.id,
    secret=args.secret,
    redirect_uri=args.redirect,
    host=args.host
)

url = oauthService.GenerateLoginURL(state="/")
# and return a redirect response(302) with this url to your frontend
```

Then you could get your user with code

```python
from authsdk import oauth

oauthService = oauth.Oauth(
    client_id=args.id,
    secret=args.secret,
    redirect_uri=args.redirect,
    host=args.host
)

user = oauthService.Login(code)
```

If you dont know how to get the code, you could generate a link by `authsdk.oauth.LoginURL("state string")` and let broswer to redirect this link when user want to login.  
Then you will get the code in the redirect uri api.

The user you could get only includes some basic infomation, you need do some extend operations to get more info like `LoadUserResource`

```python
from authsdk import oauth
oauthService = oauth.Oauth(
    client_id=args.id,
    secret=args.secret,
    redirect_uri=args.redirect,
    host=args.host
)

user = oauthService.Login(code)
res = oauthService.LoadUserResource(user)
```

Or you could login by user secret without oauth2 code as below. the faq will help you to find your user secret.

```python
from authsdk import oauth
oauthService = oauth.Oauth(
    client_id=args.id,
    secret=args.secret,
    redirect_uri=args.redirect,
    host=args.host
)

user = oauthService.LoginBySecret("username", "user secret")
```

---------------
More exmaple is in [here](sdk/python/example/main.py)

Suppose your app config is this:
```json
{
    "id": 2,
    "secret": "anpfEJX5ZtrbI792_ecZvw",
}
```
And the ac address is `http://172.16.244.6:30099`

Then example image usage: `docker run --rm registry.xxxxx.cn/platform/auth_example:python python example/main.py -id=2 -secret=anpfEJX5ZtrbI792_ecZvw -host=http://172.16.244.6:30099`  
or `docker run --rm harbor.xxxxx.cn/platform/auth_example:python3 python example/main.py -id=2 -secret=anpfEJX5ZtrbI792_ecZvw -host=http://172.16.244.6:30099` for python3

## FAQ

see [faq](faq.md)
