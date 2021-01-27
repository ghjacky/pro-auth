package db

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.xxxxx.cn/platform/auth/errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type RoleUser struct {
	Id       int64    `orm:"pk;auto" json:"-"`
	RoleId   int64    `orm:"index" json:"role_id"`
	User     *User    `orm:"rel(fk)" json:"user"`
	RoleType RoleType `json:"role_type"`
}

type SimpleRoleUser struct {
	RoleId   int64    `orm:"index" json:"role_id"`
	User     string   `orm:"user_id" json:"user_id"`
	RoleType RoleType `orm:"role_type" json:"role_type"`
}

type RoleUserJSON struct {
	RoleId   int64  `json:"role_id,omitempty"`
	UserId   string `json:"user_id,omitempty"`
	RoleType string `json:"role_type,omitempty"`
	Fullname string `json:"fullname,omitempty"`
	ClientId int64  `json:"client_id,omitempty"`
	Dn       string `json:"dn,omitempty"`
}

type RoleType int

var RoleMap = map[string]int64{
	RoleSuperStr:  RoleSuper,
	RoleAdminStr:  RoleAdmin,
	RoleNormalStr: RoleNormal,
}

func (c *RoleUser) TableUnique() [][]string {
	return [][]string{
		[]string{"RoleId", "User"},
	}
}

func (rt RoleType) MarshalJSON() ([]byte, error) {
	if rt == RoleSuper {
		return json.Marshal(RoleSuperStr)
	} else if rt == RoleAdmin {
		return json.Marshal(RoleAdminStr)
	} else {
		return json.Marshal(RoleNormalStr)
	}
}

func (rt *RoleType) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	switch text {
	case RoleNormalStr:
		*rt = RoleNormal
	case RoleAdminStr:
		*rt = RoleAdmin
	case RoleSuperStr:
		*rt = RoleSuper
	default:
		return &json.UnsupportedValueError{Str: text}
	}
	return nil
}

func GetRoleUsers(clientId, roleId int64) ([]*RoleUserJSON, error) {
	var roleUsers []*RoleUserJSON
	o := orm.NewOrm()
	sql := fmt.Sprintf("select role_id, user_id, if(role_user.role_type=%d,'%s','%s') as role_type from role_user inner join role on role.id = role_user.role_id where role.client_id = ?", RoleAdmin, RoleAdminStr, RoleNormalStr)
	if roleId > 0 {
		sql += fmt.Sprintf(" and role_id = %d", roleId)
	}
	if _, err := o.Raw(sql, clientId).QueryRows(&roleUsers); err != nil {
		return nil, err
	} else if len(roleUsers) == 0 {
		return []*RoleUserJSON{}, nil
	} else {
		return roleUsers, nil
	}
}

func UpdateRoleUser(userId string, roleId int64, roleTypeStr string, isRootRoleSuper bool) (*RoleUserJSON, error) {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return nil, err
	}
	defer o.Commit()
	var roleUser RoleUser
	roleType := RoleNormal
	if roleTypeStr == RoleSuperStr {
		roleType = RoleSuper
	} else if roleTypeStr == RoleAdminStr {
		roleType = RoleAdmin
	}
	if !isRootRoleSuper && roleType == RoleSuper {
		return nil, errors.ErrNoUnauthorizedToRoleTypeSuper
	}
	if err := o.QueryTable("role_user").Filter("role_id", roleId).Filter("user_id", userId).One(&roleUser); err != nil {
		if err == orm.ErrNoRows {
			sql := fmt.Sprintf("insert into role_user (role_id, user_id, role_type) values (%d,'%s',%d) ", roleId, userId, roleType)
			if _, err := o.Raw(sql).Exec(); err != nil {
				o.Rollback()
				return nil, err
			} else {
				return &RoleUserJSON{RoleId: roleId, UserId: userId, RoleType: RoleTypeMap[roleType]}, nil
			}
		} else {
			return nil, err
		}
	} else if roleUser.RoleType == RoleSuper && !isRootRoleSuper {
		return nil, errors.ErrNoUnauthorizedToRoleTypeSuper
	}
	if res, err := o.Raw("update role_user set role_type = ? where role_id= ? and user_id = ? ", roleType, roleId, userId).Exec(); err != nil {
		o.Rollback()
		return nil, err
	} else {
		if _, err := res.RowsAffected(); err != nil {
			o.Rollback()
			return nil, err
		} else {
			return &RoleUserJSON{UserId: roleUser.User.Id, RoleId: roleUser.RoleId, RoleType: RoleTypeMap[roleType]}, nil
		}
	}

}

