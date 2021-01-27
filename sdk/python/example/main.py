# encoding=utf-8
import sys  
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import argparse
from authsdk import api
from authsdk import oauth

def main():
    parser = argparse.ArgumentParser(description='Process args')
    parser.add_argument('-secret', metavar='secret', dest="secret", type=str, help='Your Client Secret')
    parser.add_argument('-id', metavar='id', dest="id", type=int, help='Your Client Secret')
    parser.add_argument('-host', metavar='host', dest="host", type=str, help='Your Client Secret')
    parser.add_argument('-username', metavar='username', dest="username", type=str, help='Username')
    parser.add_argument('-user-secret', metavar='user-secret', dest="user_secret", type=str, help='User Secret')
    parser.add_argument('-redirect', metavar='redirect', dest="redirect", type=str, help='Your Client Secret')
    args = parser.parse_args()
    
    service = api.API(
        client_id=args.id,
        secret=args.secret,
        host=args.host,
        redirecturi=args.redirect,
    )
    print ("Testing service")
    try:
        testService(service)
    except Exception as e:
        print ("Exception Happend when test service, err: {err}".format(err=e))
    
    print ("\nTesting oauth")
    oauthService = oauth.Oauth(
        client_id=args.id,
        secret=args.secret,
        redirect_uri=args.redirect,
        host=args.host
    )
    try:
        testOauthBySecret(oauthService, args.username, args.user_secret)
        testOauth(oauthService)
    except Exception as e:
        print ("Exception Happend when test service, err: {err}".format(err=e))


def testService(service):
    client = service.GetClient()
    if client:
        print ("Client: {client_id}, {secret}, {redirect}".format(
            client_id=client.id, 
            secret=client.secret,
            redirect=client.redirect_uri
        ))
    else:
        print ("Get client failed")

    roles = service.GetAllRole(is_tree=False)
    if not roles:
        print ("get role failed!")

    if len(roles) > 0:
        treeRole = service.GetRoleTreeByID(roles[0].id)
        print ("Tree Role: {}".format(treeRole))

    role_id = service.AddRole('test_role','test role', roles[0].id)
    if not role_id:
        raise Exception("Add Role Failed!")
    print ("Add Role Success! Role ID: {role_id}".format(role_id=role_id))
    user_name = "admin"
    print(service.AddUserToRole(role_id, [user_name]))
    roleUsers = service.GetUsersOfRole(role_id)
    if not roleUsers or user_name not in [ru.user_id for ru in roleUsers]:
        raise Exception("ADD User to Role Failed!")

    for ru in roleUsers:
        print ("Role: {role_id}, User: {user}".format(
            role_id=ru.role_id,
            user=ru.user_id,
        ))

    res_ids = service.AddResource([{
        "name": "test_resources",
        "decsription": "test_resources",
        "data": "test_resources",
    }])
    if not res_ids:
        raise Exception("add resources failed")
    else:
        print ("Add Resources success! res_ids: {res_ids}".format(res_ids=res_ids))

    resources = service.GetAllResources(user_name)
    if not resources:
        raise Exception("get all resources failed")
    
    for resource in resources:
        print ("Resource: id:{id}, description:{description}, data: {data}".format(
            id=resource.id,
            description=resource.description,
            data=resource.data,
        ))

    print(service.AddRoleResourceRelations(role_id, res_ids))

    resources = service.GetResourceByRole(role_id)
    if not resources:
        raise Exception("ADD Role Resource Relations Failed!")

    print("ADD Role Resource Relations Success!")
    for resource in  resources:
        print ("Role id: {role} Resource: {resource}, Data: {data}\n".format(
            role=role_id,
            resource=resource.name,
            data=resource.data
        ))

    print(service.DeleteRoleResourceRelations(role_id, res_ids))
    print ("Delete role resource Success!")

    print(service.DeleteResources(res_ids))
    print ("Delete Resources Success!")


    print(service.DeleteUserFromRole(role_id, [user_name]))
    print ("Delete User from role Success!")

    print(service.DeleteRole(role_id))
    print ("Delete Role Success!")

    service.CreateGroup("test_group", "test group description")
    groups = service.GetGroup("test_group")
    if groups:
        for group in groups:
            print("Group name:{name}, description: {desc}".format(
                name=group.name,
                desc=group.description,
            ))
    else:
        raise Exception("Group Get Failed!")

    service.UpdateGroup("test_group", "changed test group description")
    groups = service.GetGroup("test_group")
    if groups and groups[0].description == "changed test group description":
        print("Change Group Success!")
    else:
        print("Change Group Failed!")
    
    service.CreateGroupUser("test_group", user_name)
    print ("Create Group user success!")

    groups = service.GetGroupsByUserID(user_name)
    if groups:
        for group in groups:
            print ("name: {name}, desc: {desc}".format(
                name=group.name,
                desc=group.description,
            ))
    else:
        raise Exception("Get Group by user id failed!")
    
    users = service.GetUsersByGroup("test_group")
    if users:
        for user in users:
            print ("user id:{id}".format(id=user.id))
    else:
        raise Exception("GEt user by group failed!")


    service.DeleteGroupUser("test_group", user_name)
    print ("Delete Group user Success!")
    
    service.DeleteGroup("test_group")
    print ("Delete Group Success!")

def testOauth(service):
    print ("LoginURL: {}".format(service.GenerateLoginURL("tmp")))
    # TODO make a server to handle oauth request

def testOauthBySecret(service, username, secret):
    user = service.LoginBySecret(username, secret)
    print ("User ID: {id}, name: {name}, email: {email}".format(id=user.id, name=user.fullname, email=user.email))

if __name__ == "__main__":
    main()