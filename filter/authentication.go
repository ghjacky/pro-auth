package filter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.xxxxx.cn/platform/auth/consts"
	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sagacioushugo/oauth2"
)

type AuthInfo struct {
	ClientId  int64
	UserId    string
	LoginType int64 // 0:no auth, 1:token, 2:secret, 3:session
	Scope     string
}

const (
	TOKEN_TYPE_BEARER = "Bearer"
	TOKEN_TYPE_CLIENT = "Client"
	TOKEN_TYPE_BASIC  = "Basic"

	HEADER_AUTHORIZATION = "Authorization"

	CLAIMS_TIME = "time"
	CLAIMS_ID   = "id"

	CLIENT_TOKEN_TIMEOUT = 60 * 5
)

func CheckAuthFilter(ctx *context.Context) {
	authorizationString := ctx.Input.Header(HEADER_AUTHORIZATION)
	tokenInfo := strings.Split(authorizationString, " ")
	tokenType := TOKEN_TYPE_BEARER
	token := ""

	if len(tokenInfo) >= 2 {
		tokenType = tokenInfo[0]
		token = tokenInfo[1]
	} else if len(tokenInfo) == 1 && tokenInfo[0] != "" {
		token = tokenInfo[0]
	} else if user, ok := ctx.Input.CruSession.Get(consts.SESSION_USER_KEY).(db.User); ok {
		logs.Info("session user : %+v", user)
		authInfo := AuthInfo{}
		authInfo.LoginType = consts.AUTH_BY_SESSION
		authInfo.UserId = user.Id
		authInfo.Scope = "all:all"
		ctx.Input.SetData(consts.AUTH_KEY, authInfo)
		return
	} else {
		logs.Debug("error: no session user and no token info")
		errors.ErrorResponse(errors.ErrNoAuth, ctx)
		return
	}

	switch tokenType {
	case TOKEN_TYPE_BEARER:
		bearerTokenHandle(token, ctx)
	case TOKEN_TYPE_CLIENT:
		clientTokenHandle(token, ctx)
	default:
		errors.ErrorResponse(errors.ErrNoAuth, ctx)
	}
}

func bearerTokenHandle(token string, ctx *context.Context) {
	store := db.TokenStore{}
	if t, err := store.GetByAccess(token); err != nil {
		errors.ErrorResponse(err, ctx)
	} else if t.IsAccessExpired() {
		errors.ErrorResponse(oauth2.ErrExpiredAccessToken, ctx)
	} else {
		logs.Info("token user : %s", t.GetUserId())
		authInfo := AuthInfo{}
		authInfo.LoginType = consts.AUTH_BY_TOKEN
		authInfo.UserId = t.GetUserId()
		authInfo.Scope = t.GetScope()
		id, _ := strconv.ParseInt(t.GetClientId(), 10, 64)
		authInfo.ClientId = id
		ctx.Input.SetData(consts.AUTH_KEY, authInfo)
	}
}

func clientTokenHandle(tokenString string, ctx *context.Context) {
	client, err := ParseJWTToken(tokenString)
	if err != nil {
		logs.Error("Parse Failed, token: %s", tokenString)
		errors.ErrorResponse(fmt.Errorf("Parse Token Failed, error: %v", err), ctx)
	} else {
		logs.Info("client : %d", client.Id)
		authInfo := AuthInfo{}
		authInfo.LoginType = consts.AUTH_BY_SECRET
		authInfo.UserId = fmt.Sprintf("%s-%d", consts.AUTH_BY_SECRET_USER_Id, client.Id)
		authInfo.Scope = "all:all"
		authInfo.ClientId = client.Id
		ctx.Input.SetData(consts.AUTH_KEY, authInfo)
	}
}

func ParseJWTToken(token string) (*db.Client, error) {
	var client *db.Client
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		var err error

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("Unexpected claims type")
		}

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", claims[CLAIMS_ID])
		}

		timeStampString, ok := claims[CLAIMS_TIME].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected timeStamp format, it should be string")
		}
		timeStamp, err := strconv.Atoi(timeStampString)
		if err != nil {
			return nil, fmt.Errorf("Unexpected timeStamp length, it should be int64")
		}

		if time.Now().Unix()-int64(timeStamp) > 300 {
			return nil, fmt.Errorf("Unexpected timeStamp in claims: %d", timeStamp)
		}

		clientIDString, ok := claims[CLAIMS_ID].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected client id: %v", clientIDString)
		}
		clientID, err := strconv.Atoi(clientIDString)
		if err != nil {
			return nil, fmt.Errorf("Unexpected timeStamp length, it should be int64")
		}

		client, err = db.GetClient(int64(clientID))
		if err != nil {
			return nil, err
		}

		return []byte(client.Secret), nil
	})

	return client, err
}
