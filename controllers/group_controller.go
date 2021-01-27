package controllers

import (
	"encoding/json"

	"code.xxxxx.cn/platform/auth/models/db"
)

type GroupController struct {
	BaseController
}

func (c *GroupController) CreateGroup() {
	if !c.checkScope("group:write") {
		return
	}

	group := db.Group{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &group); err != nil {
		c.Failed(err)
		return
	}

	if res, err := db.CreateGroup(group); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

func (c *GroupController) GetGroup() {
	if !c.checkScope("group:read") {
		return
	}

	var groups []db.Group
	name := c.GetString("name")
	nameLikes := c.GetString("likes")
	if name != "" {
		// find one
		group, err := db.GetGroup(name)
		if err != nil {
			c.Failed(err)
			return
		}

		groups = append(groups, *group)
	} else if nameLikes != "" {
		allGroups, err := db.GetGroups(nameLikes)
		if err != nil {
			c.Failed(err)
			return
		}

		for _, group := range allGroups {
			groups = append(groups, *group)
		}
	} else {
		// find all
		allGroups, err := db.GetAllGroups()
		if err != nil {
			c.Failed(err)
			return
		}

		for _, group := range allGroups {
			groups = append(groups, *group)
		}
	}
	c.Ok(groups)
}

func (c *GroupController) UpdateGroup() {
	if !c.checkScope("group:write") {
		return
	}

	group := db.Group{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &group); err != nil {
		c.Failed(err)
		return
	}

	if res, err := db.UpdateGroup(group); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

func (c *GroupController) DeleteGroup() {
	if !c.checkScope("group:write") {
		return
	}

	name := c.GetString("name")

	if err := db.DeleteGroup(name); err != nil {
		c.Failed(err)
	} else {
		c.Ok("")
	}
}
