package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/ivanstamenkovic/project_nsi/controllers"
)

var LoginFilter = func(ctx *context.Context) {

	token := ctx.Input.Header("token")
	_, err := controllers.ValidateToken(token)
	if err == nil {
		return
	}
	ctx.Output.SetStatus(400)
	ctx.WriteString("Unauthorized")
}

func init() {
	beego.Router("/", &controllers.MainController{})

	beego.Router("/login", &controllers.UserController{}, "post:Login")
	beego.Router("/createuser", &controllers.UserController{}, "post:CreateUser")

	// beego.InsertFilter("/secure/*", beego.BeforeExec, LoginFilter)
	// beego.Router("/get-users", &controllers.UserController{}, "get:GetAllUsers")
	// beego.Router("/secure/get-users", &controllers.UserController{}, "get:GetAllUsers")

	beego.Router("/verify/:link", &controllers.RedisController{}, "get:VerifyUser")

	beego.InsertFilter("/servercheckin", beego.BeforeExec, LoginFilter)
	beego.Router("/servercheckin", &controllers.RedisController{}, "post:ServerCheckIn")

}
