package db

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"code.xxxxx.cn/platform/auth/errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserTable = "user"
)

type User struct {
	// 账号，数据库中不存储邮箱后缀
	Id           string     `orm:"pk" json:"id"`
	Fullname     string     `orm:"size(128)" json:"fullname,omitempty"`
	Password     string     `orm:"size(512)" json:"password,omitempty"`
	Email        string     `orm:"size(256)" json:"email,omitempty"`
	Wechat       string     `orm:"size(128)" json:"wechat,omitempty"`
	Phone        string     `orm:"size(15)" json:"phone,omitempty"`
	Type         string     `orm:"size(10)" json:"type,omitempty"`
	Dn           string     `orm:"size(512);null" json:"dn,omitempty"`
	Status       string     `orm:"size(10)" json:"status,omitempty"`
	CreatedAt    *time.Time `orm:"auto_now_add" json:"created_at,omitempty"`
	UpdatedAt    *time.Time `orm:"auto_now;auto_now_add" json:"updated_at,omitempty"`
	RoleType     string     `orm:"-" json:"role_type,omitempty"`
	Roles        []*Role    `orm:"-" json:"roles,omitempty"`
	Organization string     `orm:"default(xxxxx)" json:"organization,omitempty"`
	Secret       string     `orm:"size(32)" json:"secret,omitempty"`
}

type UserJSON struct {
	Id           string     `orm:"pk" json:"id"`
	Fullname     string     `orm:"size(128)" json:"fullname,omitempty"`
	Email        string     `orm:"size(256)" json:"email,omitempty"`
	Wechat       string     `orm:"size(128)" json:"wechat,omitempty"`
	Phone        string     `orm:"size(15)" json:"phone,omitempty"`
	Type         string     `orm:"size(10)" json:"type,omitempty"`
	Dn           string     `orm:"size(512);null" json:"dn,omitempty"`
	Status       string     `orm:"size(10)" json:"status,omitempty"`
	CreatedAt    *time.Time `orm:"auto_now_add" json:"created_at,omitempty"`
	UpdatedAt    *time.Time `orm:"auto_now;auto_now_add" json:"updated_at,omitempty"`
	RoleType     string     `orm:"-" json:"role_type,omitempty"`
	Organization string     `orm:"default(xxxxx)" json:"organization,omitempty"`
	Secret       string     `orm:"size(32)" json:"secret,omitempty"`
}

const (
	UserTypeLdap    string = "ldap"
	UserTypeSys     string = "system"
	UserTypeSpecial string = "special"

	UserStatusActive string = "active"
	UserStatusFrozen string = "frozen"
	UserStatusDelete string = "delete"

	UserDefaultPwd string = "123456"
)

func GetAllUsers(returnPwd bool) (users []*User, err error) {
	o := orm.NewOrm()
	fields := []string{
		"id", "fullname", "email", "wechat", "phone",
		"type", "organization", "dn", "status", "created_at", "updated_at"}
	if returnPwd {
		fields = append(fields, "password")
	}
	if _, err = o.QueryTable(UserTable).Exclude("status", "delete").All(&users, fields...); err != nil {
		if err == orm.ErrNoRows {
			return []*User{}, nil
		}
		return nil, err
	}
	return
}

func GetUserEncryptedPwds(userIds []string) (users []*User, err error) {
	o := orm.NewOrm()
	fields := []string{"id", "password"}
	if _, err = o.QueryTable(UserTable).Filter("id__in", userIds).All(&users, fields...); err != nil {
		if err == orm.ErrNoRows {
			return []*User{}, nil
		}
		return nil, err
	}
	return
}

