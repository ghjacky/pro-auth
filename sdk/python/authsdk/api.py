# encoding=utf-8
from __future__ import unicode_literals
import requests
import urllib
import jwt
import time
import random
from authsdk import types
import json
import sys

class Resources:
    CLIENT         = "client"
    GROUP          = "group"
    GROUPUSER      = "groupuser"
    RESOURCES      = "resources"
    USER_RESOURCES = "userResources"
    ROLE           = "roles"
    ROLE_TREE      = "roleTree"
    ROLE_RESOURCES = "roleResources"
    USER           = "users"
    USER_ROLES     = "userRoles"
    ROLE_USER      = "roleUsers"
    USER_CLIENT    = "userClients"

class RoleType:
    Normal = "normal"
    Admin  = "admin"
    Super  = "super"

class Config(object):
    def __init__(self, client_id, secret, host, redirecturi):
        self.redirecturi = redirecturi
        self.client_id = client_id
        self.secret = secret
        self.host = host

class API(object):
    def __init__(self, client_id, secret, host, redirecturi=""):
        self.config = Config(
            client_id=client_id,
            secret=secret,
            host=host,
            redirecturi=redirecturi,
        )  

    def generateHeader(self):
        token = GenerateJWTToken(self.config.client_id, self.config.secret)
        if sys.version>'3':
            token = token.decode('utf8')
        return {
            "Authorization": "Client " + token
        }

    def doRequest(self, method, resource, **kwargs):
        url = self.config.host + '/api/' + resource
        kwargs["headers"] = self.generateHeader()

        func = getattr(requests, method)
        resp = func(url, **kwargs)
        if resp.status_code != requests.status_codes.codes.OK:
            raise Exception("request:{url} response is invalid! code: {code} ".format(url=url, code=resp.status_code))
        respData = resp.json()
        if respData.get("res_code") != 0:
            raise Exception("request:{url} response is invalid! res_code: {code}, res_msg: {msg}".format(
                url=url,
                code=respData.get("res_code"),
                msg=respData.get("res_msg")
            ))

        return respData.get("data", {})
    
    def post(self, resource, **kwargs):
        return self.doRequest('post', resource, **kwargs)
    
    def get(self, resource, **kwargs):
        return self.doRequest('get', resource, **kwargs)

    def put(self, resource, **kwargs):
        return self.doRequest('put', resource, **kwargs)

    def delete(self, resource, **kwargs):
        return self.doRequest('delete', resource, **kwargs)


    def GetClient(self):
        data = self.get(Resources.CLIENT)
        if data:
            return types.Client(data)

    def UpdateClient(self, fullname, redirect_uri):
        return self.put(Resources.CLIENT, data=json.dumps({
            "fullname": fullname,
            "redirect_uri": redirect_uri,
        }))

    def GetClientByUser(self, user_id, role_type):
        data = self.get(Resources.USER_CLIENT, params={
            "user_id": user_id,
            "role_type": role_type, 
        }) 
        if data:
            return [types.UserClient(d) for d in data]


    def AddResource(self, resources):
	    return self.post(Resources.RESOURCES, data=json.dumps(resources))

    def GetAllResources(self, user_id=None):
        params={}
        if user_id:
            params['user_id'] = user_id
        data = self.get(Resources.RESOURCES, params=params)
        if data:
            return [types.ApiResource(d) for d in data]

    def UpdateResource(self, rId, rName, rDescription, rData):
        return self.put(Resources.RESOURCES+"/"+str(rId), data=json.dumps({
            "id":          rId,
            "name":        rName,
            "description": rDescription,
            "data":        rData,
        }))

    def DeleteResources(self, resource_ids):
        return self.delete(Resources.RESOURCES + '/' + ','.join(map(str, resource_ids)))


    def AddRole(self, name, description, parentId):
        return self.post(Resources.ROLE, data=json.dumps({
            "name":        name,
            "description": description,
            "parent_id":   parentId,
        }))

    def GetAllRole(self, is_related_resource=False, is_related_user=False, is_tree=True):
        data = self.get(Resources.ROLE, params={
        "is_tree": is_tree,
        "relate_user": is_related_user,
        "relate_resource": is_related_resource,
        })
        if data: 
            return [types.RoleTree(d) for d in data]

    def GetRoleByIDs(self, ids):
        data = self.get(Resources.ROLE + "/batch", params={
        "role_ids": ",".join(map(str, ids)),
        })
        if data: 
            return [types.Role(d) for d in data]

    def SearchRoles(self, roleName, is_related_resource=False, is_related_user=False):
        data = self.get("roles/search", params={
        "role_name": roleName,
        "relate_user": is_related_user,
        "relate_resource": is_related_resource,
        })
        if data: 
            return [types.RoleTree(d) for d in data]

    def GetRoleTreeByID(self, user_id, is_relate_children=False, is_relate_resource=False, is_relate_user=False, is_tree=False):
        data = self.get(Resources.ROLE_TREE, params={
        "is_tree": is_tree,
        "user_id": user_id,
        "relate_children": is_relate_children,
        "relate_user": is_relate_user,
        "relate_resource": is_relate_resource,
        })
        if data: 
            return types.RoleTree(data[0])

    def UpdateRole(self, role_id, name, description, parent_id):
        return self.put(Resources.ROLE, data=json.dumps({
            "id":          role_id,
            "name":        name,
            "description": description,
            "parent_id":   parent_id,
        }))

    def DeleteRole(self, role_id):
	    return self.delete(Resources.ROLE + '/' + str(role_id))


    def AddUserToRole(self, role_id, user_ids, role_type="admin"):
        info = [{"user_id": uid, "role_type": role_type} for uid in user_ids]
        print(json.dumps(info))
        return self.post(Resources.ROLE_USER+ "/" + str(role_id), data=json.dumps(info))

    def GetUserRole(self, user_id, is_related_resource, is_relate_user, is_all=True, is_tree=True):
        data = self.get(Resources.USER_ROLES, params={
            "is_tree": is_tree,
            "is_all": is_all,
            "user_id": user_id,
            "relate_user": is_relate_user,
            "relate_resource": is_related_resource,
        })
        if data:
            return [types.UserRole(d) for d in data]

    def GetRoleUserByUserIDs(self, user_ids):
        data = self.get(Resources.USER_ROLES + "/users", params={
            "user_ids": ",".join(map(str, user_ids)),
        })
        if data:
            return [types.RoleUser(d) for d in data]

    def GetUsersOfRole(self, role_id):
        data = self.get(Resources.ROLE_USER, params={"role_id": role_id})
        if data:
            return [types.RoleUser(d) for d in data]

    def UpdateUserOfRole(self, role_id, user_ids, role_type="admin"):
        info = [{"user_id": uid, "role_type": role_type} for uid in user_ids]
        return self.put(Resources.ROLE_USER+ "/" + str(role_id), data=json.dumps(info))

    def DeleteUserFromRole(self, role_id, user_ids): 
        return self.delete(Resources.ROLE_USER+'/'+str(role_id), data=json.dumps(user_ids))


    def AddRoleResourceRelations(self, role_id, res_ids): 
	    return self.post(Resources.ROLE_RESOURCES+ '/' + str(role_id), data=json.dumps(res_ids))

    def GetRoleResourceRelatedInfo(self, role_id=None):
        suffix = ''
        if role_id:
            suffix = '/' + str(role_id)
        data = self.get(Resources.ROLE_RESOURCES + suffix)
        if data:
            return [types.RelatedInfo(d) for d in data]

    def GetResourceByRole(self, roleID):
        relations = self.GetRoleResourceRelatedInfo(roleID)
        if not relations:
            return
        data = self.get(Resources.RESOURCES + "/list", params={
            "id": [str(relation.resource_id) for relation in relations]
        })
        if data:
            return [types.ApiResource(d) for d in data] 

    def GetUserResources(self, user_id):
        data = self.get(Resources.RESOURCES, params={"user_id": user_id})
        if data:
            return [types.ApiResource(d) for d in data]

    def UpdateRoleResourceRelations(self, role_id, res_ids): 
	    return self.put(Resources.ROLE_RESOURCES+ '/' + str(role_id), data=json.dumps(res_ids))

    def DeleteRoleResourceRelations(self, role_id, res_ids): 
	    return self.delete(Resources.ROLE_RESOURCES+'/' + str(role_id), data=json.dumps(res_ids))


    def CheckThirdToken(self, access_token, role_id, client_id):
        return self.get("/user/check", params={
            "access_token": access_token,
            "role_id": role_id,
            "client_id": client_id,
        })


    def CreateGroup(self, name, description, email=""):
	    return self.post(Resources.GROUP, data=json.dumps({
            "name": name,
            "description": description,
            "email": email,
        }))

    def GetGroup(self, name=None):
        params = {}
        if name:
            params['name'] = name
        data = self.get(Resources.GROUP, params=params)
        if data:
            return [types.Group(d) for d in data]

    def UpdateGroup(self, name, description, email=""):
        return self.put(Resources.GROUP, data=json.dumps({
            "name": name,
            "description": description,
            "email": email,
        }))

    def DeleteGroup(self, name):
        return self.delete(Resources.GROUP, params={"name": name})


    def CreateGroupUser(self, group_name, user_id):
        return self.post(Resources.GROUPUSER, data=json.dumps({
            "group": group_name,
            "user": user_id, 
        }))

    def GetGroupsByUserID(self, user_id):
        data = self.get(Resources.GROUPUSER + '/group', params={"user": user_id})
        if data:
            return [types.Group(d) for d in data]

    def GetUsersByGroup(self, name):
        data = self.get(Resources.GROUPUSER + '/user', params={"group": name})
        if data:
            return [types.User(d) for d in data]

    def DeleteGroupUser(self, group_name, user_id):
        return self.delete(Resources.GROUPUSER, data=json.dumps({
            "group": group_name,
            "user": user_id,
        }))


def GenerateJWTToken(client_id, secret):
    return jwt.encode({
        'id': str(client_id),
        'time': str(int(time.time())),
        'nonce': str(random.random()),
    }, secret, algorithm='HS256')