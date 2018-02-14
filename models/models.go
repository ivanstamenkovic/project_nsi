package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
)

//User data struct
type User struct {
	Id       int64
	Username string `orm:"unique"`
	Password string
	Email    string `orm:"unique"`
	Token    string `orm:"unique"`

	Verified bool

	//Created time.Time `orm:"auto_now_add;type(timestamp)"`
	//Updated time.Time `orm:"auto_now;type(timestamp)"`
}

func init() {

	orm.RegisterModel(new(User))

	orm.RegisterDataBase("default", "postgres",
		"user=beego password=beego host=127.0.0.1 port=5432 dbname=firstdb sslmode=disable")

	orm.RunSyncdb("default", false, true)

	orm.Debug = true
}
