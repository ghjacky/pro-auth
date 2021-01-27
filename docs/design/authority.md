# 权限接口

## 一、权限接口的访问身份

访问权限接口的请求授权方式可分为两类：

1. access token(user)

    身份：access token对应用户的身份  
    可访问范围: access token对应client下被授予的scope  
    用途: 接入auth，需要权限管理的项目，例如：查询用户在当前client下的权限  
    备注：写在request 的 header 中，如下  

    ```json
    {
        "Authorization": "Bearer WZSQF5-SOGKVMXVSWI9YBW"
    }
    ```

2. jwt(client)

    身份：该client下超级权限（相当于处于该client中rootRoleSuper的成员）  
    可访问范围：该client下全部scope  
    用途：接入auth的项目，需要自动维护角色-用户-权限三者的关系  
    备注：写在request 的 header 中, 如下：

    ```json
    {
        "Authorization": "Client eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjMiLCJub25jZSI6IjEyOTg0OTgwODEiLCJ0aW1lIjoiMTU2NzQxODQxMiJ9.atwZPiH-syeffsBxyh41_-TpplJKOvQe98_h8i7bedQ"
    }
    ```

    Client空格后为JWT Token，规则是将如下对象进行jwt编码获得，jwt编码详见[sdk/golang/security.go](../../sdk/golang/security.go)，注意JWT有效期为5分钟，失效后需要自行重新生成。
    jwt Payload:

    ```json
    {
        "id": "1",
        "time": "unix time seconds",
        "nonce": "random string"
    }
    ```

    The jwt payload should be signed with your secret by sha256 method.  
    sdk使用示例参见[example](../../sdk/golang/example/example.go)

## 二、接口返回值格式定义

1. 返回状态定义

    服务端首先判断Request是否携带AuthInfo（access token或jwt token），未找到有效AuthInfo则拒绝请求，返回http response code 401，在body中以字符串形式返回拒绝请求原因
    服务端接受并正常处理请求时http response code均为200，返回结果在response body中以application/json; charset=utf-8格式返回

2. 返回值字段格式

    全部字段使用下划线命名法

3. 返回值JSON结构

    字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
    ---| --- | --- | --- | --- | --- |
    res_code|状态码|是|int|0|非0时表示请求失败
    res_msg|错误描述|是|string|"OK"或具体错误信息描述
    data|数据|否|由具体接口决定|接口返回数据|res_code非0时，该字段为null

    请求成功示例

    ```json
    {
        "res_code": 0,
        "res_msg":"OK",
        "data":[
            {
                "id": 1,
                "name": "username",
            }
        ],
    }
    ```

    请求失败示例

    ```json
    {
        "res_code": 2,
        "res_msg":"错误信息描述",
        "data": null
    }
    ```

## 二、接口定义

