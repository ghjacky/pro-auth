package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	sdk "code.xxxxx.cn/platform/auth/sdk/golang"
	custome "code.xxxxx.cn/platform/auth/sdk/golang/custome"
	log "github.com/sirupsen/logrus"
)

var clientID = flag.Int64("id", 0, "client id")
var clientSecret = flag.String("secret", "", "client secret")
var localAddress = flag.String("local", "", "local server address")
var authHost = flag.String("host", "", "auth server host")
var authScope = flag.String("scope", "", "auth scope")

var oauthUsername = flag.String("username", "", "username")
var oauthUserSecret = flag.String("user-secret", "", "user secret")

func main() {
	fmt.Println("Init config...")
	apiConfig, oauthConfig := initConfigFromArgs()
	fmt.Println("Start test api")
	testAPI(apiConfig)
	if *oauthUsername != "" {
		fmt.Println("\nStart test oauth2 login by secret")
		testOauth2LoginBySecret(oauthConfig)
	}
	fmt.Println("\nStart test oauth2")
	testOauth2(apiConfig, oauthConfig, *localAddress)
}

func initConfigFromArgs() (*sdk.APIConfig, *custome.OauthConfig) {
	flag.Parse()

	apiConfig := sdk.APIConfig{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		APIHost:      *authHost,
	}

	oauthConfig := custome.OauthConfig{
		ClientId:     *clientID,
		ClientSecret: *clientSecret,
		RedirectUri:  "",
		Host:         *authHost,
		Scope:        *authScope,
	}

	return &apiConfig, &oauthConfig
}

