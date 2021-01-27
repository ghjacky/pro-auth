package authcustome

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	authsdk "code.xxxxx.cn/platform/auth/sdk/golang"
	"github.com/dgrijalva/jwt-go"
)

const (
	AUTHORIZATION = "Authorization"
)

type Oauth struct {
	config *OauthConfig
	client *http.Client
}

type OauthConfig struct {
	ClientId     int64  `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUri  string `json:"redirect_uri"`
	Host         string `json:"host"`
	Scope        string `json:"scope"`
}

func NewAuthService(config *OauthConfig) *Oauth {
	if config == nil {
		panic("auth service init failed: config counld not be nil")
	}

	if config.Scope == "" {
		config.Scope = "all:all"
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	auth := &Oauth{
		config,
		client,
	}

	return auth
}

func (a *Oauth) doRequest(req *http.Request) ([]byte, error) {
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Response Status Not Right: %v", resp.StatusCode)
	}

	if resp.Body == nil {
		return []byte{}, nil
	}

	return ioutil.ReadAll(resp.Body)
}

func (a *Oauth) getURL(path string) string {
	return a.config.Host + path
}

func (a *Oauth) request(method, path, token string, params map[string]string) ([]byte, error) {
	urlParams := url.Values{}
	for k, v := range params {
		urlParams.Add(k, v)
	}

	url := a.getURL(path + "?" + urlParams.Encode())

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(AUTHORIZATION, token)

	data, err := a.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("Do request error: %v", err)
	}

	return data, nil
}

// 用code向sso请求获取access token
func (a *Oauth) queryTokenFromOauth2(code string) (*authsdk.Token, error) {
	data, err := a.request(
		"POST",
		"/oauth2/token",
		authsdk.GenerateJWTToken(a.config.ClientId, a.config.ClientSecret),
		map[string]string{
			"client_id":    strconv.FormatInt(a.config.ClientId, 10),
			"redirect_uri": a.config.RedirectUri,
			"grant_type":   "authorization_code",
			"code":         code,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("request err:%v ", err)
	}

	token := authsdk.Token{}
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("to Json Failed! error: %v", err)
	}

	if token.Error != "" {
		return nil, errors.New(token.Error + ":" + token.ErrorDescription)
	}

	return &token, nil
}

// 用secret向sso请求获取access token
func (a *Oauth) queryTokenFromOauth2BySecret(username, secret string) (*authsdk.Token, error) {
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"time":     fmt.Sprintf("%d", time.Now().Unix()),
		"username": username,
		"nonce":    fmt.Sprintf("%d", rand.Int31()),
	}).SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	data, err := a.request(
		"POST",
		"/oauth2/token?",
		authsdk.GenerateJWTToken(a.config.ClientId, a.config.ClientSecret),
		map[string]string{
			"client_id":  strconv.FormatInt(a.config.ClientId, 10),
			"grant_type": "password",
			"scope":      "all:all",
			"username":   tokenString,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Do request error: %v", err)
	}

	token := authsdk.Token{}
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("to Json Failed! error: %v", err)
	}

	if token.Error != "" {
		return nil, errors.New(token.Error + ":" + token.ErrorDescription)
	}

	return &token, nil
}

// LoginURL 生成重定向url
func (a *Oauth) LoginURL(state string) string {
	params := url.Values{}

	params.Add("client_id", strconv.FormatInt(a.config.ClientId, 10))
	params.Add("redirect_uri", a.config.RedirectUri)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", a.config.Scope)

	return a.getURL("/oauth2/authorize") + "?" + params.Encode()
}

// Login 登录（用于前后端分离的项目，sso登录授权后重定向到指定地址将参数中的code传给后端获取用户）
func (a *Oauth) Login(code string) (*authsdk.User, error) {
	token, err := a.queryTokenFromOauth2(code)
	if err != nil {
		return nil, fmt.Errorf("Query Token Failed! %v", err)
	}

	user, err := a.getUserByToken(token)
	if err != nil {
		return nil, fmt.Errorf("Get user by Token Failed! %v", err)
	}
	user.Token = *token

	return user, nil
}

// LoginBySecret 登录用于cli等场景，需要用户提供在auth界面上获取到的secret
func (a *Oauth) LoginBySecret(username, secret string) (*authsdk.User, error) {
	token, err := a.queryTokenFromOauth2BySecret(username, secret)
	if err != nil {
		return nil, fmt.Errorf("Query Token Failed! %v", err)
	}

	user, err := a.getUserByToken(token)
	if err != nil {
		return nil, fmt.Errorf("Get user by Token Failed! %v", err)
	}
	user.Token = *token

	return user, nil
}

// Logout 登出
func (a *Oauth) Logout(token *authsdk.Token) error {
	_, err := a.request("DELETE", "/oauth2/token", a.generateOauthToken(token), map[string]string{"access_token": token.AccessToken})
	return err
}

func (a *Oauth) generateOauthToken(token *authsdk.Token) string {
	return fmt.Sprintf("%s %s", token.TokenType, token.AccessToken)
}

func (a *Oauth) getUserByToken(token *authsdk.Token) (*authsdk.User, error) {
	data, err := a.request("GET", "/api/user", a.generateOauthToken(token), map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("Do Request Err: %v", err)
	}

	res := authsdk.RespBody{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("to Json Failed! error: %v", err)
	}

	if res.ResCode != 0 {
		return nil, fmt.Errorf("code wrong: code %v, msg: %v", res.ResCode, res.ResMsg)
	}

	user := authsdk.User{}
	userJSON, _ := json.Marshal(res.Data)
	if err := json.Unmarshal(userJSON, &user); err != nil {
		return nil, fmt.Errorf("res data error : %+v", res.Data)
	}

	return &user, nil
}

func (a *Oauth) LoadUserResource(user *authsdk.User) ([]*authsdk.Resource, error) {
	data, err := a.request("GET", "/api/userResources", a.generateOauthToken(&user.Token), map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("Do Request err: %v", err)
	}

	res := authsdk.RespBody{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	if res.ResCode != 0 {
		return nil, errors.New(res.ResMsg)
	}

	resources := []*authsdk.Resource{}
	resourcesJson, _ := json.Marshal(res.Data)
	if err := json.Unmarshal(resourcesJson, &resources); err != nil {
		return nil, fmt.Errorf("res data error : %+v", res.Data)
	}

	return resources, nil
}
