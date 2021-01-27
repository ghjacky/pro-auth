# encoding=utf-8
from __future__ import unicode_literals

import json
import requests
from authsdk import api
from authsdk import types
import sys
import jwt
import time
import random

if sys.version>'3':
    import urllib.parse as parse	
else:
    import urllib as parse



Authorization = "Authorization"

class Config(object):
    def __init__(self, client_id, secret, redirect_uri, host="ac.xxxxx.cn", scope="all:all"):
        self.client_id = client_id
        self.client_secret = secret
        self.redirect_uri = redirect_uri
        self.host = host
        self.scope = scope

class Oauth(object):
    def __init__(self, client_id, secret, redirect_uri, host="ac.xxxxx.cn", scope="all:all"):
        self.config = Config(
            client_id=client_id,
            secret=secret,
            redirect_uri=redirect_uri,
            host=host,
            scope=scope,
        )
    
    def getURL(self, path):
	    return self.config.host + path
        
    def Login(self, code):
        token = self.queryTokenFromOauth2(code)
        if not token.__data__:
            raise Exception("Request token from oauth2 failed! no any data")
        if token.error:
            raise Exception("Request token from oauth2 failed! error: {error}, description: {desc}".format(error=token.error, desc=token.error_description))
        user = self.getUserByToken(token)
        user.token = token
        return user

    def LoginBySecret(self, username, secret):
        token = self.queryTokenFromOauth2BySecret(username, secret)
        if not token.__data__:
            raise Exception("Request token from oauth2 failed! no any data")
        if token.error:
            raise Exception("Request token from oauth2 failed! error: {error}, description: {desc}".format(error=token.error, desc=token.error_description))
        user = self.getUserByToken(token)
        user.token = token
        return user

    def queryTokenFromOauth2(self, code):
        resp = requests.post(self.getURL("/oauth2/token"), params={
            "client_id": self.config.client_id,
            "redirect_uri": self.config.redirect_uri,
            "grant_type": "authorization_code",
            "code": code,
        }, headers={
            Authorization: api.GenerateJWTToken(self.config.client_id, self.config.client_secret),
        })
        data = resp.json()
        return types.Token(data)

    def queryTokenFromOauth2BySecret(self, username, secret):
        token = jwt.encode({
            'username': username,
            'time': str(int(time.time())),
            'nonce': str(random.random()),
        }, secret, algorithm='HS256')

        resp = requests.post(self.getURL("/oauth2/token"), params={
            "client_id": self.config.client_id,
            "redirect_uri": self.config.redirect_uri,
            "grant_type": "password",
            "scope": "all:all",
            "username": token,
        }, headers={
            Authorization: api.GenerateJWTToken(self.config.client_id, self.config.client_secret),
        })
        data = resp.json()
        return types.Token(data)
    
    def generateOauthToken(self, token):
	    return " ".join([token.token_type, token.access_token])

    def getUserByToken(self, token):
        resp = requests.get(self.getURL("/api/user"), headers={
            Authorization: self.generateOauthToken(token)
        })
        
        if resp.status_code != requests.status_codes.codes.OK:
            raise Exception("Request user failed! code: {code}".format(code=resp.status_code))
            
        respData = resp.json()
        if respData.get("res_code") != 0:
            raise Exception("request:apiuser response is invalid! res_code: {code}, res_msg: {msg}".format(
                code=respData.get("res_code"),
                msg=respData.get("res_msg")
            ))

        return types.User(respData.get("data", {}))

    def Logout(self, token):
        resp = requests.delete(self.getURL("/oauth2/token"),
            params={"access_token": token.access_token},
            headers={Authorization: self.generateOauthToken(token)}
        )
        if resp.status_code != requests.status_codes.codes.OK:
            raise Exception("Logout failed! code: {code}".format(code=resp.status_code))
    
    def GenerateLoginURL(self, state):
        params = parse.urlencode({
            "client_id": self.config.client_id,
            "redirect_uri": self.config.redirect_uri,
            "response_type": "code",
            "state": state,
            "scope": self.config.scope,
        })
        return self.getURL("/oauth2/authorize") + "?" + params

    def LoadUserResource(self, user):
        resp = requests.get(self.getURL("/api/userResources"), headers={
            Authorization: self.generateOauthToken(user.token),
        })

        if resp.status_code != requests.status_codes.codes.OK:
            raise Exception("Load user resources failed! code: {code}".format(code=resp.status_code))
            
        respData = resp.json()
        if respData.get("res_code") != 0:
            raise Exception("request: api/userresources response is invalid! res_code: {code}, res_msg: {msg}".format(
                code=respData.get("res_code"),
                msg=respData.get("res_msg")
            ))
        
        return types.Resource(respData.get("data", {}))