func testAPI(config *sdk.APIConfig) {
	service := sdk.NewApiAuth(config)
	fmt.Printf("Test API with config: id: %d, secret: %s, host: %s\n", config.ClientID, config.ClientSecret, config.APIHost)

	fmt.Println("Security test:")
	token := sdk.GenerateJWTToken(config.ClientID, config.ClientSecret)
	fmt.Printf("jwt token %s\n", token)
	fmt.Println(sdk.ParseJWTToken(token, config.ClientSecret))

	// Client API
	client, err := service.GetClient()
	if err != nil {
		log.WithError(err).Error("Get Client Error")
		return
	}
	fmt.Printf("Client Info: id: %d, name: %s, uri: %s, created: %s\n", client.Id, client.Fullname, client.RedirectUri, client.Created)

	newClient, err := service.UpdateClient(client.Fullname, "http://test.test")
	if err != nil {
		log.WithError(err).Error("Update Client Error")
		return
	}
	fmt.Printf("New Client update uri: %s\n", newClient.RedirectUri)

	service.UpdateClient(client.Fullname, client.RedirectUri)
	if err != nil {
		log.WithError(err).Error("Update Client Error")
	}

	// Roles API
	roles, err := service.GetAllRole(true, true, false)
	if err != nil {
		log.WithError(err).Error("Get all roles Failed")
	}
	fmt.Println("Roles num:", len(roles))

	printRolesFunc(roles, service, "")

	if len(roles) > 0 {
		fmt.Printf("Get Role: %v, id: %v", roles[0].Name, roles[0].Id)
		roleTrees, err := service.GetRoleTreeByID(roles[0].Id, true, false, false, true)
		if err != nil {
			log.WithError(err).Error("Get Role Tree Failed")
		} else {
			printRolesFunc(roleTrees, service, "")
		}
	}

	resourceIDs, err := service.AddResource([]sdk.ResourceInfo{
		sdk.ResourceInfo{
			Name:        "testresources",
			Description: "test",
			Data:        "test",
		},
	})
	if err != nil {
		log.WithError(err).Error("Add Resources Failed")
	}

	for _, role := range roles {
		if role.ParentId != -1 {
			_, err = service.AddRoleResourceRelations(role.Id, resourceIDs)
			if err != nil {
				log.WithError(err).Error("Add Role Resources Relation Failed")
			}
		}

		resources, err := service.GetResourceByRole(role.Id)
		if err != nil {
			log.WithError(err).Error("Get Resource By Role Failed!")
			break
		}
		for _, resource := range resources {
			fmt.Printf("Role: %v Resource: %v, Data: %v\n", role.Name, resource.Name, resource.Data)
		}

		if role.ParentId != -1 {
			_, err = service.DeleteRoleResourceRelations(role.Id, resourceIDs)
			if err != nil {
				log.WithError(err).Error("Delete Role Resource Failed")
			}
		}
	}

	_, err = service.DeleteResources(resourceIDs)
	if err != nil {
		log.WithError(err).Error("Delete Resources Failed")
	}

	// Test Group API
	GroupName := "testgroup"
	log.Info("Create Group")
	err = service.CreateGroup(GroupName, "this is a test group", "")
	if err != nil {
		log.WithError(err).Error("Create Group Failed")
	}

	log.Info("Fetch Group")
	groups, err := service.GetAllGroup()
	if err != nil {
		log.WithError(err).Error("Get Group Failed")
	}
	for _, group := range groups {
		fmt.Printf("Group Name: %s, description: %v\n", group.Name, group.Description)
	}

	log.Info("Update Group")
	err = service.UpdateGroup(GroupName, "the test group is changed", "")
	if err != nil {
		log.WithError(err).Error("Update Group Failed!")
	}
	groups, err = service.GetGroup(GroupName)
	if err != nil {
		log.WithError(err).Error("Get Group Failed!")
	}

	if len(groups) > 0 && groups[0].Description == "the test group is changed" {
		log.Info("Update Group Succeed!")
	} else {
		log.Infof("Update Group is not success, group: %v", groups)
	}

	groups, err = service.SearchGroups("a")
	if err != nil {
		log.WithError(err).Error("Get Group Failed!")
	} else if len(groups) > 0 {
		log.Infof("Groups Num %v, first name: %v", len(groups), groups[0].Name)
	}

	err = service.CreateGroupUser(GroupName, "airflow")
	if err != nil {
		log.WithError(err).Error("Create Group User Failed")
	}

	relatedGroups, err := service.GetGroupsByUserID("airflow")
	if err != nil {
		log.WithError(err).Error("Get Group By User Failed")
	}
	fmt.Printf("Airflow related Groups:\n")
	for _, group := range relatedGroups {
		fmt.Printf("group name: %s, group description: %s\n", group.Name, group.Description)
	}

	relatedUser, err := service.GetUsersByGroup(GroupName)
	if err != nil {
		log.WithError(err).Error("Get User by Group Failed")
	}
	fmt.Printf("Group: %s 's user\n", GroupName)
	for _, user := range relatedUser {
		fmt.Printf("user name: %s, dn: %s, email: %s\n", user.Fullname, user.Dn, user.Email)
	}

	err = service.DeleteGroupUser(GroupName, "airflow")
	if err != nil {
		log.WithError(err).Error("Delete Group user failed")
	}

	log.Info("Delete Group")
	err = service.DeleteGroup(GroupName)
	if err != nil {
		log.WithError(err).Error("Delete Group Failed")
	}
}

