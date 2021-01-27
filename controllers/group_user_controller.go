package controllers

import (
	"encoding/json"
	"errors"

	"code.xxxxx.cn/platform/auth/models/db"
)

type GroupUserController struct {
	BaseController
}

func (c *GroupUserController) AddGroupUser() {
	if !c.checkScope("groupuser:write") {
		return
	}

	groupUser := db.GroupUser{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &groupUser); err != nil {
		c.Failed(err)
		return
	}

	if _, err := db.GetGroup(groupUser.Group); err != nil {
		c.Failed(err)
		return
	}

	if user, err := db.GetUsersByID([]string{groupUser.User}); err != nil || len(user) == 0 {
		c.Failed(errors.New("User does not exist"))
		return
	}

	err := db.RelateGroupAndUser(&groupUser)
	if err != nil {
		c.Failed(err)
		return
	}

	c.Ok("")
}

func (c *GroupUserController) GetUserByGroup() {
	if !c.checkScope("groupuser:read") {
		return
	}

	group := c.GetString("group")
	userIDs, err := db.GetUsersByGroup(group)
	if err != nil {
		c.Failed(err)
		return
	}

	users, err := db.GetUsersByID(userIDs)
	if err != nil {
		c.Failed(err)
		return
	}

	for _, user := range users {
		user.Password = ""
	}

	c.Ok(users)
}

func (c *GroupUserController) GetGroupByUser() {
	if !c.checkScope("groupuser:read") {
		return
	}

	user := c.GetString("user")
	groupNames, err := db.GetGroupsByUser(user)
	if err != nil {
		c.Failed(err)
		return
	}

	groups, err := db.GetGroupsByName(groupNames)
	if err != nil {
		c.Failed(err)
		return
	}

	c.Ok(groups)
}

func (c *GroupUserController) RemoveGroupUser() {
	if !c.checkScope("groupuser:write") {
		return
	}

	groupUser := db.GroupUser{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &groupUser); err != nil {
		c.Failed(err)
		return
	}

	err := db.RemoveGroupUser(&groupUser)
	if err != nil {
		c.Failed(err)
		return
	}
	c.Ok("")
}
