package controllers

import "code.xxxxx.cn/platform/auth/service"

type SystemController struct {
	BaseController
}

type Setting struct {
	LdapEnabled bool `json:"ldap_enabled"`
}

// Sync upload users and groups
func (c *SystemController) SystemSetting() {
	ldapEnabled := service.Ldap.Enabled()
	setting := &Setting{
		LdapEnabled: ldapEnabled,
	}
	c.Ok(setting)
}
