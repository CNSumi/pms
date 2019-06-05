package main

import (
	"github.com/astaxie/beego"
	_ "pms.cnsumi.com/models"
	_ "pms.cnsumi.com/routers"
)

func main() {
	beego.Run()
}
