package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"gopkg.in/ldap.v2"
)

var Ldap *LdapWrap

const (
	DEFAULT_ORG = "xxxxx"
)

func init() {
	if len(beego.AppConfig.String("ldap_host")) == 0 {
		logs.Info("LDAP not set, LDAP NOT Enabled.")
		Ldap = &LdapWrap{client: nil, enabled: false}
		return
	}
	if client, err := NewClient(
		beego.AppConfig.String("ldap_host"),
		beego.AppConfig.DefaultInt("ldap_port", 389),
		beego.AppConfig.String("ldap_base"),
		beego.AppConfig.String("ldap_user"),
		beego.AppConfig.String("ldap_password"),
		beego.AppConfig.Strings("ldap_employee_ous"),
	); err != nil {
		panic(err)
	} else {
		Ldap = &LdapWrap{client: client, enabled: true}
	}
}

type LdapWrap struct {
	client  *Client
	enabled bool
}

func (l *LdapWrap) Auth(mail string, passwd string) (bool, error) {
	var err error
	for i := 0; i < 2; i++ {
		if err = l.client.refreshConn(); err != nil {
			logs.Error(err)
		} else {
			break
		}
	}
	if err != nil {
		return false, fmt.Errorf("failed to connect ldap")
	}
	return l.client.Auth(mail, passwd)
}

func (l *LdapWrap) Enabled() bool {
	return l.enabled
}

func (l *LdapWrap) GetUserInfo(mail string) (dn, cnName string, err error) {
	var result *ldap.SearchResult
	result, err = l.client.SearchForUser(mail)
	if err != nil {
		return
	}
	if len(result.Entries) == 0 || len(result.Entries[0].Attributes) == 0 {
		err = errors.New("can't get user's detail info from ldap")
		return
	}
	dn = result.Entries[0].DN
	for _, attr := range result.Entries[0].Attributes {
		if strings.ToUpper(attr.Name) == "CN" {
			if len(attr.Values) == 0 {
				return
			}
			cnName = attr.Values[0]
			err = nil
			return
		}
	}
	return
}

func (l *LdapWrap) GetAllLdapUsers() ([]*db.User, error) {
	allUsers, err := l.searchOUUsers(l.client.baseDn)
	if err != nil {
		return nil, err
	}
	deletedUsers, err := l.searchOUUsers("ou=Disable-user,dc=xxxxx,dc=cn")
	if err != nil {
		logs.Error("Get deleted users failed: %v", err)
		return nil, err
	}
	allUsers = append(allUsers, deletedUsers...)
	return allUsers, nil
}

func (l *LdapWrap) searchOUUsers(dn string) ([]*db.User, error) {
	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(&(objectClass=User)(objectCategory=Person))",
		[]string{"cn", "sAMAccountName", "mail", "telephoneNumber"},
		nil,
	)
	var result *ldap.SearchResult
	result, err := l.client.search(searchRequest)
	if err != nil {
		return nil, err
	}
	var users []*db.User
	for _, entry := range result.Entries {
		id := entry.GetAttributeValue("sAMAccountName")
		email := entry.GetAttributeValue("mail")
		fullname := entry.GetAttributeValue("cn")
		phone := entry.GetAttributeValue("telephoneNumber")
		orgDomain := beego.AppConfig.String("org_domain")
		org := DEFAULT_ORG
		if len(org) > 0 && len(strings.Split(org, ".")) > 0 {
			org = strings.Split(orgDomain, ".")[0]
		}
		status := l.checkUserStatus(entry.DN)
		logs.Info("user %s status: %v", entry.DN, status)
		user := &db.User{
			Id:           id,
			Fullname:     fullname,
			Dn:           entry.DN,
			Email:        email,
			Phone:        phone,
			Type:         db.UserTypeLdap,
			Status:       status,
			Organization: org,
		}
		users = append(users, user)

	}
	return users, nil
}

func (l *LdapWrap) checkUserStatus(dn string) string {
	for _, ou := range l.client.employeeOUs {
		if strings.Contains(dn, fmt.Sprintf("OU=%s", ou)) {
			return db.UserStatusActive
		}
	}
	return db.UserStatusDelete
}