func GetUsersByOffsetAndLimit(offset, limit int, returnPwd bool, filters map[string]string) (users []*User, err error) {
	o := orm.NewOrm()
	fields := []string{
		"id", "fullname", "email", "wechat", "phone",
		"type", "organization", "dn", "status", "created_at", "updated_at"}
	if returnPwd {
		fields = append(fields, "password")
	}
	qs := o.QueryTable(UserTable).Exclude("status", "delete")
	for k, v := range filters {
		if len(v) > 0 {
			qs = qs.Filter(fmt.Sprintf("%s__contains", k), v)
		}
	}
	if _, err = qs.Limit(limit, offset).All(&users, fields...); err != nil {
		if err == orm.ErrNoRows {
			return []*User{}, nil
		}
		return nil, err
	}
	return
}

func CountUsers(filters map[string]string) int64 {
	o := orm.NewOrm()
	qs := o.QueryTable(UserTable).Exclude("status", "delete")
	for k, v := range filters {
		if len(v) > 0 {
			qs = qs.Filter(fmt.Sprintf("%s__contains", k), v)
		}
	}
	count, err := qs.Count()
	if err != nil {
		return 0
	}
	return count
}

func CreateUser(user User) (resUser *User, err error) {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return nil, err
	}
	if num, _ := o.QueryTable(UserTable).Filter("id", user.Id).Count(); num > 0 {
		o.Commit()
		return nil, errors.ErrUserExisted
	}

	pwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.Password = base64.StdEncoding.EncodeToString(pwd)
	if _, err := o.Insert(&user); err != nil {
		o.Rollback()
		logs.Error(err)
		return nil, err
	}

	o.Commit()
	return &user, err
}

func Update(user User) (resUser *User, err error) {
	o := orm.NewOrm()
	var old User
	if err := o.QueryTable(UserTable).Filter("id", user.Id).One(&old); err != nil {
		if err == orm.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	cols := []string{
		"updatedAt",
	}
	if user.Fullname != "" {
		cols = append(cols, "fullname")
	}
	if user.Email != "" {
		cols = append(cols, "email")
	}
	if user.Phone != "" {
		cols = append(cols, "phone")
	}
	if user.Wechat != "" {
		cols = append(cols, "wechat")
	}
	if user.Dn != "" {
		cols = append(cols, "dn")
	}
	if user.Status != "" {
		cols = append(cols, "status")
	}
	if user.Password != "" {
		cols = append(cols, "password")
	}
	if user.Organization != "" {
		cols = append(cols, "organization")
	}
	if user.Secret != "" {
		cols = append(cols, "secret")
	}
	if _, err := o.Update(&user, cols...); err != nil {
		return nil, err
	}

	// Actually I dont understand what this code done ; from @rpzhang
	user.Fullname = old.Fullname
	user.Phone = old.Phone
	user.Wechat = old.Wechat
	user.Dn = old.Dn
	user.Status = old.Status
	user.RoleType = old.RoleType
	user.Organization = old.Organization
	user.UpdatedAt = old.UpdatedAt
	return &user, nil
}

func GetUsersByID(ids []string) ([]*User, error) {
	o := orm.NewOrm()
	var users []*User
	if _, err := o.QueryTable(UserTable).FilterRaw("id", " in ('"+strings.Join(ids, "','")+"')").Exclude("status", "delete").All(&users); err != nil {
		if err == orm.ErrNoRows {
			return []*User{}, nil
		}
		return nil, err
	}
	return users, nil
}

// Save update user, insert if not exist
func SaveUser(user User) (resUser *User, err error) {
	o := orm.NewOrm()
	var old User
	if err := o.QueryTable("user").Filter("id", user.Id).One(&old); err != nil {
		if err == orm.ErrNoRows {
			// insert user
			if len(user.Password) == 0 {
				pwd, _ := bcrypt.GenerateFromPassword([]byte(UserDefaultPwd), bcrypt.DefaultCost)
				user.Password = base64.StdEncoding.EncodeToString(pwd)
			}
			if _, err := o.Insert(&user); err != nil {
				logs.Error(err)
				return nil, err
			}
			return &user, err
		}
	}
	fields := []string{"fullname", "email", "wechat", "phone", "type", "organization", "dn", "status", "created_at", "updated_at"}
	if _, err := o.Update(&user, fields...); err != nil {
		return nil, err
	}
	return &user, nil
}
