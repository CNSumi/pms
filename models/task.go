package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"log"
)

var (
	zero	= uint16(0)
	qs_task orm.QuerySeter
	qs_task_initAlreadyFlag = make(chan bool, 1)

	encoder2profile = map[string][]string{
		"H264": {"main", "baseline", "high"},
		"HEVC": {"main"},
	}
)

func init() {
	<- initDBFlag
	close(initDBFlag)
	qs_task = o.QueryTable("pms_task").OrderBy("id")
	qs_task_initAlreadyFlag <- true
}

type Task struct {
	ID int64 `json:"id" orm:"pk;column(id);auto"`

	Name          string  `json:"name" orm:"column(name);unique"`
	BitRateV      string  `json:"bitrate_v" orm:"column(bitrate_v)"`
	BitRateA      string  `json:"bitrate_a" orm:"column(bitrate_a)"`
	FPS           *uint64 `json:"fps" orm:"column(fps)"`
	GOP           *uint64 `json:"gop" orm:"column(gop)"`
	Encoder       string  `json:"encoder"`
	Profile       string  `json:"profile"`
	RTSPTransPort string  `json:"rtsp_transport" orm:"column(rtsp_transport)"`
	RTSPAddr      string  `json:"rtsp_addr" orm:"column(rtsp_addr)"`
	ONVIF_IP      string  `json:"onvif_ip;omitempty" orm:"column(onvif_ip)"`
	ONVIF_user    string  `json:"onvif_user;omitempty" orm:"column(onvif_user)"`
	ONVIF_pwd     string  `json:"onvif_pwd;omitempty" orm:"column(onvif_pwd)"`
	Channel       *uint16 `json:"channel" orm:"column(channel);null;unique"`
	IsONVIF       bool    `json:"-" orm:"column(is_onvif)"`
}

func ListTask() ([]*Task, error) {
	tasks := []*Task{}

	if _, err := qs_task.All(&tasks); err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.Channel == nil {
			task.Channel = &zero
		}
	}
	return tasks, nil
}

func (t *Task) selfCheck() error {
	if err := t.checkName(); err != nil {
		return err
	}

	if err := t.checkBitRateVA(); err != nil {
		return err
	}

	if err := t.checkFPSAndGOP(); err != nil {
		return err
	}

	if err := t.checkEncoderAndProfile(); err != nil {
		return err
	}

	if err := t.checkRTSPTransPort(); err != nil {
		return err
	}

	if err := t.checkRTSPAddr(); err != nil {
		return err
	}

	if err := t.checkONVIF(); err != nil {
		return err
	}

	if err := t.checkChannel(); err != nil {
		return err
	}

	return nil
}

func (t *Task) checkName() error {
	if t.Name == "" {
		return fmt.Errorf("配置名不能为空")
	}

	return nil
}

func (t *Task) checkBitRateVA() error {
	if t.BitRateV == "" || t.BitRateA == "" {
		return fmt.Errorf("bitRateV/A不能为空")
	}
	return nil
}

func (t *Task) checkFPSAndGOP() error {
	if t.FPS == nil || t.GOP == nil {
		return fmt.Errorf("fps/gop不能为空")
	}
	return nil
}

func (t *Task) checkEncoderAndProfile() error {
	encoder2profile = map[string][]string{
		"H264": {"main", "baseline", "high"},
		"HEVC": {"main"},
	}

	if ps, ok := encoder2profile[t.Encoder]; ok {
		for _, profile := range ps {
			if profile == t.Profile {
				return nil
			}
		}
	}
	return fmt.Errorf("invalid encoder(%s) with profile(%s)", t.Encoder, t.Profile)
}

func (t *Task) checkRTSPTransPort() error {
	v := t.RTSPTransPort
	if v == "tcp" || v == "udp" {
		return nil
	}
	return fmt.Errorf("非法RTSPTransPort(%s)", t.RTSPTransPort)
}

func (t *Task) checkRTSPAddr() error {
	if t.RTSPAddr == "" {
		return fmt.Errorf("rtsp_addr不能为空: %s", t.RTSPAddr)
	}
	return nil
}

func (t *Task) checkONVIF() error {
	if t.ONVIF_IP == "" || t.ONVIF_user == "" || t.ONVIF_pwd == "" {

	}
	t.IsONVIF = true

	return nil
}

func (t *Task) checkRTSPAndONVIF() error {
	if t.RTSPAddr == "" {
		return fmt.Errorf("rtsp地址不能为空")
	}
	return nil
}

func (t *Task) checkChannel() error {
	if t.Channel == nil {
		return fmt.Errorf("channel cannot empty")
	}
	if *t.Channel > 80 {
		return fmt.Errorf("channel cannot > 80")
	}
	if *t.Channel == 0 {
		t.Channel = nil
	}
	return nil
}

func AddConfig(t *Task) (int64, error) {
	if err := t.selfCheck(); err != nil {
		return 0, fmt.Errorf("check fail: %+v", err)
	}

	id, err := o.Insert(t)
	if err != nil {
		return 0, fmt.Errorf("add config fail: %+v", err)
	}

	if t.Channel != nil {
		if err := startTask(t); err != nil {
			log.Printf("start process fail: %+v", err)
		}
	}

	return id, nil
}

type UpdateRet struct {
	ID		int64	`json:"id"`
	Code	int	`json:"code"`
	Message	string	`json:"message"`
}

func UpdateConfig(t []*Task) []*UpdateRet {
	ret := []*UpdateRet{}

	for _, tmp := range t {
		if tmp == nil || tmp.ID == 0 {
			continue
		}

		code, message := updateSingle(tmp)

		ret = append(ret, &UpdateRet{
			ID: tmp.ID,
			Code: code,
			Message: message,
		})
	}

	return ret
}

func updateSingle(t *Task) (int, string) {
	if t == nil {
		return -1, fmt.Sprintf("Task is empty")
	}
	if err := t.selfCheck(); err != nil {
		return -3, fmt.Sprintf("self check fail: %s", err.Error())
	}
	flag, err := o.Update(t)
	if err != nil {
		return -4, fmt.Sprintf("exec update fail: %+v", err)
	}
	if flag != 1 {
		return -5, fmt.Sprintf("pk(%d) not exists", t.ID)
	}

	if err := startTask(t); err != nil {
		log.Printf("start task tail: %+v", err)
	}


	return 0, "ok"
}

func RemoveTask(id int64) error {
	t := &Task{
		ID: id,
	}
	if err := o.Read(t, "id"); err != nil {
		return fmt.Errorf("read config fail: %+v", err)
	}

	if err := stopTask(t); err != nil {
		log.Printf("%d stop fail: %+v", t.Channel, err)
	}

	if _, err := o.Delete(t, "id"); err != nil {
		return err
	}

	return nil
}