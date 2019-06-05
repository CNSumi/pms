package controllers

import (
	"encoding/json"
	"fmt"
	"pms.cnsumi.com/models"
)

const (
	count_per_page = 20
)

type TaskController struct {
	ApiController
}

// @router /listConfig [post]
func (c *TaskController) List() {
	t := &struct {
		Page	uint32	`json:"page"`
		Count	uint32	`json:"count"`
	}{}
	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &t)
	if t.Count == 0 {
		t.Count = count_per_page
	}

	tasks, err := models.ListTask()
	if err != nil {
		c.Resp.Code = models.ERR_CODE_DB_QUERY_FAIL
		c.Resp.Message = fmt.Sprintf("query task fail: %+v", err)
		return
	}
	hasMore := false
	if uint32(len(tasks)) > (t.Page + 1) * t.Count {
		hasMore = true
		tasks = tasks[t.Page * t.Count: (t.Page + 1) * t.Count]
	}

	c.Resp.Data = &struct {
		HasMore	bool	`json:"hasMore"`
		List	[]*models.Task	`json:"list"`
	}{
		HasMore: hasMore,
		List: tasks,
	}
}

// @router /newConfig [post]
func (c *TaskController) NewTask() {
	t := &models.Task{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}

	id, err := models.AddConfig(t)
	if err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = err.Error()
		return
	}

	c.Resp.Data = id
}

// @router /updateConfig [post]
func (c *TaskController) UpdateTask() {
	t := &models.Task{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}

	if t.ID == 0 {
		c.Resp.Code = models.ERR_CODE_ARGS_MISS_REQUIRED
		c.Resp.Message = fmt.Sprintf("missing pk(id)")
		return
	}

	if err := models.UpdateConfig(t); err != nil {
		c.Resp.Code = models.ERR_CODE_UNKNOWN_ERROR
		c.Resp.Message = fmt.Sprintf("update config fail: %+v", err)
		return
	}
}

// @router /removeConfig [post]
func (c *TaskController) RemoveConfig() {
	t := &struct {
		ID	int64	`json:"id"`
	}{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}
	if err := models.RemoveTask(t.ID); err != nil {
		c.Resp.Code = models.ERR_CODE_DB_QUERY_FAIL
		c.Resp.Message = fmt.Sprintf("remove config fail: %+v", err)
		return
	}
}

// @router /getRTSPAddress [post]
func (c *TaskController) GetRTSP() {
	t := &struct {
		Host	string	`json:"host"`
		User	string	`json:"user"`
		Password	string	`json:"password"`
	}{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}

	addr, err := models.GetRTSPAddress(t.Host, t.User, t.Password)
	if err != nil {
		c.Resp.Code = models.ERR_CODE_GET_RTSP_FAIL
		c.Resp.Message = fmt.Sprintf("get rtsp addr fail: %+v", err)
		return
	}
	c.Resp.Data = addr
}