package controllers

import (
	"encoding/json"
	"fmt"
	"pms.cnsumi.com/models"
)

type SystemController struct {
	ApiController
}

// @router /getStatus [post]
func (c *SystemController) SystemInfo() {
	c.Resp.Data = models.GetSystemStat()
}

// @router /getAppVersion [post]
func (c *SystemController) AppVersion() {
	c.Resp.Data = models.AppVersion()
}

// @router /network [post]
func (c *SystemController) Network() {
	c.Resp.Data = models.LocalNet()
}

// @router /reboot [post]
func (c *SystemController) Reboot() {
	//err := models.Reboot()
}

// @router /setNetwork [post]
func (c *SystemController) SetNetwork() {
	t := &struct {
		INet      string `json:"inet"`
		Mask      string `json:"mask"`
		Broadcast string `json:"broadcast"`
	}{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}
	if err := models.SetNetwork(t.INet, t.Mask, t.Broadcast); err != nil {
		c.Resp.Code = models.ERR_CODE_CMD_EXEC_FAIL
		c.Resp.Message = fmt.Sprintf("set network fail: %+v", err)
		return
	}
}
