package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const (
	GroupTable = "group"
)

var (
	errGroupConflicted = errors.New("this group is conflicted")
)

type Group struct {
	ID          int64    `orm:"pk;auto" json:"id"`
	Name        string   `orm:"size(128);unique" json:"name,omitempty"`
	Email       string   `orm:"size(128)" json:"email,omitempty"`
	Description string   `orm:"size(256)" json:"description,omitempty"`
	Members     []string `orm:"-" json:"members,omitempty"`
}

func GetAllGroups() (groups []*Group, err error) {
	o := orm.NewOrm()
	if _, err = o.QueryTable(GroupTable).All(&groups); err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	return
}

func CreateGroup(group Group) (resGroup *Group, err error) {
	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return nil, err
	}

	if _, err := o.Insert(&group); err != nil {
		o.Rollback()
		logs.Error(err)
		return nil, err
	}
	o.Commit()
	return &group, err
}

func GetGroup(name string) (*Group, error) {
	o := orm.NewOrm()
	group := Group{}
	if err := o.QueryTable(GroupTable).Filter("name", name).One(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// GetGroups return searched result by name
func GetGroups(name string) ([]*Group, error) {
	o := orm.NewOrm()
	var groups []*Group
	if _, err := o.QueryTable(GroupTable).Filter("name__contains", name).All(&groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func GetGroupsByName(names []string) ([]*Group, error) {
	o := orm.NewOrm()
	var groups []*Group
	if _, err := o.QueryTable(GroupTable).FilterRaw("name", " in ('"+strings.Join(names, "','")+"')").All(&groups); err != nil {
		if err == orm.ErrNoRows {
			return []*Group{}, nil
		}
		return nil, err
	}
	return groups, nil
}

func UpdateGroup(group Group) (resGroup *Group, err error) {
	o := orm.NewOrm()
	params := orm.Params{
		"description": group.Description,
		"email":       group.Email,
	}
	if _, err := o.QueryTable(GroupTable).Filter("name", group.Name).Update(params); err != nil {
		logs.Error(err)
		return nil, err
	}

	return &group, err
}

func DeleteGroup(groupName string) error {
	o := orm.NewOrm()
	// Only allowed delete empty group
	count, err := o.QueryTable(GroupUserTable).Filter("group", groupName).Count()
	if err != nil {
		return fmt.Errorf("Query related user error: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("This Group Still has some user, should not be deleted")
	}

	_, err = o.QueryTable(GroupTable).Filter("name", groupName).Delete()

	return err
}

func SaveGroup(group *Group) (*Group, error) {
	o := orm.NewOrm()
	if exist := o.QueryTable(GroupTable).Filter("name", group.Name).Exist(); !exist {
		//logs.Info("group %+v", group)
		if _, err := o.Insert(group); err != nil {
			logs.Error(err)
			return nil, err
		}
	}
	params := orm.Params{
		"description": group.Description,
		"email":       group.Email,
	}
	if _, err := o.QueryTable(GroupTable).Filter("name", group.Name).Update(params); err != nil {
		logs.Error(err)
		return nil, err
	}
	return group, nil
}
