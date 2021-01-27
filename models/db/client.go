package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"code.xxxxx.cn/platform/auth/errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type Client struct {
	Id          int64      `orm:"pk;auto" json:"id"`
	Fullname    string     `orm:"size(128);unique" json:"fullname"`
	Secret      string     `orm:"size(22);index" json:"secret"`
	RedirectUri string     `orm:"size(256)" json:"redirect_uri"`
	UserId      string     `orm:"size(256);index" json:"user_id"`
	CreatedAt   *time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt   *time.Time `orm:"auto_now;auto_now_add" json:"updated_at"`
}

type ClientJson struct {
	Id           int64           `json:"id"`
	Fullname     string          `json:"fullname"`
	Secret       string          `json:"secret,omitempty"`
	RedirectUri  string          `json:"redirect_uri,omitempty"`
	UserId       string          `json:"created_by,omitempty"`
	CreatedAt    *time.Time      `json:"created_at,omitempty"`
	UpdatedAt    *time.Time      `json:"updated_at,omitempty"`
	RootRoleId   int64           `json:"root_role_id,omitempty"`
	RootRoleName string          `json:"root_role_name,omitempty"`
	Roles        []*RoleJson     `json:"roles,omitempty"`
	Users        []*RoleUserJSON `json:"users,omitempty"`
}

// Create client with root role
func CreateClient(client Client) (resClient *ClientJson, err error) {

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return nil, err
	}
	if num, _ := o.QueryTable("client").Filter("fullname", client.Fullname).Count(); num > 0 {
		o.Commit()
		return nil, errors.ErrClientFullNameExisted
	}
	buf := make([]byte, 16)
	if _, err = rand.Read(buf); err != nil {
		o.Commit()
		return nil, err
	}
	var clientId int64
	var rootRoleId int64
	var rootRoleName string
	client.Secret = strings.TrimRight(base64.URLEncoding.EncodeToString(buf), "=")
	if clientId, err = o.Insert(&client); err != nil {
		o.Rollback()
		logs.Error(err)
		return nil, err
	}
	rootRoleName = client.Fullname + "-root"
	role := Role{Name: rootRoleName, ParentId: -1, ClientId: clientId, Description: "root role created by sys", CreatedBy: client.UserId, UpdatedBy: client.UserId}
	if rootRoleId, err = o.Insert(&role); err != nil {
		o.Rollback()
		return nil, err
	} else {
		if res, err := o.Raw("insert into role_user (role_id, user_id, role_type) values (?, ?, ?)", rootRoleId, client.UserId, RoleSuper).Exec(); err != nil {
			o.Rollback()
			return nil, err
		} else if num, err := res.RowsAffected(); err != nil {
			o.Rollback()
			return nil, err
		} else if num != 1 {
			o.Rollback()
			return nil, fmt.Errorf("add %s to rootRoleAdmin failed", client.UserId)
		} else {
			resClient = &ClientJson{
				Id:           clientId,
				Fullname:     client.Fullname,
				Secret:       client.Secret,
				RedirectUri:  client.RedirectUri,
				UserId:       client.UserId,
				RootRoleId:   rootRoleId,
				RootRoleName: rootRoleName,
			}
			o.Commit()
			return resClient, nil
		}
	}
}

func UpdateClient(client Client) (resClient *ClientJson, err error) {
	if client.Fullname == "" {
		return nil, errors.ErrClientFullNameInvalid
	}
	o := orm.NewOrm()
	old := Client{}

	if err = o.QueryTable("client").Filter("id", client.Id).One(&old); err != nil {
		if err == orm.ErrNoRows {
			return nil, errors.ErrClientNotFound
		} else {
			return nil, err
		}
	}
	cols := []string{
		"fullname",
		"redirect_uri",
		"updated_at",
	}
	if _, err = o.Update(&client, cols...); err != nil {
		return nil, err
	} else {
		resClient = &ClientJson{
			Id:          client.Id,
			Fullname:    client.Fullname,
			RedirectUri: client.RedirectUri,
		}
		return
	}
}

//查询全部client，任何人可见，供查看管理员是谁便于申请权限
func GetAllClients() (clients []*ClientJson, err error) {
	o := orm.NewOrm()
	sql := "select id, fullname, user_id from client order by created_at desc"
	if _, err = o.Raw(sql).QueryRows(&clients); err != nil {
		return
	} else if len(clients) == 0 {
		return []*ClientJson{}, nil
	}
	var clientMap = make(map[int64]*ClientJson)
	for i := range clients {
		clientMap[clients[i].Id] = clients[i]
	}
	var users []*RoleUserJSON
	sql2 := fmt.Sprintf("select role.client_id, case role_user.role_type when %d then '%s' when %d then '%s' else '%s' end as role_type, user.id as user_id, user.fullname,user.dn from role_user join role on role_user.role_id = role.id join user on role_user.user_id = user.id where role_user.role_type > %d and role.parent_id = -1", RoleSuper, RoleSuperStr, RoleAdmin, RoleAdminStr, RoleNormalStr, RoleNormal)
	if _, err := o.Raw(sql2).QueryRows(&users); err != nil {
		logs.Error(err)
	} else if len(users) > 0 {
		for i := range users {
			if client, ok := clientMap[users[i].ClientId]; ok {
				client.Users = append(client.Users, users[i])
			}

		}
	}
	return clients, nil
}

func GetClientsByUser(userId string, roleType int) (clients []*ClientJson, err error) {
	o := orm.NewOrm()
	sql := "select id, fullname, user_id from client where id in (select distinct(client_id) from role where id in(select role_id from role_user where user_id =? and role_type >= ?))"
	if roleType == RoleAdmin || roleType == RoleSuper {
		sql = "select * from client where id in (select distinct(client_id) from role where id in(select role_id from role_user where user_id =? and role_type >= ?))"
	}
	if _, err = o.Raw(sql, userId, roleType).QueryRows(&clients); err != nil {
		return
	}
	if len(clients) == 0 {
		return []*ClientJson{}, nil
	} else {
		var clientIds string
		var clientMap = make(map[int64]*ClientJson)
		var roles []*RoleJson
		for i := range clients {
			clientIds += fmt.Sprintf("%d,", clients[i].Id)
			clientMap[clients[i].Id] = clients[i]
		}
		clientIds = clientIds[0 : len(clientIds)-1]
		sql := fmt.Sprintf("select role.*, case role_user.role_type when %d then '%s' when %d then '%s' else '%s' end as role_type from role inner join role_user on role.id = role_user.role_id where role.client_id in (%s) and role_user.user_id = ? and role_user.role_type >= ?", RoleSuper, RoleSuperStr, RoleAdmin, RoleAdminStr, RoleNormalStr, clientIds)
		if _, err := o.Raw(sql, userId, roleType).QueryRows(&roles); err != nil {
			logs.Error(err)
		} else if len(roles) > 0 {
			for i := range roles {
				if client, ok := clientMap[roles[i].ClientId]; ok {
					client.Roles = append(client.Roles, roles[i])
				}

			}
		}
	}
	return
}

func GetClient(id int64) (*Client, error) {
	o := orm.NewOrm()
	client := Client{}
	if err := o.QueryTable("client").Filter("id", id).One(&client); err != nil {
		if err == orm.ErrNoRows {
			return nil, errors.ErrClientNotFound
		} else {
			return nil, err
		}
	} else {
		return &client, nil
	}
}
