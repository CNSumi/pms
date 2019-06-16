package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	GPU_COUNT = 8
)

var (
	procs = [80]*Proc{}
	gpus  = [GPU_COUNT]uint8{}
	clearOnvifFlag = make(chan bool, 1)
)

type Proc struct {
	Task    *Task
	GPU     uint8
	Decoder string

	OnvifArgs []string
	OnvifPid  int

	TNGVideoToolArgs        []string
	TNGVideoToolPid         int
	TNGVideoToolRebootCount int

	ctx    context.Context
	cancel context.CancelFunc
	logger *log.Logger
}

func nextGPU() uint8 {
	index := uint8(0)
	for i := uint8(0); i < GPU_COUNT; i++ {
		if gpus[i] < gpus[index] {
			index = i
		}
	}
	gpus[index]++
	return index
}

func (p *Proc) startTNGVideoTool() {
	t := p.Task
	for {
		decoder, err := getStreamType(t.RTSPAddr)
		if err != nil {
			p.logger.Printf("getStreamType err: %+v, waiting 3 seconds to restart or return", err)
			tick := time.NewTicker(time.Second * 3)
			select {
			case <-p.ctx.Done():
				return
			case <-tick.C:
			}
			continue
		}
		p.Decoder = decoder
		p.makeTNGVideoToolArgs()
		p.TNGVideoToolRebootCount++

		cmd := exec.Command(p.TNGVideoToolArgs[0], p.TNGVideoToolArgs[1:]...)
		if err := cmd.Start(); err != nil {
			p.logger.Printf("start TNGVideoTool fail: %+v, waiting 3 seconds to restart", err)
			time.Sleep(time.Second * 3)
			continue
		}
		p.TNGVideoToolPid = cmd.Process.Pid

		_, _ = cmd.Process.Wait()
		p.logger.Printf("TNGVideoTool crash, waiting 3 seconds to reboot or exit")
		select {
		case <-p.ctx.Done():
			return
		case <-time.Tick(time.Second * 3):

		}
	}
}

func startTask(t *Task) error {
	if t == nil || t.Channel == nil || *t.Channel <= 0 || *t.Channel >= 80 {
		return fmt.Errorf("empty Task")
	}

	if procs[*t.Channel] != nil {
		return fmt.Errorf("task already running")
	}

	proc := &Proc{}
	proc.Task = t
	proc.GPU = nextGPU()
	proc.logger = log.New(os.Stdout, fmt.Sprintf("[id: %d, channel: %d, gpu: %d, name: %s]", t.ID, *t.Channel, proc.GPU, t.Name), log.Ldate|log.Ltime|log.Lshortfile)
	proc.ctx, proc.cancel = context.WithCancel(context.Background())
	procs[*t.Channel] = proc

	go detail(*t.Channel)

	go proc.startOnvif()

	go proc.startTNGVideoTool()

	go func() {
		select {
		case <-proc.ctx.Done():
			if err := killOnvif(*proc.Task.Channel); err != nil {
				proc.logger.Printf("stop onvif fail: %+v", err)
			}

			if err := kill(proc.TNGVideoToolPid, "onvif"); err != nil {
				proc.logger.Printf("kill TNGVideoToolPid(%d) fail", proc.TNGVideoToolPid)
			}
		}
	}()

	return nil
}

func stopTask(t *Task) error {
	if t == nil || t.Channel == nil || *t.Channel == 0 || *t.Channel >= 80 {
		return fmt.Errorf("err stop Task")
	}
	proc := procs[*t.Channel]
	if proc == nil || proc.Task == nil || *proc.Task.Channel != *t.Channel {
		return fmt.Errorf("stop err")
	}
	proc.cancel()
	gpus[proc.GPU]--
	procs[*t.Channel] = nil
	return nil
}

func init() {
	for i := uint16(1); i <= 80; i++ {
		killOnvif(i)
	}

	go func() {
		<-qs_task_initAlreadyFlag
		log.Printf("received db init already")
		tasks, err := ListTask()
		if err != nil {
			log.Fatalf(fmt.Sprintf("init list Task fail: %+v", err))
		}
		for _, t := range tasks {
			if err := startTask(t); err != nil {
				log.Printf("start proce fail: %+v", err)
			}
		}
	}()
}

