package controllers

import (
	"github.com/astaxie/beego"
	"pms.cnsumi.com/models"
)

type ApiController struct {
	beego.Controller

	Resp *ApiResp
}

func (c *ApiController) Prepare() {
	c.EnableRender = false
	c.Resp = &ApiResp{}
	c.Resp.Code = models.ERR_CODE_OK
	c.Resp.Message = "ok"
}

func (c *ApiController) Finish() {
	if c.Resp == nil {
		c.Resp = &ApiResp{
			Code:    models.ERR_CODE_UNKNOWN_ERROR,
			Message: "未知错误",
		}
	}
	c.Data["json"] = c.Resp
	c.ServeJSON(true)
}

type ApiResp struct {
	Code    models.ERR_CODE `json:"code"`
	Message string          `json:"message"`
	Data    interface{}     `json:"data,omitempty"`
}
