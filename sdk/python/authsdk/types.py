# encoding=utf-8

# 接口获取的完整resource
class ApiResource :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # 资源id
        self.name = data.get('name') # 资源名
        self.description = data.get('description') # 资源描述
        self.client_id = data.get('client_id') # ClientId
        self.data = data.get('data') # 资源内容
        self.created_by = data.get('created_by') # 创建者
        self.updated_by = data.get('updated_by') # 更新者
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间


# 用户所在的Client
class UserClient(object):
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id')
        self.fullname = data.get('fullname')
        self.roles = data.get('roles')


# # 通过ClientId所查询到的Client
class Client :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # clientId
        self.fullname = data.get('fullname') # client全名
        self.secret = data.get('secret') # ClientSecret
        self.redirect_uri = data.get('redirect_uri') # 重定向uri
        self.user_id = data.get('user_id') # 用户Id
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间

# 角色
class Role :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # 角色id
        self.name = data.get('name') # 角色名
        self.description = data.get('description') # 角色类型
        self.parent_id = data.get('parent_id') # 父角色id
        self.created_by = data.get('create_by') # 创建者
        self.updated_by = data.get('update_by') # 更新者
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间
        self.resources = data.get('resources') # 角色相关资源
        self.users = data.get('users') # 角色相关用户

# 角色树
class RoleTree :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # 角色id
        self.name = data.get('name') # 角色名
        self.description = data.get('description') # 角色类型
        self.parent_id = data.get('parent_id ') # 父角色id
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间
        self.resources = data.get('resources') # 角色相关资源
        self.users = data.get('users') # 角色相关用户
        self.children = data.get('children') # 拥有该角色的用户


# 用户角色
class UserRole :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # 角色id
        self.name = data.get('name') # 角色名
        self.description = data.get('description') # 角色类型
        self.parent_id = data.get('parent_id') # 父角色id
        self.created_by = data.get('created_by') # 创建者
        self.updated_by = data.get('updated_by') # 更新者
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间
        self.role_type = data.get('role_type') # 角色类型
        self.resources = data.get('resources') # 角色相关资源
        self.users = data.get('users') # 角色相关用户
        self.children = data.get('children') # 拥有该角色的用户


# 用户角色树
class UserRoleTree :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # 角色id
        self.name = data.get('name') # 角色名
        self.description = data.get('description') # 角色类型
        self.parent_id = data.get('parent_id') # 父角色id
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间
        self.role_type = data.get('role_type') # 角色类型
        self.resources = data.get('resources') # 角色相关资源
        self.users = data.get('users') # 角色相关用户
        self.children = data.get('children') # 拥有该角色的用户


# 角色相关用户
class UserOfRole :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id') # 用户id
        self.fullname = data.get('fullname') # 用户全名
        self.dn = data.get('dn') # dn
        self.created = data.get('created') # 创建时间
        self.updated = data.get('updated') # 更新时间


# 角色中的用户
class RoleUser :
    def __init__(self, data):
        self.__data__ = data
        self.role_id = data.get('role_id') # 角色Id
        self.user_id = data.get('user_id') # 用户Id
        self.role_type = data.get('role_type') # 角色类型


# 角色资源关联信息
class RelatedInfo :
    def __init__(self, data):
        self.__data__ = data
        self.role_id = data.get('role_id') # 角色id
        self.resource_id = data.get('resource_id') # 角色关联资源id


# user
class User :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id')       
        self.fullname = data.get('fullname') 
        self.email = data.get('email')    
        self.wechat = data.get('wechat')   
        self.phone = data.get('phone')    
        self.type = data.get('type')     
        self.dn = data.get('dn') 
        self.resources = data.get('resources')   
        self.resource_map = data.get('resources_map') 
        self.cache_time = data.get('cache_time')   
	
        self.token = data.get('token')    
        self.updated_at = data.get('updated_at') 


class Resource :
    def __init__(self, data):
        self.__data__ = data
        self.id = data.get('id')   
        self.description = data.get('description') 
        self.data = data.get('data') 


class Token :
    def __init__(self, data):
        self.__data__ = data
        self.access_token = data.get('access_token')      
        self.expires_in = data.get('expires_in') 
        self.refresh_token = data.get('refresh_token')     
        self.scope = data.get('scope')     
        self.token_type = data.get('token_type') 
        self.error = data.get('error')     
        self.error_description = data.get('error_description') 


class Group :
    def __init__(self, data):
        self.__data__ = data
        self.name = data.get('name') 
        self.email = data.get('email')
        self.description = data.get('description') 


class GroupUser :
    def __init__(self, data):
        self.__data__ = data
        self.group = data.get('group') 
        self.user = data.get('user')  
