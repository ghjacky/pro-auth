package db

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/sagacioushugo/oauth2"
	"github.com/sagacioushugo/oauth2/manager"
	"github.com/sagacioushugo/oauth2/store"
	"time"
)

type Token struct {
	Id              int64 `orm:"pk;auto"`
	ClientId        string
	UserId          string
	Scope           string
	Access          string    `orm:"null;index"`
	AccessCreateAt  time.Time `orm:"null"`
	AccessExpireIn  int64     `orm:"default(0)"`
	Refresh         string    `orm:"null;index"`
	RefreshCreateAt time.Time `orm:"null"`
	RefreshExpireIn int64     `orm:"default(0)"`
	Code            string    `orm:"null;index"`
	CodeCreateAt    time.Time `orm:"null"`
	CodeExpireIn    int64     `orm:"default(0)"`
	GrantSessionId  string    `orm:"default();index"`
}

func (token *Token) GetClientId() string {
	return token.ClientId
}
func (token *Token) SetClientId(clientId string) {
	token.ClientId = clientId
}
func (token *Token) GetUserId() string {
	return token.UserId
}
func (token *Token) SetUserId(userId string) {
	token.UserId = userId
}
func (token *Token) GetScope() string {
	return token.Scope
}
func (token *Token) SetScope(scope string) {
	token.Scope = scope
}

//code info
func (token *Token) GetCode() string {
	return token.Code
}
func (token *Token) SetCode(code string) {
	token.Code = code
}
func (token *Token) GetCodeCreateAt() time.Time {
	return token.CodeCreateAt
}
func (token *Token) SetCodeCreateAt(codeCreateAt time.Time) {
	token.CodeCreateAt = codeCreateAt
}
func (token *Token) GetCodeExpireIn() int64 {
	return token.CodeExpireIn
}
func (token *Token) SetCodeExpireIn(codeExpireIn int64) {
	token.CodeExpireIn = codeExpireIn
}

//access info
func (token *Token) GetAccess() string {
	return token.Access
}
func (token *Token) SetAccess(access string) {
	token.Access = access
}
func (token *Token) GetAccessCreateAt() time.Time {
	return token.AccessCreateAt
}
func (token *Token) SetAccessCreateAt(accessCreateAt time.Time) {
	token.AccessCreateAt = accessCreateAt
}
func (token *Token) GetAccessExpireIn() int64 {
	return token.AccessExpireIn

}
func (token *Token) SetAccessExpireIn(accessExpireIn int64) {
	token.AccessExpireIn = accessExpireIn
}

// refresh info
func (token *Token) GetRefresh() string {
	return token.Refresh
}
func (token *Token) SetRefresh(refresh string) {
	token.Refresh = refresh
}
func (token *Token) GetRefreshCreateAt() time.Time {
	return token.RefreshCreateAt
}
func (token *Token) SetRefreshCreateAt(refreshCreateAt time.Time) {
	token.RefreshCreateAt = refreshCreateAt
}
func (token *Token) GetRefreshExpireIn() int64 {
	return token.RefreshExpireIn
}
func (token *Token) SetRefreshExpireIn(refreshExpireIn int64) {
	token.RefreshExpireIn = refreshExpireIn
}

func (token *Token) IsCodeExpired() bool {
	if token.CodeExpireIn == 0 {
		return true
	} else {
		ct := time.Now()
		return ct.After(token.CodeCreateAt.Add(time.Second * time.Duration(token.CodeExpireIn)))
	}
}

func (token *Token) IsAccessExpired() bool {
	if token.AccessExpireIn == 0 {
		return true
	} else {
		ct := time.Now()
		return ct.After(token.AccessCreateAt.Add(time.Second * time.Duration(token.AccessExpireIn)))
	}
}

func (token *Token) IsRefreshExpired() bool {
	if token.RefreshExpireIn == 0 {
		return true
	} else {
		ct := time.Now()
		return ct.After(token.RefreshCreateAt.Add(time.Second * time.Duration(token.RefreshExpireIn)))
	}
}

type TokenStore struct {
}

func (s *TokenStore) Init(tokenConfig string) error {
	return nil
}

func (s *TokenStore) NewToken(ctx *context.Context) store.Token {
	token := Token{
		GrantSessionId: ctx.Input.CruSession.SessionID(),
	}
	return &token
}

func (s *TokenStore) Create(token store.Token) error {
	t, _ := token.(*Token)
	o := orm.NewOrm()
	_, err := o.Insert(t)
	return err
}

func (s *TokenStore) GetByAccess(access string) (store.Token, error) {
	o := orm.NewOrm()
	var token Token

	if err := o.QueryTable("token").Filter("access", access).One(&token); err != nil {
		if err == orm.ErrNoRows {
			return nil, oauth2.ErrInvalidAccessToken
		} else {
			return nil, err
		}
	} else {
		return &token, nil
	}
}
func (s *TokenStore) GetByRefresh(refresh string) (store.Token, error) {
	o := orm.NewOrm()
	var token Token

	if err := o.QueryTable("token").Filter("refresh", refresh).One(&token); err != nil {
		if err == orm.ErrNoRows {
			return nil, oauth2.ErrInvalidRefreshToken
		} else {
			return nil, err
		}
	} else {
		return &token, nil
	}

}
func (s *TokenStore) GetByCode(code string) (store.Token, error) {
	o := orm.NewOrm()
	var token Token

	if err := o.QueryTable("token").Filter("code", code).One(&token); err != nil {
		if err == orm.ErrNoRows {
			return nil, oauth2.ErrInvalidAuthorizeCode
		} else {
			return nil, err
		}
	} else {
		return &token, nil
	}
}
func (s *TokenStore) CreateAndDel(tokenNew store.Token, tokenDel store.Token) error {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}
	new, _ := tokenNew.(*Token)
	del, _ := tokenDel.(*Token)

	// code <= token + refresh <= new token + refresh  保存请求code那个sessionId
	new.GrantSessionId = del.GrantSessionId

	if _, err := o.QueryTable("token").Filter("id", del.Id).Delete(); err != nil {
		if rerr := o.Rollback(); rerr != nil {
			logs.Error(rerr)
		}
		return err
	} else if _, err := o.Insert(new); err != nil {
		if rerr := o.Rollback(); rerr != nil {
			logs.Error(rerr)
		}
		return err
	} else {
		if cerr := o.Commit(); cerr != nil {
			logs.Error(cerr)
		}
		return nil
	}
}
func (s *TokenStore) GC(gcInterval int64) {
	timestamp := time.Now().Unix() - gcInterval
	sql := fmt.Sprintf("delete from token where access_create_at is null or unix_timestamp(access_create_at) < %d", timestamp)
	o := orm.NewOrm()

	if Raw, err := o.Raw(sql).Exec(); err != nil {
		logs.Error(fmt.Errorf("token gc failed: %s", err.Error()))
	} else {
		num, _ := Raw.RowsAffected()
		logs.Info("gc token num: %d", num)
	}
}

func init() {
	manager.Register("mysql", &TokenStore{})
}

func TokenClearBySessionId(sessionId string) error {
	o := orm.NewOrm()

	params := orm.Params{}
	params["access_expire_in"] = 0
	params["refresh_expire_in"] = 0
	params["code_expire_in"] = 0

	if num, err := o.QueryTable("token").Filter("grant_session_id", sessionId).Update(params); err != nil {
		return err
	} else {
		logs.Info("clear token num :%d", num)
		return nil
	}

}