func (p *Proc) makeTNGVideoToolArgs() {
	t := p.Task

	ret := []string{}
	ret = append(ret, "TNGVideoTool")
	ret = append(ret, "--prefix", t.Name)
	ret = append(ret, "--rand", "100000")
	ret = append(ret, "-hide_banner")
	ret = append(ret, "-loglevel", "warning")
	ret = append(ret, "-stimeout", "3000000")
	ret = append(ret, "-rtsp_transport", t.RTSPTransPort)
	ret = append(ret, "-hwaccel", "cuvid")
	ret = append(ret, "-vcodec", fmt.Sprintf("%s_cuvid", p.Decoder))
	ret = append(ret, "-hwaccel_device", fmt.Sprintf("%d", p.GPU))
	ret = append(ret, "-GPU", fmt.Sprintf("%d", p.GPU))
	ret = append(ret, "-i", t.RTSPAddr)
	ret = append(ret, "-f", "rtsp")
	ret = append(ret, "-rtsp_transport", t.RTSPTransPort)
	ret = append(ret, "-g", fmt.Sprintf("%d", *t.GOP))
	ret = append(ret, "-b:v", t.BitRateV)
	ret = append(ret, "-zerolatency", "1")
	ret = append(ret, "-vcodec", fmt.Sprintf("%s_nevnc", t.Encoder))
	ret = append(ret, "-profile:v", t.Profile)
	ret = append(ret, "-GPU", fmt.Sprintf("%d", p.GPU))
	ret = append(ret, "-acodec", "aac")
	ret = append(ret, "-b:a", t.BitRateA)
	ret = append(ret, t.RTSPAddr)

	p.TNGVideoToolArgs = ret
}

func getStreamType(addr string) (string, error) {
	content, _ := execCommand("getStreamType", addr)
	log.Printf("addr: %s", addr)
	matcher := getStreamTypeRegex.FindStringSubmatch(content)
	if len(matcher) == 2 {
		if matcher[1] == "hevc" || matcher[1] == "h264" {
			return matcher[1], nil
		}
	}
	log.Printf("content: %s", content)
	return "", fmt.Errorf("get StreamType fail")
}

func (p *Proc) startOnvif() {
	if !p.Task.IsONVIF {
		p.logger.Printf("not onvif proc, return")
		return
	}

	if len(p.OnvifArgs) == 0 {
		p.makeOnvifArgs()
	}

	cmd := exec.Command(p.OnvifArgs[0], p.OnvifArgs[1:]...)
	if err := cmd.Start(); err != nil {
		p.logger.Printf("start onvif fail, %+v, reboot", err)
		time.Sleep(time.Second * 3)
		return
	}

	p.OnvifPid = cmd.Process.Pid

	_, _ = cmd.Process.Wait()
}

func (p *Proc) makeOnvifArgs() {
	t := p.Task

	num := 9000 + *t.Channel
	ret := []string{}
	ret = append(ret, "onvif_srvd")
	ret = append(ret, "--ifs", localNet.Name)
	ret = append(ret, "--port", fmt.Sprintf("%d", num))
	ret = append(ret, "--pid_file", fmt.Sprintf("/tmp/%d.pid", num))
	ret = append(ret, "--scope", "onvif://www.onvif.org/name/RTSPSever")
	ret = append(ret, "--scope", "onvif://www.onvif.org/Profile/S")
	ret = append(ret, "--name", "RTSPSever")
	ret = append(ret, "--width", "1920")
	ret = append(ret, "--height", "1080")
	ret = append(ret, "--url", t.RTSPAddr)
	ret = append(ret, "--type", "JPEG")

	p.OnvifArgs = ret
}

func detail(channel uint16) {
	tick := time.NewTicker(time.Second * 10)
	for {
		p := procs[channel]
		if p == nil {
			log.Printf("channel: %d 's detail exit by proc is nil", channel)
			return
		}
		logger := p.logger
		if logger == nil {
			log.Printf("channel: %d 's detail exit by proc logger is nil", channel)
			return
		}

		select {
		case <-p.ctx.Done():
			p.logger.Printf("log stop by cancel")
			return
		case <-tick.C:
			p.logger.Printf("[onvif]. pid: %d", p.OnvifPid)
			p.logger.Printf("[TNGVideoTool]. [%d]pid: %d", p.TNGVideoToolRebootCount, p.TNGVideoToolPid)
		}
	}
}

func kill(pid int, name string) error {
	if pid == 0	{
		return fmt.Errorf("%s(%d) not running", name, pid)
	}

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("kill %s(%d) fail: %+v", name, pid, err)
	}
	return nil
}

func killOnvif(channel uint16) error {
	text := fmt.Sprintf("kill -9 `cat /tmp/%d.pid`", channel + 9000)
	log.Printf("stop onvif(%d), cmd: %s", channel, text)

	cmd := exec.Command("/bin/sh", "-c", text)
	return cmd.Start()
}