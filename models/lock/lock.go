package lock

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"sync"
)

func init() {
	orm.RegisterModel(new(DistributedLock))
}

type DistributedLock struct {
	Id   int64      `orm:"pk;auto"`
	Item string     `orm:"unique"`
	lock sync.Mutex `orm:"-"`
	orm  orm.Ormer  `orm:"-"`
}

type Lock interface {
	Lock()
	UnLock()
}

func NewAppLock(appId int64) Lock {
	return New(fmt.Sprintf("client%d", appId))
}

func New(item string) Lock {
	lock := DistributedLock{Item: item}
	o := orm.NewOrm()
	if _, _, err := o.ReadOrCreate(&lock, "Item"); err != nil {
		panic(err)
	} else {
		lock.orm = o
		return &lock
	}
}

func (l *DistributedLock) Lock() {
	l.lock.Lock()
	defer func() {
		if err := recover(); err != nil {
			l.orm.Commit()
			logs.Error(err)
			panic(err)
		}
	}()
	if err := l.orm.Begin(); err != nil {
		panic(err)
	}
	if _, err := l.orm.Raw("select item from distributed_lock where item = ? for update", l.Item).Exec(); err != nil {
		panic(err)
	}
}

func (l *DistributedLock) UnLock() {
	defer l.lock.Unlock()
	if err := l.orm.Commit(); err != nil {
		logs.Error(err)
	}
}
