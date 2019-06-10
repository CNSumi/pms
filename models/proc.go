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
				log.Printf("exit by break signal")
				break
			}

			proc.RebootCount++
			args := makeArgs(proc)
			cmd := exec.Command(args[0], args[1:]...)

			err := cmd.Start()
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
				log.Fatalf("start proce fail: %+v", err)
			}
		}
	}()
}
/**
-hwaccel cuvid
-vcodec hevc_cuvid
-hwaccel_device 0
-GPU 0
-i rtsp://admin:admin12345@192.168.1.12
-f rtsp
-rtsp_transport tcp
-g 250
-b:v 500k
-zerolatency 1
-profile:v high
-vcodec h264_nvenc
-GPU 0
rtsp://192.168.1.227/main
 */
func makeArgs(p *Proc) []string {
	t := p.Task

	ret := []string{}
	ret = append(ret, "TNGVideoTool")
	ret = append(ret, "--prefix", t.Name)
	ret = append(ret, "--rand", "5")
	ret = append(ret, "-hide_banner")
	ret = append(ret, "-loglevel", "warning")
	ret = append(ret, "-stimeout", "3000000")
	ret = append(ret, "-rtsp_transport", t.RTSPTransPort)
	ret = append(ret, "-hwaccel", "cuvid")
	ret = append(ret, "-vcodec", t.Encoder)
	ret = append(ret, "-hwaccel_device", fmt.Sprintf("%d", p.GPU))
	ret = append(ret, "-GPU", fmt.Sprintf("%d", p.GPU))
	ret = append(ret, "-i", t.RTSPAddr)
	ret = append(ret, "-f", "rtsp")
	ret = append(ret, "-rtsp_transport", t.RTSPTransPort)
	ret = append(ret, "-g", fmt.Sprintf("%d", *t.GOP))
	ret = append(ret, "-b:v", t.BitRateV)
	ret = append(ret, "-zerolatency", "1")
	ret = append(ret, "-profile:v", t.Profile)
	ret = append(ret, "-vcodec", t.Encoder)
	ret = append(ret, "-GPU", fmt.Sprintf("%d", p.GPU))
	ret = append(ret, "-rtsp", t.RTSPAddr)

	return ret
}