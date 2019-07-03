package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Worker struct {
	Task    *Task  // task id must ge 1

	// task about
	Channel int // channel, port, etc

	// onvif about
	OnvifArgs    string
	OnvifPidPath string
	OnvifPPid	int

	// TNGVideoTool
	TNGArgs        string
	TNGPid         int
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
	w.Channel = int(*w.Task.Channel)
	w.TNGRebootCount = 0
	w.logger.SetPrefix(fmt.Sprintf("[worker:%d, gpu: %d, id: %d, %s]", w.Index, w.GPU, w.Task.ID, w.Task.Name))
	for {
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
	w.initTNGVideoToolArgs()
	w.TNGRebootCount++

	w.logger.Printf("[EXEC]: %s", w.TNGArgs)
	cmd := exec.Command("sh", "-c", w.TNGArgs)
	_ = cmd.Start()
	w.TNGPid = cmd.Process.Pid
	w.TNGStartTime = time.Now()
	go func() {
		_, err := cmd.Process.Wait()
		w.logger.Printf("TNG crash: %+v", err)
		select {
		case <-w.ctx.Done():
			w.logger.Printf("[SIGNAL] TNG GoRoutine Received cancel, return")
			return
		default:
			w.logger.Printf("[SIGNAL] SEND RESTART")
			w.SIG_RESTART <- true
		}
	}()
}

func (w *Worker) startOnvif() {
	if !w.Task.IsONVIF {
		return
	}

	w.OnvifPidPath = fmt.Sprintf("/tmp/%d.pid", w.Channel+9000)
	w.initOnvifArgs()

	w.logger.Printf("[EXEC]: %s", w.OnvifArgs)
	cmd := exec.Command("sh", "-c", w.OnvifArgs)
	_ = cmd.Start()
	w.OnvifPPid = cmd.Process.Pid
	_, _ = cmd.Process.Wait()

	return
}

func (w *Worker) initOnvifArgs() {
	t := w.Task

	text := "onvif_srvd "
	text += fmt.Sprintf("--ifs %s ", localNet.Name)
	text += fmt.Sprintf("--port %d ", w.Channel + 9000)
	text += fmt.Sprintf("--pid_file /tmp/%d.pid ", w.Channel + 9000)
	text += fmt.Sprintf("--scope onvif://www.onvif.org/name/RTSPSever ")
	text += fmt.Sprintf("--scope onvif://www.onvif.org/Profile/S ")
	text += fmt.Sprintf("--name RTSPSever ")
	text += fmt.Sprintf("--width 1920 ")
	text += fmt.Sprintf("--height 1080 ")
	text += fmt.Sprintf("--url %s ", fmt.Sprintf("rtsp://127.0.0.1/%d", *t.Channel))

	oType := "JPEG"
	if strings.ToLower(t.Encoder) == "h264" {oType = "H264"}
	text += fmt.Sprintf("--type %s", oType)

	w.OnvifArgs = text
}

func (w *Worker) initTNGVideoToolArgs() {
	t := w.Task

	text := "exec TNGVideoTool "
	text += "-hide_banner "
	text += "-loglevel warning "
	text += "-stimeout 3000000 "
	text += fmt.Sprintf("-rtsp_transport %s ", t.RTSPTransPort)
	text += "-hwaccel cuvid "

	d := "h264"
	if t.Decoder == "h265" {
		d = "hevc"
	}

	text += fmt.Sprintf("-vcodec %s_cuvid ", d)
	text += fmt.Sprintf("-hwaccel_device %d ", w.GPU)
	text += fmt.Sprintf("-gpu %d ", w.GPU)
	text += fmt.Sprintf("-i '%s' ", t.RTSPAddr)
	text += "-f rtsp "
	text += fmt.Sprintf("-rtsp_transport %s ", "tcp")
	text += fmt.Sprintf("-g %d ", (*t.GOP) * (*t.FPS))
	text += fmt.Sprintf("-b:v %s ", t.BitRateA)
	text += "-zerolatency 1 "

	encoder := "h264"
	if strings.ToLower(t.Encoder) == "h265" {
		encoder = "hevc"
	}
	text += fmt.Sprintf("-vcodec %s_nvenc ", encoder)
	text += fmt.Sprintf("-profile:v %s ", t.Profile)
	text += fmt.Sprintf("-gpu %d ", w.GPU)
	text += "-acodec aac "
	text += fmt.Sprintf("-b:a %s ", t.BitRateA)
	text += fmt.Sprintf("rtsp://127.0.0.1/%d", *t.Channel)


	w.TNGArgs = text
}

func (w *Worker) start() {
	w.killOnvif(fmt.Sprintf("/tmp/%d.pid", 9001+w.Index))

	go w.log()
}

func (w *Worker) killOnvif(path string) {
	defer func() {
		w.OnvifPPid = 0
		w.OnvifPidPath = ""
	}()
	if w.OnvifPPid != 0 {
		w.logger.Printf("[onvif] kill ppid: %d", w.OnvifPPid)
		_ = syscall.Kill(w.OnvifPPid, syscall.SIGKILL)
	}

	if path != "" {
		content, _ := execCommand("cat", path)
		_ = os.Remove(path)
		if pid, err := strconv.ParseInt(strings.TrimSpace(content), 10, 64); err == nil && pid > 0 {
			_ = syscall.Kill(int(pid), syscall.SIGKILL)
			w.logger.Printf("pid: %d, err: %+v", pid, err)
		}
	}
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
