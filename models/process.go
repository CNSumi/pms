package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Worker struct {
	Task    *Task  // task id must ge 1
	Decoder string // use for TNGVideoTool.args

	// task about
	Channel int // channel, port, etc

	// onvif about
	OnvifArgs    []string
	OnvifPidPath string

	// TNGVideoTool
	TNGArgs        []string
	TNGPid         int
	TNGRunningFlag chan bool
	TNGRebootCount int
	TNGStartTime   time.Time
	TNGMessage     string

	Msg   string
	Lock  bool
	Index int // index at worker array, same GPU
	GPU   int // run at which gpu, number [0-8), value equal to index % 8

	logger          *log.Logger
	logWorkerTicker *time.Ticker
	logTaskTicker   *time.Ticker

	SIG_RESTART chan bool
	cancel      context.CancelFunc
	ctx         context.Context
}

var (
	workers [80]*Worker
	id2idx  = map[int64]int{}
)

func initProcess() {
	for i := 0; i < len(workers); i++ {
		sub := &Worker{}
		sub.Index = i
		sub.GPU = sub.Index % 8
		sub.logger = log.New(os.Stdout, fmt.Sprintf("[worker %d, gpu: %d]", sub.Index, sub.GPU), log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

		workers[i] = sub
		go workers[i].start() // worker never stop, until pms stop
	}

	tasks, _ := ListTask()
	for _, task := range tasks {
		if err := applyWorker(task); err != nil {
			log.Printf("apply worker fail: %+v", err)
		}
	}
}

func applyWorker(t *Task) error {
	if t.Channel == nil || *t.Channel == 0 {
		return fmt.Errorf("apply worker fail, not open channel")
	}

	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	for i := 0; i < len(workers); i++ {
		if !workers[i].Lock {
			workers[i].Lock = true
			id2idx[t.ID] = i
			workers[i].Task = t
			log.Printf("set id2idx: %d -> %d", t.ID, i)
			go workers[i].doTask()
			return nil
		}
	}
	return nil
}

func (w *Worker) doTask() {
	w.ctx, w.cancel = context.WithCancel(context.Background())

	w.SIG_RESTART = make(chan bool, 1)
	w.Channel = int(*w.Task.Channel + 1)
	w.OnvifPidPath = fmt.Sprintf("/tmp/%d.pid", w.Channel+9000)
	w.TNGRebootCount = 0
	w.logger.SetPrefix(fmt.Sprintf("[worker:%d, gpu: %d, id: %d, %s]", w.Index, w.GPU, w.Task.ID, w.Task.Name))
	for {
		w.setStreamType()

		w.startTNGVideoTool()

		w.startOnvif()

		select {
		case <-w.ctx.Done():
			w.logger.Printf("[SIGNAL] REVEIVED CANCEL")
			w.killOnvif(w.OnvifPidPath)
			w.killTNG()
			w.Lock = false
			return
		case <-w.SIG_RESTART:
			w.logger.Printf("[SIGNAL] RECEIVED RESTART")
			w.killOnvif(w.OnvifPidPath)
		}
		time.Sleep(time.Second * 3)
	}
}

func (w *Worker) killTNG() {
	if w.TNGPid == 0 {
		return
	}
	err := syscall.Kill(w.TNGPid, syscall.SIGKILL)
	w.logger.Printf("stop TNG, pid: %d, err: %+v", w.TNGPid, err)
}

func (w *Worker) startTNGVideoTool() {
	w.TNGRunningFlag = make(chan bool, 1)
	w.initTNGVideoToolArgs()
	w.TNGRebootCount++

	cmd := exec.Command(w.TNGArgs[0], w.TNGArgs[1:]...)
	_ = cmd.Start()
	w.TNGPid = cmd.Process.Pid
	w.TNGStartTime = time.Now()
	go func() {
		_, err := cmd.Process.Wait()
		w.logger.Printf("TNG crash: %+v", err)
		w.logger.Printf("[SIGNAL] SEND RESTART")
		w.SIG_RESTART <- true
	}()
}

func (w *Worker) startOnvif() {
	if !w.Task.IsONVIF {
		return
	}

	w.initOnvifArgs()
	cmd := exec.Command(w.OnvifArgs[0], w.OnvifArgs[1:]...)
	_ = cmd.Start()

	return
}

func (w *Worker) initOnvifArgs() {
	t := w.Task

	ret := []string{}
	ret = append(ret, "onvif_srvd")
	ret = append(ret, "--ifs", localNet.Name)
	ret = append(ret, "--port", fmt.Sprintf("%d", w.Channel))
	ret = append(ret, "--pid_file", w.OnvifPidPath)
	ret = append(ret, "--scope", "onvif://www.onvif.org/name/RTSPSever")
	ret = append(ret, "--scope", "onvif://www.onvif.org/Profile/S")
	ret = append(ret, "--name", "RTSPSever")
	ret = append(ret, "--width", "1920")
	ret = append(ret, "--height", "1080")
	ret = append(ret, "--url", t.RTSPAddr)
	ret = append(ret, "--type", "JPEG")

	w.OnvifArgs = ret
}

func (w *Worker) initTNGVideoToolArgs() {
	t := w.Task

	ret := []string{}
	ret = append(ret, "TNGVideoTool")
	ret = append(ret, "--prefix", t.Name)
	ret = append(ret, "--rand", "30")
	ret = append(ret, "-hide_banner")
	ret = append(ret, "-loglevel", "warning")
	ret = append(ret, "-stimeout", "3000000")
	ret = append(ret, "-rtsp_transport", t.RTSPTransPort)
	ret = append(ret, "-hwaccel", "cuvid")
	ret = append(ret, "-vcodec", fmt.Sprintf("%s_cuvid", w.Decoder))
	ret = append(ret, "-hwaccel_device", fmt.Sprintf("%d", w.GPU))
	ret = append(ret, "-GPU", fmt.Sprintf("%d", w.GPU))
	ret = append(ret, "-i", t.RTSPAddr)
	ret = append(ret, "-f", "rtsp")
	ret = append(ret, "-rtsp_transport", t.RTSPTransPort)
	ret = append(ret, "-g", fmt.Sprintf("%d", *t.GOP))
	ret = append(ret, "-b:v", t.BitRateV)
	ret = append(ret, "-zerolatency", "1")
	ret = append(ret, "-vcodec", fmt.Sprintf("%s_nevnc", t.Encoder))
	ret = append(ret, "-profile:v", t.Profile)
	ret = append(ret, "-GPU", fmt.Sprintf("%d", w.GPU))
	ret = append(ret, "-acodec", "aac")
	ret = append(ret, "-b:a", t.BitRateA)
	ret = append(ret, fmt.Sprintf("rtsp://127.0.0.1/%d", *t.Channel+1))

	w.TNGArgs = ret
}

func (w *Worker) setStreamType() {
	for {
		content, _ := execCommand("getStreamType", w.Task.RTSPAddr)
		w.logger.Printf("[%s]getStreamType content: %s", w.Task.RTSPAddr, content)
		matches := getStreamTypeRegex.FindStringSubmatch(content)
		if len(matches) == 2 && (matches[1] == "hevc" || matches[1] == "h264") {
			w.Decoder = matches[1]
			w.logger.Printf("set decoder: %s", w.Decoder)
			return
		}
		tick := time.NewTicker(time.Second * 3)
		select {
		case <-w.ctx.Done():
			return
		case <-tick.C:
			w.logger.Printf("getStreamType again")
		}
	}

}

func (w *Worker) start() {
	w.killOnvif(fmt.Sprintf("/tmp/%d.pid", 9001+w.Index))

	go w.log()
}

func (w *Worker) killOnvif(path string) {
	content, _ := execCommand("cat", path)
	pid, _ := strconv.ParseInt(content, 10, 64)
	if pid <= 0 {
		return
	}

	err := syscall.Kill(int(pid), syscall.SIGKILL)
	w.logger.Printf("stop onvif(%d), pid: %d, err: %+v", w.Index, pid, err)
}

func (w *Worker) log() {
	w.logWorkerTicker = time.NewTicker(time.Second * 60)
	w.logTaskTicker = time.NewTicker(time.Second * 3)
	for {
		select {
		case <-w.logWorkerTicker.C:
			msg := "waiting for task"
			if w.Lock {
				msg = fmt.Sprintf("id: %d, name: %s", w.Task.ID, w.Task.Name)
			}
			w.logger.Printf(msg)
		case <-w.logTaskTicker.C:
			if w.Lock {
				w.logTNG()
			}
		}
	}
}

func (w *Worker) logTNG() {
	msg := fmt.Sprintf("TNG.Pid: %d, reboot: %d, start: %+v", w.TNGPid, w.TNGRebootCount, w.TNGStartTime)
	w.logger.Printf(msg)
}