func (l *LdapWrap) GetAllLdapGroupsAndMembers() ([]*db.Group, error) {
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("OU=groups,OU=xxxxx,DC=xxxxx,DC=cn"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		[]string{"name", "mail", "member"},
		nil,
	)
	var result *ldap.SearchResult
	result, err := l.client.search(searchRequest)
	if err != nil {
		return nil, err
	}
	var groups []*db.Group
	for _, entry := range result.Entries {
		name := entry.GetAttributeValue("name")
		email := entry.GetAttributeValue("mail")
		group := &db.Group{
			Name:        name,
			Email:       email,
			Description: "",
			Members:     l.GetAccountsByDn(entry.GetAttributeValues("member")),
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (l *LdapWrap) GetAccountsByDn(dns []string) (accounts []string) {
	for _, dn := range dns {
		searchRequest := ldap.NewSearchRequest(
			dn,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			"(objectClass=person)",
			[]string{"sAMAccountName"},
			nil,
		)
		result, err := l.client.search(searchRequest)
		if err == nil && len(result.Entries) > 0 {
			accounts = append(accounts, result.Entries[0].GetAttributeValue("sAMAccountName"))
		}
	}
	return
}

var (
	ErrNoSuchObject = errors.New("no such object in ldap")
	ErrAuthFail     = errors.New("user or pwd wrong in ldap")
)

// Client
type Client struct {
	ldapHost        string
	ldapPort        int
	username        string
	password        string
	baseDn          string
	employeeOUs     []string
	conn            *ldap.Conn
	lastConnectTime time.Time //上次重连时间点，这里不区分是否重连成功
	lock            *sync.Mutex
}

// NewLdapClient 返回一个暴露了查询方法的ldap客户端
// ldapHost 1.1.1.1
// port     389
// baseDn   OU=xxx,DC=xxx
func NewClient(ldapHost string, ldapPort int, baseDn string, username string, password string, ous []string) (*Client, error) {
	client := &Client{
		ldapHost:    ldapHost,
		ldapPort:    ldapPort,
		username:    username,
		password:    password,
		baseDn:      baseDn,
		employeeOUs: ous,
		lock:        &sync.Mutex{},
	}

	// 下面为初始化代码，和重连代码分开
	err := client.refreshConn()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// refreshConn 更新当前的连接, 该方法应该只在reconnect中和初始化时候被调用
func (c *Client) refreshConn() error {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.ldapHost, c.ldapPort))
	if err != nil {
		// todo 这里应该发送邮件报警
		logs.Error("failed to connect ldap server, err %+v", err)
		return err
	}

	// First bind with a read only user
	err = conn.Bind(c.username, c.password)
	if err != nil {
		logs.Error("failed to login ldap, err %+v", err)
		return err
	}
	c.conn = conn
	c.lastConnectTime = time.Now()
	logs.Info("successfully refresh ldap connection")
	return nil
}

// reconnect 我们通过限制重连间隔来防止雪崩, 我们只有在实际发生了重连并且失败的情况下才会返回error
func (c *Client) reconnect() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// 我们只有在连接为空，或者上次重连超过5秒后才会进行重连
	if !(c.conn == nil || c.lastConnectTime.IsZero() || time.Now().After(c.lastConnectTime.Add(5*time.Second))) {
		logs.Warn("ldap connection reconnect too often, will skip")
		return nil
	}
	return c.refreshConn()
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	logs.Info("successfully close ldap client")
	c.conn = nil
	c.lastConnectTime = time.Time{}
	return nil
}

func (c *Client) search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		logs.Error("failed to search ldap for request %+v, err %+v", searchRequest, err)
		// 这里需要判断是否是连接错误，如果是的话，那么就需要重连，然后重试
		if !strings.Contains(err.Error(), "Network Error") {
			return nil, err
		}
		if c.reconnect() == nil {
			sr, err = c.conn.Search(searchRequest)
			if err != nil {
				logs.Error("failed to search ldap for request %+v again, will not retry, err %+v", searchRequest, err)
				return nil, err
			}
			return sr, nil
		}
		logs.Error("failed to reconnect ldap for request %+v, will not retry, err %+v", searchRequest, err)
		return nil, err
	}
	return sr, nil
}

// SearchForUser 通过upn也就是邮箱来查询用户, todo 根据上层调用方使用情况，返回值可以做一定的包装
func (c *Client) SearchForUser(upn string) (*ldap.SearchResult, error) {
	username := strings.Split(upn, "@")[0]
	searchRequest := ldap.NewSearchRequest(
		c.baseDn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"cn", "userPrincipalName"},
		nil,
	)

	return c.search(searchRequest)
}

// SearchForOU  OUs名字类似于 OU=***,OU=***,OU=**,OU=**,DC=*,DC=*
// 这个方法其实很不实用, todo 根据上层调用方使用情况，返回值可以做一定的包装
func (c *Client) SearchForOU(OUs string) (*ldap.SearchResult, error) {
	searchRequest := ldap.NewSearchRequest(
		OUs,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(&(objectClass=organizationalUnit)(objectClass=top))",
		[]string{"name", "ou"},
		nil,
	)

	sr, err := c.search(searchRequest)
	if err != nil {
		logs.Error("failed to search ldap for request %+v, err %+v", searchRequest, err)
		if strings.Contains(err.Error(), "No Such Object") {
			return nil, ErrNoSuchObject
		}
		return nil, err
	}
	return sr, nil
}

func checkEmail(email string) (b bool) {
	if m, _ := regexp.MatchString("^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(.[a-zA-Z0-9_-])+", email); !m {
		return false
	}
	return true
}

// Auth 验证用户名密码， todo ut 测试， 明文没有问题么
// mail需要带上后缀
func (c *Client) Auth(mail string, passwd string) (bool, error) {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.ldapHost, c.ldapPort))

	if err != nil {
		// todo 这里应该发送邮件报警
		logs.Error("failed to connect ldap server, err %+v", err)
		return false, err
	}

	defer conn.Close()

	if !checkEmail(mail) || passwd == "" {
		return false, errors.New("mail invalid or password is empty")
	}

	// First bind with a read only user
	// 经测试，Bind方法如果接收的参数mail和password有一个为空，则返回nil
	// Search for the given username
	username := strings.Split(mail, "@")[0]
	//logs.Error(mail)
	//logs.Error(username)
	searchRequest := ldap.NewSearchRequest(
		"dc=xxxxx,dc=cn",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"dn"},
		nil,
	)
	sr, err := c.search(searchRequest)
	if err != nil || len(sr.Entries) == 0 {
		logs.Error("failed to find %s, err %+v", mail, err)
		return false, ErrAuthFail
	}

	userdn := sr.Entries[0].DN
	err = conn.Bind(userdn, passwd)
	if err != nil {
		logs.Error("failed to login ldap, err %+v", err)
		if strings.Contains(err.Error(), "Invalid Credentials") {
			return false, ErrAuthFail
		}
		return false, err
	}
	return true, nil

}
