package db

import (
	"errors"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const (
	GroupUserTable = "group_user"
)

var (
	errGroupUserConflicted = errors.New("group already has a conflicted user")
)

type GroupUser struct {
	ID    int64  `orm:"pk;auto" json:"id"`
	Group string `orm:"size(128)" json:"group,omitempty"`
	User  string `orm:"size(256)" json:"user,omitempty"`
}

func RelateGroupAndUser(gu *GroupUser) error {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	// group and use name is unique
	if o.QueryTable(GroupUserTable).Filter("group", gu.Group).Filter("user", gu.User).Exist() {
		return errGroupUserConflicted
	}

	if _, err := o.Insert(gu); err != nil {
		o.Rollback()
		logs.Error(err)
		return err
	}
	return o.Commit()
}

func RemoveGroupUser(gu *GroupUser) error {
	o := orm.NewOrm()
	_, err := o.QueryTable(GroupUserTable).Filter("group", gu.Group).Filter("user", gu.User).Delete()
	return err
}

func GetUsersByGroup(name string) ([]string, error) {
	o := orm.NewOrm()
	groupUsers := []GroupUser{}
	if _, err := o.QueryTable(GroupUserTable).Filter("group", name).All(&groupUsers); err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	var userIDs []string
	for _, gu := range groupUsers {
		userIDs = append(userIDs, gu.User)
	}

	return userIDs, nil
}

func GetGroupsByUser(name string) (groupNames []string, err error) {
	o := orm.NewOrm()
	groupUsers := []GroupUser{}
	if _, err = o.QueryTable(GroupUserTable).Filter("user", name).All(&groupUsers); err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	for _, gu := range groupUsers {
		groupNames = append(groupNames, gu.Group)
	}
	return
}

func SaveGroupUser(gu *GroupUser) error {
	o := orm.NewOrm()

	// group and use name is unique
	if o.QueryTable(GroupUserTable).Filter("group", gu.Group).Filter("user", gu.User).Exist() {
		return nil
	}

	if _, err := o.Insert(gu); err != nil {
		logs.Error(err)
		return err
	}
	return nil
}