func testOauth2(apiconfig *sdk.APIConfig, config *custome.OauthConfig, address string) {
	service := sdk.NewApiAuth(apiconfig)

	client, err := service.GetClient()
	if err != nil {
		log.WithError(err).Error("Get Client Error")
		return
	}

	redirectUri := "http://" + address + "/oauth2"
	_, err = service.UpdateClient(client.Fullname, redirectUri)
	if err != nil {
		log.WithError(err).Error("Update Client Error")
		return
	}

	config.RedirectUri = redirectUri
	oauth2 := custome.NewAuthService(config)
	fmt.Printf("Login url: %s\n", oauth2.LoginURL("test"))
	fmt.Println("Please Click the Login url to sign up this server")

	http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		code, ok := r.URL.Query()["code"]
		if !ok || len(code) == 0 {
			log.Error("Oauth callback has no any code!")
			w.Write([]byte("Sorry, The server could not find `code` in url"))
			return
		}
		user, err := oauth2.Login(code[0])
		if err != nil {
			log.WithError(err).Errorf("Login with code: %s failed!, could not get User", code[0])
			w.Write([]byte(fmt.Sprintf("Login Failed! code:%s not right, error: %v", code[0], err)))
			return
		}
		fmt.Println("Login Success!")
		fmt.Printf("User: %s, email: %s, accesstoken: %s\n", user.Fullname, user.Email, user.Token.AccessToken)

		resources, err := oauth2.LoadUserResource(user)
		if err != nil {
			log.WithError(err).Errorf("Load user resource Failed!")
			w.Write([]byte(fmt.Sprintf("Load user resource Failed!, error: %v", err)))
			return
		}
		user.Resources = resources
		for _, resource := range resources {
			fmt.Printf("Resource: %d, desc: %s, data: %s\n", resource.Id, resource.Description, resource.Data)
		}

		_, err = service.UpdateClient(client.Fullname, client.RedirectUri)
		if err != nil {
			log.WithError(err).Error("Update Client Error")
			w.Write([]byte(fmt.Sprintf("Update Client Error, error:%v", err)))
			return
		}

		// Check third token
		ok, err = service.CheckThirdToken(user.Token.AccessToken, strconv.Itoa(int(apiconfig.ClientID)), user.Id)
		if err != nil {
			log.WithError(err).Error("Exception happend when check thrid party token")
			w.Write([]byte("Check Third Token failed!"))
			return
		}
		fmt.Printf("Third Party Check ans: %v \n", ok)

		w.Write([]byte("Please return to your program :)"))

		// Logout mimic
		// oauth2.Logout(user.Token)
	})

	addr := strings.Split(address, ":")
	if len(addr) <= 1 {
		log.Error("local address should be this format: '{ip}:{port}'")
	}

	err = http.ListenAndServe(":"+addr[1], nil)
	if err != nil {
		log.WithError(err).Error("Start Local Server Failed!")
	}
}

func testOauth2LoginBySecret(config *custome.OauthConfig) {
	oauth2 := custome.NewAuthService(config)
	user, err := oauth2.LoginBySecret(*oauthUsername, *oauthUserSecret)
	if err != nil {
		log.WithError(err).Error("Login by Secret Failed")
	}
	log.Infof("user: %v, id: %v, email: %v", user.Fullname, user.Id, user.Email)
}

func printRolesFunc(roles []*sdk.Role, service *sdk.APIAuth, prefix string) {
	for _, role := range roles {
		fmt.Printf("%sTree name: %v, id: %v, parent: %v, resource num: %d \n", prefix, role.Name, role.Id, role.ParentId, len(role.Resources))
		for _, roleResource := range role.Resources {
			resource := roleResource.Resource
			fmt.Printf("%s\tResource: %v, id:%v, Data: %v, Description: %v, Role id: %d \n", prefix, resource.Name, resource.Id, resource.Data, resource.Description, roleResource.RoleID)
		}

		for _, roleUser := range role.Users {
			user := roleUser.User
			fmt.Printf("%s\tUser: %v, name: %s, email %s, type: %s, Role_id: %d \n", prefix, user.Id, user.Fullname, user.Email, user.Type, roleUser.RoleId)
		}

		users, err := service.GetUsersOfRole(role.Id)
		if err == nil {
			for _, user := range users {
				userRoles, err := service.GetUserRoles(user.UserId, false, false, false, false)
				if err != nil {
					log.WithError(err).Error("Get User of Role Failed!")
					continue
				}
				for _, ur := range userRoles {
					fmt.Printf("%v\t\t user roles: %v, %v", prefix, ur.Id, ur.Name)
				}
			}
		} else {
			log.WithError(err).Error("Get User of Role Failed!")
		}

		if len(role.Children) > 0 {
			printRolesFunc(role.Children, service, "   "+prefix)
		}
	}
	fmt.Println()
	fmt.Println()
}
