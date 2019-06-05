package controllers

import (
	"encoding/json"
	"fmt"
	"pms.cnsumi.com/models"
)

type UserController struct {
	ApiController
}

// @router /login [post]
func (c *UserController) Login() {
	u := &models.User{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &u); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}

	if u == nil || u.Name == "" || u.Password == "" {
		c.Resp.Code = models.ERR_CODE_ARGS_CHECK_FAIL
		c.Resp.Message = "Missing required args"
		return
	}

	if code, err := models.Login(u); err != nil {
		c.Resp.Code = code
		c.Resp.Message = fmt.Sprintf("登录失败: %+v", err)
	}
}

// @router /changePassword [post]
func (c *UserController) ChangePassword() {
	t := &struct {
		Name        string `json:"name"`
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}

	if code, err := models.ChangePassword(t.Name, t.OldPassword, t.NewPassword); err != nil {
		c.Resp.Code = code
		c.Resp.Message = fmt.Sprintf("密码更新失败: %+v", err)
		return
	}
}
