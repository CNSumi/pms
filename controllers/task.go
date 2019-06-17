package controllers

import (
	"encoding/json"
	"fmt"
	"pms.cnsumi.com/models"
)

const (
	count_per_page = 5
)

type TaskController struct {
	ApiController
}

// @router /listConfig [post]
func (c *TaskController) List() {
	t := &struct {
		Page	int	`json:"page"`
	}{}
	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &t)
	if t.Page < 0 {
		c.Resp.Code = models.ERR_CODE_ARGS_CHECK_FAIL
		c.Resp.Message = fmt.Sprintf("page cannot < 0")
		return
	}

	tasks, err := models.ListTask()
	if err != nil {
		c.Resp.Code = models.ERR_CODE_DB_QUERY_FAIL
		c.Resp.Message = fmt.Sprintf("query task fail: %+v", err)
		return
	}
	count := len(tasks)
	start := t.Page * count_per_page
	end := start + count_per_page
	if len(tasks) < start {
		c.Resp.Code = models.ERR_CODE_ARGS_CHECK_FAIL
		c.Resp.Message = fmt.Sprintf("err page, current max page: %d", len(tasks) / count_per_page)
		return
	}
	hasMore := false
	if len(tasks) > (t.Page + 1) * count_per_page {
		hasMore = true
	} else {
		end = len(tasks)
	}

	tasks = tasks[start: end]

	c.Resp.Data = &struct {
		HasMore	bool	`json:"hasMore"`
		List	[]*models.Task	`json:"list"`
		Count 	int	`json:"count"`
	}{
		HasMore: hasMore,
		List: tasks,
		Count: count,
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
	t := []*models.Task{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &t); err != nil {
		c.Resp.Code = models.ERR_CODE_BODY_DECODE_FAIL
		c.Resp.Message = fmt.Sprintf("body decode fail: %+v", err)
		return
	}

	if len(t) == 0 {
		c.Resp.Code = models.ERR_CODE_ARGS_MISS_REQUIRED
		c.Resp.Message = fmt.Sprintf("update task.count == 0")
		return
	}

	ret := []*models.UpdateRet{}
	for _, t := range t {
		sub := models.UpdateConfig(t)
		if sub != nil {
			ret = append(ret, sub)
		}
	}
	c.Resp.Data = ret
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