package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"code.xxxxx.cn/platform/auth/controllers"
	"code.xxxxx.cn/platform/auth/models/db"
	sdk "code.xxxxx.cn/platform/auth/sdk/golang"
	"code.xxxxx.cn/platform/auth/service"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type SyncConfig struct {
	ClientId     int64
	ClientSecret string
	SyncUrls     []string
	PeriodHour   int64
	SyncClients  []string
	SyncOptions  []string
	MaxRetry     int
}

var clientId = flag.Int64("id", 0, "client id")
var clientSecret = flag.String("secret", "", "client secret")
var syncUrls = flag.String("urls", "", "sync api urls")
var periodHour = flag.Int64("period", 1, "period hour")
var syncClients = flag.String("client-ids", "", "client ids to sync")
var syncOptions = flag.String("options", "group", "sync options, available options: 'client', 'user', 'group', split with comma")
var maxRetry = flag.Int("retry", 3, "max retry times")

func main() {
	syncConfig := initArgs()
	if len(syncConfig.ClientSecret) == 0 || len(syncConfig.SyncUrls) == 0 {
		logs.Error("[ERROR]secret or url is empty, exited")
		os.Exit(-1)
		return
	}
	t := time.NewTicker(time.Duration(syncConfig.PeriodHour) * time.Hour)
	defer t.Stop()

	for {
		syncConfig.syncData()
		<-t.C
	}
}

func initArgs() *SyncConfig {
	flag.Parse()
	config := &SyncConfig{
		ClientId:     *clientId,
		ClientSecret: *clientSecret,
		SyncUrls:     strings.Split(*syncUrls, ","),
		PeriodHour:   *periodHour,
		SyncClients:  strings.Split(*syncClients, ","),
		SyncOptions:  strings.Split(*syncOptions, ","),
		MaxRetry:     *maxRetry,
	}
	logs.Info("Sync Config: %+v\n", config)
	return config
}

func (s *SyncConfig) syncData() {
	if len(s.SyncOptions) == 0 {
		logs.Error("Empty options string")
	}

	body := &controllers.BodyData{}

	for _, option := range s.SyncOptions {
		switch strings.ToLower(option) {
		case "client":
			//Sync Client
			clients, err := s.getSyncClientResources()
			if err != nil {
				logs.Error("Get Synced Client Resources Failed, err: %v", err)
			}
			body.ClientRows = clients
		case "user":
			// Sync ldap users
			users, err := getSyncUserResources()
			if err != nil {
				logs.Error("Get Synced User Resources Failed, err: %v", err)
			}
			logs.Info("Got %d ldap users", len(users))
			body.Users = users
		case "group":
			// Sync ldap groups
			groups, err := getSyncGroupResources()
			if err != nil {
				logs.Error("Get Synced Group Resources Failed, err: %v", err)
			}
			body.Groups = groups
			logs.Info("Got %d ldap groups", len(groups))
		default:
			logs.Error("Invalid options")
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		logs.Info("Json marshal failed: %v", err)
		return
	}

	for _, url := range s.SyncUrls {
		go func(requestUrl string) {
			succeed := false
			count := 0
			for !succeed && count < 3 {
				succeed, err = s.doPost(requestUrl, string(data))
				if !succeed {
					count++
					logs.Error("Sync the %d times failed: %v\n", count, err)
				}
			}

			if succeed {
				logs.Info("Sync all data succeed for %s", requestUrl)
			}
		}(url)
	}
	logs.Info("Posted all data")
}

func (s *SyncConfig) getSyncClientResources() ([]*controllers.ClientRow, error) {
	o := orm.NewOrm()
	// synced app 1,19,21,26,35
	var clients []*db.Client
	count, err := o.QueryTable("client").Filter("id__in", s.SyncClients).All(&clients)
	if err != nil {
		logs.Error("Get clients failed: %v", err)
		return nil, err
	}
	logs.Info("Sync %d clients", count)

	var clientRows []*controllers.ClientRow
	for _, client := range clients {
		// sync roles
		rolesTree, err := db.GetRoleByClient(client.Id, true, true, true)
		if err != nil {
			logs.Error("Get roles tree failed: %v", err)
			return nil, err
		}
		// sync resources
		var resources []*db.Resource
		count, err = o.QueryTable("resource").Filter("client_id", client.Id).All(&resources)
		if err != nil {
			logs.Error("Get resources failed: %v", err)
			return nil, err
		}
		logs.Info("Got %d resources", count)
		var resourceRows []*controllers.ResourceRow
		for _, resource := range resources {
			var roleResources []*db.RoleResource
			count, err = o.QueryTable("role_resource").Filter("role_id", resource.Id).All(&roleResources)
			var roles []*db.Role
			for _, roleRes := range roleResources {
				var role = &db.Role{}
				o.QueryTable("role").Filter("id", roleRes.RoleId).One(role)
				roles = append(roles, role)
			}
			var resourceRow = &controllers.ResourceRow{
				Resource: resource,
				Roles:    roles,
			}
			resourceRows = append(resourceRows, resourceRow)
		}

		clientRow := &controllers.ClientRow{
			Client: client,
			//RoleRows:     roleRows,
			ResourceRows: resourceRows,
			RoleTree:     rolesTree,
		}
		clientRows = append(clientRows, clientRow)
	}
	return clientRows, nil
}

func getSyncUserResources() ([]*db.User, error) {
	return service.Ldap.GetAllLdapUsers()
}

func getSyncGroupResources() ([]*db.Group, error) {
	return service.Ldap.GetAllLdapGroupsAndMembers()
}

func (s *SyncConfig) doPost(url, data string) (bool, error) {
	jwt := sdk.GenerateJWTToken(s.ClientId, s.ClientSecret)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Client %v", jwt),
		"content-type":  "application/json",
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 30 * time.Minute,
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return false, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	logs.Info("Response: %v\n", resp)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}