func AddRoleUser(users []RoleUserJSON, roleId int64) (int64, error) {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return 0, err
	}
	var role Role
	var userIds = make([]string, len(users))
	for i, v := range users {
		userIds[i] = v.UserId
	}
	if err := o.QueryTable("role").
		Filter("id", roleId).One(&role); err != nil {
		o.Commit()
		if err == orm.ErrNoRows {
			return 0, errors.ErrRoleNotFound
		} else {
			logs.Error(err)
			return 0, err
		}
	}
	if num, err := o.QueryTable("user").
		Filter("id__in", userIds).Count(); err != nil && err != orm.ErrNoRows {
		o.Commit()
		return 0, err
	} else if num != int64(len(users)) {
		return 0, errors.ErrRoleNotFound
	}

	sql := "insert into role_user (role_id, user_id, role_type) values "
	for _, v := range users {
		roleType := RoleNormal
		if v.RoleType == RoleSuperStr {
			roleType = RoleSuper
		} else if v.RoleType == RoleAdminStr {
			roleType = RoleAdmin
		}
		sql += fmt.Sprintf("(%d,'%s',%d),", roleId, v.UserId, roleType)
	}
	if res, err := o.Raw(sql[0 : len(sql)-1]).Exec(); err != nil {
		o.Rollback()
		return 0, err
	} else {
		if num, err := res.RowsAffected(); err != nil {
			o.Rollback()
			return 0, err
		} else if num != int64(len(userIds)) {
			o.Rollback()
			return 0, errors.ErrRoleUserExisted
		} else {
			o.Commit()
			return num, nil
		}
	}
}

func AddRoleUserBatch(roleusers []RoleUserJSON) (int64, error) {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return 0, err
	}

	values := []string{}
	for _, v := range roleusers {
		roleType := RoleMap[v.RoleType]
		values = append(values, fmt.Sprintf("(%d,'%s',%d)", v.RoleId, v.UserId, roleType))
	}

	sql := "REPLACE INTO role_user (role_id, user_id, role_type) VALUES " + strings.Join(values, ",")
	res, err := o.Raw(sql).Exec()

	if err != nil {
		o.Rollback()
		return 0, err
	}

	num, err := res.RowsAffected()
	if err != nil {
		o.Rollback()
		return 0, err
	}

	if num < int64(len(roleusers)) {
		o.Rollback()
		return 0, errors.ErrRoleUserExisted
	}

	o.Commit()
	return num, nil
}

func DeleteRoleUser(userIds []string, roleId int64) (int64, error) {
	if len(userIds) == 0 {
		return 0, nil
	}
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return 0, err
	}
	if num, err := o.QueryTable("role_user").Filter("user_id__in", userIds).Filter("role_id", roleId).Delete(); err != nil {
		o.Rollback()
		return 0, err
	} else if num != int64(len(userIds)) {
		o.Rollback()
		return 0, errors.ErrUserNotFoundInRole
	} else {
		o.Commit()
		return num, nil
	}
}

func DeleteUserRole(userId string, roleIds []int64) (int64, error) {
	if len(roleIds) == 0 {
		return 0, nil
	}
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return 0, err
	}
	if num, err := o.QueryTable("role_user").Filter("user_id", userId).Filter("role_id__in", roleIds).Delete(); err != nil {
		o.Rollback()
		return 0, err
	} else if num != int64(len(roleIds)) {
		o.Rollback()
		return 0, errors.ErrUserNotFoundInRole
	} else {
		o.Commit()
		return num, nil
	}
}
