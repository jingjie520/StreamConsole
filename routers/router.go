package routers

import (
	"github.com/astaxie/beego"
	"streamConsole/controllers"
)

func init() {
	// 默认进入后台
	beego.Router("/", &controllers.HomeController{}, "*:Index")

	beego.Router("/login", &controllers.LoginController{}, "*:LoginIn")
	beego.Router("/login_out", &controllers.LoginController{}, "*:LoginOut")
	beego.Router("/no_auth", &controllers.LoginController{}, "*:NoAuth")

	beego.Router("/home", &controllers.HomeController{}, "*:Index")
	beego.Router("/home/start", &controllers.HomeController{}, "*:Start")

	beego.AutoRouter(&controllers.GuardController{})
	beego.AutoRouter(&controllers.SerialController{})
	beego.AutoRouter(&controllers.AuthController{})
	beego.AutoRouter(&controllers.RoleController{})
	beego.AutoRouter(&controllers.AdminController{})
	beego.AutoRouter(&controllers.UserController{})
	beego.AutoRouter(&controllers.ChannelController{})

}
