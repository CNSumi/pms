package routers

import (
	"github.com/astaxie/beego"
	"pms.cnsumi.com/controllers"
)

func init() {
	ns := beego.NewNamespace("/api",
		beego.NSInclude(&controllers.UserController{}),
		beego.NSInclude(&controllers.SystemController{}),
		beego.NSInclude(&controllers.TaskController{}),
	)

	beego.AddNamespace(ns)
}
