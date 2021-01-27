package main

import (
	"encoding/base64"
	"log"

	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego/orm"
)

func main() {

	o := orm.NewOrm()
	var users []*db.User
	count, _ := o.QueryTable("user").All(&users)
	log.Printf("Get %d users", count)

	for _, user := range users {
		_, err := base64.StdEncoding.DecodeString(user.Password)
		if err == nil {
			log.Printf("skip %s, %s", user.Id, user.Password)
			continue
		}
		base64Pwd := base64.StdEncoding.EncodeToString([]byte(user.Password))
		log.Printf("%s, %s, %s", user.Id, user.Password, base64Pwd)
		user.Password = base64Pwd

		_, err = o.Update(user, "password")
		if err != nil {
			log.Fatalf("update %s password failed: %v", user.Id, err)
		} else {
			log.Printf("update %s password succeed", user.Id)
		}
	}
}
