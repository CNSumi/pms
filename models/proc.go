package models

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

const (
	GPU_COUNT = 8
)

var (
	procs = [80]*Proc{}
	gpus = [GPU_COUNT]uint8{}
)

func init() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			ret := fmt.Sprintf("\n----------------------GPU---------------------\n")
			for _, count := range gpus {
				ret += fmt.Sprintf("|")
				for i := uint8(0); i < count; i++ {
					ret += "*"
				}
				ret += "\n"
			}
			ret += fmt.Sprintf("----------------------GPU---------------------")
			log.Printf("%s", ret)
		}
	}()
}

type Proc struct {
	Task         *Task
	RebootCount  uint64
	GPU          uint8
	Pid          uint32
	OnvifPid	 uint32
	Decoder			string

	OnvifArgs	[]string

	stopSignal   chan bool
	rebootSignal chan bool
	isBreak      bool

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

func startTask(t *Task) error {
	if t == nil || t.Channel == nil || *t.Channel <= 0 || *t.Channel >= 80 {
		return fmt.Errorf("empty Task")
	}

	if procs[*t.Channel] != nil {
		return fmt.Errorf("Task already running")
	}

	proc := &Proc{}
	proc.Task = t
	proc.stopSignal = make(chan bool, 1)
	proc.rebootSignal = make(chan bool, 1)
	proc.GPU = nextGPU()
	procs[*t.Channel] = proc

	go func() {
		for {
			log.SetPrefix(fmt.Sprintf("[%d-%d-%d,%d:%s]", t.ID, *t.Channel, proc.GPU, proc.RebootCount, t.Name))
			if proc.isBreak {
				procs[*proc.Task.Channel] = nil
				log.Printf("exit by stop(break) signal")
				break
			}

			decoder, err := getStreamType(t.RTSPAddr)
			if err != nil {
				log.Printf("getStreamType err: %+v", err)
				time.Sleep(time.Second)
				continue
			}
			proc.Decoder = decoder

			proc.RebootCount++
			args := makeArgsTrans(proc)
			cmd := exec.Command(args[0], args[1:]...)

			err = cmd.Start()
			if err != nil {
				log.Printf("cmd start fail: %+v", err)
				time.Sleep(time.Second * 3)
				continue
			}
			log.Printf("running at Pid: %d", cmd.Process.Pid)

			go func() {
				_, _ = cmd.Process.Wait()
				proc.rebootSignal <- true
			}()

			select {
			case <- proc.stopSignal:
				proc.isBreak = true
				log.Printf("set stop signal")
			case <- proc.rebootSignal:
				log.Printf("reboot by crash")
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
	procs[*t.Channel].stopSignal <- true
	gpus[proc.GPU]--
	return nil
}

func init() {
	go func() {
		<- qs_task_initAlreadyFlag
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

func makeArgsTrans(p *Proc) []string {
	t := p.Task

	ret := []string{}
	ret = append(ret, "TNGVideoTool")
	ret = append(ret, "--prefix", t.Name)
	ret = append(ret, "--rand", "50")
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

	return ret
}

func getStreamType(addr string) (string, error) {
	content, _ := execCommand("getStreamType", "--rtsp", addr)
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
	if len(p.OnvifArgs) == 0 {
		p.makeOnvifArgs()
	}

}

func (p *Proc) makeOnvifArgs() {
	t := p.Task

	num := 9000 + *t.Channel
	ret := []string{}
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