package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"strings"
	"time"
)

var (
	zero	= uint16(0)
	qs_task orm.QuerySeter

	encoder2profile = map[string][]string{
		"H264": {"main", "baseline", "high"},
		"HEVC": {"main"},
	}
)

func init() {
	qs_task = o.QueryTable("pms_task").OrderBy("id")
}

type Task struct {
	ID int64 `json:"id" orm:"pk;column(id);auto"`

	Name          string  `json:"name" orm:"column(name);unique"`
	BitRateV      string  `json:"bitrate_v" orm:"column(bitrate_v)"`
	BitRateA      string  `json:"bitrate_a" orm:"column(bitrate_a)"`
	FPS           *uint64 `json:"fps" orm:"column(fps)"`
	GOP           *uint64 `json:"gop" orm:"column(gop)"`
	Encoder       string  `json:"encoder" orm:"column(encoder)"`
	Profile       string  `json:"profile" orm:"column(profile)"`
	RTSPTransPort string  `json:"rtsp_transport" orm:"column(rtsp_transport)"`
	RTSPAddr      string  `json:"rtsp_addr" orm:"column(rtsp_addr)"`
	ONVIF_IP      string  `json:"onvif_ip;omitempty" orm:"column(onvif_ip)"`
	ONVIF_user    string  `json:"onvif_user;omitempty" orm:"column(onvif_user)"`
	ONVIF_pwd     string  `json:"onvif_pwd;omitempty" orm:"column(onvif_pwd)"`
	Channel       *uint16 `json:"channel" orm:"column(channel);null;unique"`
	IsONVIF       bool    `json:"-" orm:"column(is_onvif)"`

	Created		time.Time	`json:"created" orm:"auto_now_add;type(datetime)"`
	Updated		time.Time	`json:"updated" orm:"auto_now;type(datetime)"`
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
	err := o.Read(t, "name")
	if err != nil {
		if err == orm.ErrNoRows {
			return nil
		}
		return fmt.Errorf("未知错误: %+v", err)
	}
	return fmt.Errorf("此配置名已存在(%s)", t.Name)
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
	if t.ONVIF_IP == "" && t.ONVIF_user == "" && t.ONVIF_pwd == "" {
		return nil
	}
	t.IsONVIF = true

	if !isIPV4Addr(t.ONVIF_IP) {
		return fmt.Errorf("非法ONVIF_IP(%s)", t.ONVIF_IP)
	}
	if t.ONVIF_user == "" {
		return fmt.Errorf("ONVIF_USER不能为空(%s)", t.ONVIF_user)
	}
	if t.ONVIF_pwd == "" {
		return fmt.Errorf("ONVIF_PWD不能为空(%s)", t.ONVIF_pwd)
	}

	if strings.Index(t.RTSPAddr, fmt.Sprintf("rtsp://%s:%s@%s", t.ONVIF_user, t.ONVIF_pwd, t.ONVIF_IP)) != 0 {
		return fmt.Errorf("onvif配置与rtsp地址不符")
	}
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

	return id, nil
}

func UpdateConfig(t *Task) error {
	if err := t.selfCheck(); err != nil {
		return err
	}

	flag, err := o.Update(t)
	if err != nil {
		return err
	}
	if flag != 1 {
		return fmt.Errorf("error id(%d))", t.ID)
	}

	return nil
}

func RemoveTask(id int64) error {
	t := &Task{
		ID: id,
	}
	if err := o.Read(t, "id"); err != nil {
		return fmt.Errorf("read config fail: %+v", err)
	}
	if t.Channel != nil {
		return fmt.Errorf("cannot remove running config")
	}
	if _, err := o.Delete(t, "id"); err != nil {
		return err
	}
	return nil
}