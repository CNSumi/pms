package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"],
        beego.ControllerComments{
            Method: "AppVersion",
            Router: `/getAppVersion`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"],
        beego.ControllerComments{
            Method: "SystemInfo",
            Router: `/getStatus`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"],
        beego.ControllerComments{
            Method: "Network",
            Router: `/network`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:SystemController"],
        beego.ControllerComments{
            Method: "Reboot",
            Router: `/rebootabc`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"],
        beego.ControllerComments{
            Method: "GetRTSP",
            Router: `/getRTSPAddress`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"],
        beego.ControllerComments{
            Method: "List",
            Router: `/listConfig`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"],
        beego.ControllerComments{
            Method: "NewTask",
            Router: `/newConfig`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"],
        beego.ControllerComments{
            Method: "RemoveConfig",
            Router: `/removeConfig`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:TaskController"],
        beego.ControllerComments{
            Method: "UpdateTask",
            Router: `/updateConfig`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:UserController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:UserController"],
        beego.ControllerComments{
            Method: "ChangePassword",
            Router: `/changePassword`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["pms.cnsumi.com/controllers:UserController"] = append(beego.GlobalControllerRouter["pms.cnsumi.com/controllers:UserController"],
        beego.ControllerComments{
            Method: "Login",
            Router: `/login`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