1. [Client](#Client)  
    [创建client](#创建client)  
    [查询client](#查询client)  
    [更新client](#更新client)  

2. [User](#User)  
    [查询user](#查询user)  
    [查询全部user](#查询全部user)  
    [查询用户密码密文](#查询用户密码密文)
    [分页查询用户列表](#分页查询用户列表)  

3. [Resource](#Resource)  
    [查询resource by client](#查询resource-by-client)  
    [查询resource by client and user](#查询resource-by-client-and-user)  
    [新增resource（批量）](#新增resource批量)  
    [修改resource](#修改resource)  
    [删除resource（批量）](#删除resource批量)  

4. [Role](#Role)  
    [查询role by client](#查询role-by-client)  
    [查询role by client and user](#查询role-by-client-and-user)  
    [新增sub role](#新增sub-role)  
    [更新role](#更新role)  
    [删除role](#删除role)  
    [父子角色之间插入role](#父子角色之间插入role)  

5. [Role User](#Role-User)  
    [查询role中的user](#查询role中的user)  
    [添加user到role（批量）](#添加user到role批量)  
    [修改user在role中的类型](#修改user在role中的类型)  
    [从role中删除user（批量）](#从role中删除user批量)  

6. [Role Resource](#Role-Resource)  
    [查询role关联的resource](#查询role关联的resource)  
    [增加role-resource关联（批量）](#增加role-resource关联批量)  
    [修改role-resource关联（批量）](#修改role-resource关联批量)  
    [删除role—resource关联（批量）](#删除role—resource关联批量)  

7. [Group](#Group)  
    [Group](#Group)

## Client

### 创建Client

**请求方式以及接口地址**  
  `POST /api/client`

**请求权限**  
  scope >= client:write

**body请求参数**  

```json
{
    "fullname":"client_name",
    "redirect_uri":"http://xx.com/xxx"
}
```

**返回结果示例**

```json
{
    "res_code":0,
    "res_msg":"ok",
    "data": {
        "id":20,
        "fullname":"client_name",
        "secret":"3I1zixZmPLdPRYjLtv6lKg",
        "redirect_uri":"http://xx.com/xxx",
        "created_by":"user2",
        "root_role_id":34,
        "root_role_name":"client_name-root"
    },
}
```

### 查询Client

#### 获取token本身关联的Client

**请求方式以及接口地址**  
  `GET /api/client`

**请求权限**  
  scope: >= client:read

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "id": 1,
        "fullname": "client_name",
        "secret": "3I1zixZmPLdPRYjLtv6lKg",
        "redirect_uri": "http://localhost:9094/oauth2",
        "user_id": "user2",
        "created_at": "2018-10-24T08:00:00+08:00",
        "updated_at": "2018-10-25T17:30:32+08:00",
    }
}
```

#### 获取用户有权限的Client

**请求方式以及接口地址**  
  `GET /api/userClients?user_id=user2&role_type=admin`

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
user_id|账号|否|string|user2|默认为session登录的用户
role_type|角色类型|否|string|admin/normal/super|默认为normal

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": 2,
            "fullname": "airflow",
            "roles": [{
                "id": 18,
                "name": "airflow-root",
                "description": "描述",
                "parent_id": -1,
                "created_by": "user2",
                "updated_by": "user2",
                "created": "2018-10-24T08:00:00+08:00",
                "updated": "2018-10-26T11:43:55+08:00",
                "role_type": "normal"
            }]
        }
    ]
}
```

### 更新client

**请求方式以及接口地址**  
  `PUT /api/client`

**请求权限**  
  scope:    >=  client:write

**body参数**

```json
{
    "fullname": "new_client_name",
    "redirect_uri":"new_redirect_uri"
}
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "id": 1,
        "fullname": "client_name",
        "redirect_uri": "http://xx.com/xxx"
    }
}
```

## User

### 查询user

**请求方式以及接口地址**  
  `GET /api/user`

**请求权限**  
   scope:    >=  user:read

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "id": "user2",
        "fullname": "name",
        "dn": "xxx",
        "created_at": "2018-10-24T17:09:36+08:00",
        "updated_at": "2018-10-24T17:09:36+08:00"
    }
}
```

### 查询client的全部user

**请求方式以及接口地址**  
  `GET /api/user?client_id=1`

**请求权限**  
  scope:    >=  user:read

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|是|int|1|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": "user1",
            "fullname": "用户名1",
            "dn": "xx",
            "created_at": "2018-10-25T14:11:25+08:00",
            "updated_at": "2018-10-25T14:11:25+08:00"
        },
        {
            "id": "user2",
            "fullname": "用户名2",
            "dn": "xx",
            "created_at": "2018-10-24T17:09:36+08:00",
            "updated_at": "2018-10-24T17:09:36+08:00"
        }
    ]
}
```

### 查询用户密码密文

**请求方式以及接口地址**  
  `POST /api/users/pwds`

**请求权限**  
  scope:    >=  user:read

**body参数**

```json
[
    "admin", 
    "user-id"
]
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": "admin",
            "password": "$2a$10$asBCIKgVStmS8j6Qh1KODuH59I5Fse7F6Z9m/ClJD2KmGNwtS9Gie"
        },
        {
            "id": "user-id",
            "password": "JDJhJDEwJFBwYnlqSnZsQi9SaHpRUG92NWNTSXU2cnI4aW4yd3plcDlVZ1gwMklxMVBGUVZiOVZJSXoy"
        }
    ]
}
```

### 分页查询用户列表

**请求方式以及接口地址**  
  `GET /api/users?page_size=10&p=1&id=admin&fullname=admin`

**请求权限**  
  scope:    >=  user:read

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
page_size|每页大小|否，默认10|int|1|
p|当前页|否，默认1|int|1|
id|模糊匹配账号|否|string|admin|
fullname|模糊匹配账户全名|否|string|admin|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "paginator": {
            "page_size": 2,   // 每页大小
            "total_size": 6,  // 数据总条数
            "total_pages": 3, // 总页数
            "current_page": 1 // 当前页
        },
        "data": [
            {
                "id": "admin",
                "fullname": "管理员",
                "password": "$2a$10$ZBIfv/dbJBvVMc0Gq2a9KeRq3LVx/nVazcNAv1a18UyRCcbmOtTeu",
                "email": "admin@xxxxx.com",
                "type": "special",
                "dn": "special user",
                "status": "active",
                "created_at": "2019-09-05T11:01:51+08:00",
                "updated_at": "2019-09-05T15:45:44+08:00"
            },
            {
                "id": "airflow",
                "fullname": "airflow",
                "password": "$2a$10$XvXMwa1JEt0xz6NLt4X5tuSVEDCoa6ayvWcSPyzu0eBJIgG1P35fG",
                "email": "airflow@xxxxx.com",
                "type": "special",
                "dn": "special user",
                "status": "active",
                "created_at": "2019-09-05T11:17:55+08:00",
                "updated_at": "2019-09-05T15:45:44+08:00"
            }
        ]
    }
}
```

## Resource

### 查询Resource

**请求权限**  
  scope:    >=  resource:read

**请求方式以及接口地址**  
  `GET /api/resources?client_id=1&relate_role=true`

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|
relate_role|是否查询关联的角色|否|bool|true|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": 8,
            "name": "权限名1",
            "description": "描述",
            "client_id": 1,
            "data": "dag-read",
            "created_by": "user2",
            "updated_by": "user2",
            "created": "2018-10-24T11:47:40+08:00",
            "updated": "2018-10-25T17:09:44+08:00",
            "roles": [
                {
                    "resource_id": 16,
                    "id": 1,
                    "name": "pallas管理员",
                    "description": "描述",
                    "parent_id": -1
                }
            ]
        },
        {
            "id": 9,
            "name": "权限名2",
            "description": "描述",
            "client_id": 1,
            "data": "dag-write",
            "created_by": "user2",
            "updated_by": "测试修改人",
            "created": "2018-10-24T11:47:35+08:00",
            "updated": "2018-10-24T14:44:24+08:00",
            "roles": [
                {
                    "resource_id": 16,
                    "id": 1,
                    "name": "pallas管理员",
                    "description": "描述",
                    "parent_id": -1
                }
            ]
        },
    ]
}
```

### 查询resource by client and user

查询用户在client下的resource

**请求权限**  
  scope:    >=  resource:read

**请求方式以及接口地址**  
  `GET /api/userResources?client_id=1&user_id=user2`

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|
user_id|账号|否|string|user2|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": 8,
            "name": "权限名1",
            "description": "描述",
            "client_id": 1,
            "data": "dag-read",
            "created_by": "user2",
            "updated_by": "user2",
            "created": "2018-10-24T11:47:40+08:00",
            "updated": "2018-10-25T17:09:44+08:00"
        },
    ]
}
```

### 新增resource（批量）

**请求方式以及接口地址**  
  `POST /api/resources?client_id=1`

**请求权限**  
  scope:    >=  resource:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|

**body参数**

```json
[
    {
        "name":"权限1",
        "description":"描述",
        "data":"权限内容"

    },
    {
        "name":"权限2",
        "description":"描述",
        "data":"权限内容"
    }
]
```

**返回结果示例**

```json
{
    "res_code":0,
    "res_msg":"ok",
    "data":[35,36] //[resource_id]
}
```

### 修改resource

**请求方式以及接口地址**  
`PUT /api/resources?client_id=1`

**请求权限**  
  scope:    >=  resource:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|

**body参数**

```json
{
    "id": 1,
    "name":"权限1",
    "description":"描述",
    "data":"权限内容",
}
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "id": 1,
        "name": "权限1",
        "description": "描述",
        "client_id": 1,
        "data": "权限内容",
        "created_by": "user2",
        "updated_by": "user2",
        "created": "2018-08-24T11:47:35+08:00",
        "updated": "2018-10-26T18:47:45.747125354+08:00"
    }
}
```

### 删除resource（批量）

**请求方式以及接口地址**  
  `DELETE /api/resources/:ids?client_id=1`

**请求权限**  
  scope:    >=  resource:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|
ids|待删除的resourceIds|string|是|1,2,3|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "del_resource_num": 2,
        "del_role_resource_num": 2
    }
}
```

## Role

### 查询role by client
  
查询该client下全部role

**请求方式以及接口地址**  
  `GET /api/roles?client_id=1&relate_user=false&relate_resource=false&is_tree=false`

**请求权限**  
  scope:    >=  role:read

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|
relate_user|查询关联的user|否|bool|false|默认为false，如不需要不传该参数即可
relate_resource|查询关联的resource|否|bool|false|默认为false，如不需要不传该参数即可
is_tree|返回树|否|bool|false|默认为false，如不需要不传该参数即可

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": 2,
            "name": "client name",
            "description": "测试",
            "parent_id": 1,
            "created_by": "user2",
            "updated_by": "user2",
            "created": "2018-10-29T11:08:30+08:00",
            "updated": "2018-10-29T17:23:18+08:00",
            "resources": [ //relate_resource=true并且resources长度大于0，有该字段
                {
                    "role_id": 2,
                    "resource": {
                        "id": 14,
                        "name": "测试14",
                        "description": "描述",
                        "client_id": 1,
                        "data": "",
                        "created_by": "user2",
                        "updated_by": "user2",
                        "created": "2018-08-24T11:47:40+08:00",
                        "updated": "2018-10-26T18:45:39+08:00"
                    }
                }
            ],
            "users": [//relate_user=true并且users长度大于0，有该字段
                {
                    "user": {
                        "id": "test",
                        "fullname": "test",
                        "dn": "xxx",
                        "created_at": "2018-10-25T16:43:27+08:00",
                        "updated_at": "2018-10-25T16:43:27+08:00"
                    },
                    "role_type": "normal"
                }
            ]
        }
    ]
}
```

### 查询role by client and user

查询该某个user在client下直接关联的role或全部role（根据角色树层次遍历获得的）

**请求方式以及接口地址**  
  `GET /api/userRoles?client_id=1&role_type=normal&is_all=false&relate_user=false&relate_resource=false&is_tree=false&is_route=false&user_id=user2`

**请求权限**  
  scope:    >=  role:read

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|
role_type|角色类型|否|string|admin或normal或super|normal：查询用户是normal或admin或super的角色，admin：查询用户是admin或super的角色，super查询用户是super的角色
is_all|是否查询用户全部角色|否|bool|false|默认为false，只查询用户直接关联的角色
relate_user|查询关联的user|否|bool|false|默认为false
relate_resource|查询关联的resource|否|bool|false|默认为false
is_tree|返回树|否|bool|false|默认为false，is_all && is_tree 为true 才能返回树
is_route|返回路径|否|bool|false|默认为false,is_all && is_route为true可以返回路径
user_id|查询的user_id|否|bool|user2|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": 2,
            "name": "指旺",
            "description": "测试",
            "parent_id": 1,
            "created_by": "user2",
            "updated_by": "user2",
            "created": "2018-10-29T11:08:30+08:00",
            "updated": "2018-10-29T17:23:18+08:00",
            "role_type": "normal", //与接口4.1相比 多了角色类型字段
            "resources": [//relate_resource=true并且resources长度大于0，有该字段
                {
                    "role_id": 2,
                    "resource": {
                        "id": 14,
                        "name": "测试14",
                        "description": "描述",
                        "client_id": 1,
                        "data": "dag-create",
                        "created_by": "user2",
                        "updated_by": "user2",
                        "created": "2018-08-24T11:47:40+08:00",
                        "updated": "2018-10-26T18:45:39+08:00"
                    }
                }
            ],
            "users": [//relate_user=true并且users长度大于0，有该字段
                {
                    "user": {
                        "id": "user2",
                        "fullname": "用户名2",
                        "dn": "xxx",
                        "created_at": "2018-10-25T16:43:27+08:00",
                        "updated_at": "2018-10-25T16:43:27+08:00"
                    },
                    "role_type": "normal"
                }
            ]
        }
    ]
}
```

**is_all && is_tree && is_route == true**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "id": 18,
            "name": "rootRole",
            "description": "",
            "parent_id": -1,
            "created_by": "user2",
            "updated_by": "user2",
            "created": "2018-10-29T11:08:30+08:00",
            "updated": "2018-10-29T17:23:19+08:00",//用户没有的角色无role_type字段，is_all && is_route == true 会返回用户有的角色子树到根节点的路径
            "children": [ //树形结构
                {
                    "id": 61,
                    "name": "角色A",
                    "description": "描述",
                    "parent_id": 18,
                    "created_by": "user2",
                    "updated_by": "user2",
                    "created": "2018-11-16T10:59:01+08:00",
                    "updated": "2018-11-16T10:59:01+08:00", //用户没有的角色无role_type字段，is_all && is_route == true 会返回用户有的角色子树到根节点的路径
                    "children": [
                        {
                            "id": 63,
                            "name": "角色C",
                            "description": "描述",
                            "parent_id": 61,
                            "created_by": "user2",
                            "updated_by": "user2",
                            "created": "2018-11-16T10:59:24+08:00",
                            "updated": "2018-11-16T10:59:24+08:00",
                            "role_type": "normal" //用户有的角色显示role_type
                        }
                    ]
                }
            ]
        }
    ]
}
```

### 新增sub role

1、只允许添加子角色操作，client的rootRole是在创建角色的时候自动生成的
2、已经关联权限的角色，不可成为父角色

**请求方式以及接口地址**  
  `POST /api/roles?client_id=1`

**请求权限**  
  scope:    >=  role:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|

**body参数**

```json
{
    "name": "新增子角色名称",
    "description": "描述",
    "parent_id": 19
}
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 47 //新增成功的子角色的id
}
```

### 更新role

**请求方式以及接口地址**  
  `PUT /api/roles?client_id=1`

**请求权限**  
  scope:    >=  role:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|

**body参数**

```json
{
    "id": 12,
    "name": "修改后的子角色名称",
    "description": "修改后的描述",
}
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "id": 1,
        "name": "修改后的子角色名称",
        "description": "修改后的描述",
        "parent_id": 19,
        "created_by": "user2",
        "updated_by": "user2",
        "created": "2018-10-22T16:14:42+08:00",
        "updated": "2018-10-29T17:50:10.103286888+08:00"
    }
}
```

### 删除role

1、rootRole不可删除
2、会将角色的子角色的父角色设置为该角色的父角色（ parent <- role <- children  修改后=> parent <- children）
3、自动断开role_resource关联，role_user关联

**请求方式以及接口地址**  
  `DELETE /api/roles/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|12|
client_id|client的id|否|int|1|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "del_role_num": 1,
        "del_role_resource_num": 2,
        "del_role_user_num": 0
    }
}
```

### 父子角色之间插入role

可以在父子角色之前插入一个新角色

A -> B  
 \-> C  

在A与BC之前插入D

A -D -> B  
    \-> C  

**请求方式以及接口地址**  
  `POST /api/roles/:children_ids?client_id=1`

**请求权限**  
  scope:    >=  role:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
client_id|client的id|否|int|1|
children_ids|子角色的一组id|是|string|1,2,3|

**body参数**

```json
{
    "name": "插入的子角色名称",
    "description": "描述",
    "parent_id": 19
}
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 47 //新增成功的子角色的id
}
```

## Role User

### 查询role中的user

**请求方式以及接口地址**  
  `GET /api/roleUsers/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-user:read

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|否|int|1|不传则查询该client下全部role_user关系，但同时需要rootRoleAdmin权限
client_id|client的id|否|int|1|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "role_id": 1,
            "user_id": "xiaoniu2",
            "role_type": "normal"
        },
        {
            "role_id": 1,
            "user_id": "user2",
            "role_type": "admin"
        }
    ]
}
```

### 添加user到role（批量）

**请求方式以及接口地址**  
  `POST /api/roleUsers/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-user:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|1|
client_id|client的id|否|int|1|

**body参数**

```json
[
    {
        "user_id": "user2",
        "role_type": "admin"
    },
    {
        "user_id": "user1",
        "role_type": "normal"
    }
]
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 3 //添加成功的人数
}
```

### 修改user在role中的类型

**请求方式以及接口地址**  
  `PUT /api/roleUsers/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-user:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|1|
client_id|client的id|否|int|1|

**body参数**

```json
{
    "user_id": "user2",
    "role_type": "normal"
}
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": {
        "role_id": 1,
        "user_id": "user2",
        "role_type": "normal"
    }
}
```

### 从role中删除user（批量）

**请求方式以及接口地址**  
  `DELETE /api/roleUsers/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-user:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|1|
client_id|client的id|否|int|1|

**body参数**

```json
[
    "xiaoniu2",
    "user2"
]
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 2 //从角色中删除的人数
}
```

## Role Resource

### 查询role关联的resource

**请求方式以及接口地址**  
  `GET /api/roleResources/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-resource:read

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|否|int|3|不传则查询该client下全部role_resource关系，但同时需要rootRoleAdmin的权限
client_id|client的id|否|int|1|

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": [
        {
            "role_id": 3,
            "resource_id": 15
        },
        {
            "role_id": 3,
            "resource_id": 16
        },
        {
            "role_id": 3,
            "resource_id": 32
        }
    ]
}
```

### 增加role-resource关联（批量）

**请求方式以及接口地址**  
  `POST /api/roleResources/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-resource:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|3|
client_id|client的id|否|int|1|

**body参数**

```json
[
    14,//resource_id
    13,
    15,
    16,
    32
]
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 5   //该角色新增的role_resource关联数目（增量）
}
```

### 修改role-resource关联（批量）

**请求方式以及接口地址**  
  `PUT /api/roleResources/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-resource:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|3|
client_id|client的id|否|int|1|

**body参数**

```json
[
    14,//resource_id
    13,
    15,
    16,
    32
]
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 5   //该角色修改后关联的角色数目（全量）
}
```

### 删除role—resource关联（批量）

**请求方式以及接口地址**  
  `DELETE /api/roleResources/:role_id?client_id=1`

**请求权限**  
  scope:    >=  role-resource:write

**uri params**

字段名 |变量名 | 是否必填 | 类型 | 示例 | 描述 |
---| --- | --- | --- | --- | --- |
role_id|role的id|是|int|3|
client_id|client的id|否|int|1|

**body参数**

```json
[
    14,//resource_id
    13,
    15,
    16,
    32
]
```

**返回结果示例**

```json
{
    "res_code": 0,
    "res_msg": "ok",
    "data": 5   //该角色成功删除的role_resource的数目
}
```

## Group

### Group Management

Path: `/api/group`

#### POST

This method will create group

Body:

```json
{
    "name": "groupname",
    "description": "group description"
}
```

#### GET

This method return group infomation by name

params: name
The request url likes `/api/group?name=groupname`

#### PUT

This method could change group info.

Body:

```json
{
    "name": "groupname",
    "description": "group description"
}
```

#### DELETE

This method could delete the whole group when group has no related user.
parans: name
The request url likes `/api/group?name=group`

### Group User Management

Path: `/api/groupuser`

#### POST

This method will create the relation about group and user.
Body:

```json
{
    "group": "groupname",
    "user": "userid",
}
```

#### DELETE

This method will delete the relation about group and user.

Body:

```json
{
    "group": "groupname",
    "user": "userid"
}
```

#### GET Group By User

Path: '/api/groupuser/group', method: GET

This method will get the groups by user id

The request uri likes: `/api/groupuser/group?user={your id}`
return json object likes:

```json
[
    {
        "name": "group name",
        "description": "decription"
    }
]
```

#### GET User By Group

Path: '/api/groupuser/user'

This method will get the related user by group name
The request uri likes: `/api/groupuser/user?group={groupname}`

return json object likes:

```json
[
    {
       "fullname": "name",
       "email": "email",
       "...": "other user things"
    }
]
```
