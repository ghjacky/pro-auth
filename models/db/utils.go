package db

//RoleType
const (
	RoleNormal = iota
	RoleAdmin
	RoleSuper
	RoleNormalStr = "normal"
	RoleAdminStr  = "admin"
	RoleSuperStr  = "super"
)

var RoleTypeMap = map[int]string{
	RoleNormal: RoleNormalStr,
	RoleAdmin:  RoleAdminStr,
	RoleSuper:  RoleSuperStr,
}

func BuildRoleTree(roles []*Role) []*Role {
	if len(roles) == 0 {
		return []*Role{}
	}
	restRoleMap := map[int64]*Role{} // 还未关联成树的role
	rootRoles := make([]*Role, 0)
	for i := 0; i < len(roles); i++ {
		if roles[i].ParentId == -1 {
			rootRoles = append(rootRoles, roles[i])
		} else {
			restRoleMap[roles[i].Id] = roles[i]
		}
	}
	if len(rootRoles) == 0 {
		for i := 0; i < len(roles); i++ {
			if _, ok := restRoleMap[roles[i].ParentId]; ok {
				continue
			}
			rootRoles = append(rootRoles, roles[i])
		}

		for _, r := range rootRoles {
			delete(restRoleMap, r.Id)
		}
	}

	return doBuildRoleTree(rootRoles, restRoleMap)
}

func doBuildRoleTree(rootRoles []*Role, restRoleMap map[int64]*Role) []*Role {
	for i := 0; i < len(rootRoles); i++ {
		children := make([]*Role, 0)
		for _, v := range restRoleMap {
			if v.ParentId == rootRoles[i].Id {
				children = append(children, v)
			}
		}

		for _, r := range children {
			delete(restRoleMap, r.Id)
		}

		if len(children) > 0 {
			rootRoles[i].Children = doBuildRoleTree(children, restRoleMap)
		}
	}
	return rootRoles
}

/**
将树解构成数组，上层角色中身份高的，覆盖下层低的身份
*/
func unbuildRoleTreeWithRoleType(roleTree []*Role, userRoles map[int64]*RoleJson, parentRoleType string) []*Role {
	roles := make([]*Role, 0)
	for _, v := range roleTree {
		switch parentRoleType {
		case RoleSuperStr:
			v.RoleType = RoleSuperStr
		case RoleAdminStr:
			if role, ok := userRoles[v.Id]; ok {
				if role.RoleType == RoleSuperStr {
					v.RoleType = RoleSuperStr
				} else {
					v.RoleType = RoleAdminStr
				}
			} else {
				v.RoleType = RoleAdminStr
			}
		default:
			if role, ok := userRoles[v.Id]; ok {
				v.RoleType = role.RoleType
			} else {
				v.RoleType = RoleNormalStr
			}
		}
		roles = append(roles, v)
		if len(v.Children) > 0 {
			roles = append(roles, unbuildRoleTreeWithRoleType(v.Children, userRoles, v.RoleType)...)
		}
		v.Children = nil
	}
	return roles
}

func unbuildRoleTree(roleTree []*Role) []*Role {
	roles := make([]*Role, 0)
	for _, v := range roleTree {
		roles = append(roles, v)
		if len(v.Children) > 0 {
			roles = append(roles, unbuildRoleTree(v.Children)...)
		}
		v.Children = nil
	}
	return roles
}
