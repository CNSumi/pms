package controllers

import (
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

// @router /rebootabc [post]
func (c *SystemController) Reboot() {
	err := models.Reboot()
	if err != nil {
		c.Resp.Code = models.ERR_CODE_CMD_EXEC_FAIL
		c.Resp.Message = fmt.Sprintf("reboot fail: %+v", err)
	}
}