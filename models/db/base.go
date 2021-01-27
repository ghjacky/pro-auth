package db

import (
	"net/url"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

/*
role表相关的依赖关系
                             client
                        ↗             ↖
     user			role             resource      ↑下层的表依赖上层表中对应记录的存在
       ↖        ↗         ↖             ↗
		role_user          role_resource

避免死锁
(1)其中由于role记录和记录之间存在依赖关系构造role tree，当需要判断role tree是否符合某种条件或修改role tree结构时，需要使用分布式锁（确保当前只有一个请求可以修改role tree结构）
(2)role_user依赖role和user，修改role_user的逻辑中，加锁顺序 role → user
(3)role_resource依赖role和resource，修改role_resource的逻辑中，加锁顺序 role → resource


*/

func init() {

	orm.RegisterModel(new(Resource))
	orm.RegisterModel(new(Role))
	orm.RegisterModel(new(RoleUser))
	orm.RegisterModel(new(RoleResource))
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Client))
	orm.RegisterModel(new(Token))
	orm.RegisterModel(new(Group))
	orm.RegisterModel(new(GroupUser))

	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", beego.AppConfig.String("db_sso")+url.QueryEscape("Asia/Shanghai"))

	orm.RunSyncdb("default", false, false)
	orm.DefaultRowsLimit = -1
	if beego.BConfig.RunMode == "dev" {
		orm.Debug = true
	}
	db, _ := orm.GetDB("default")
	db.SetConnMaxLifetime(time.Minute * 4 * 60)
}
