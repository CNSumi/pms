package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	OnvifPid	int

	// TNGVideoTool
	TNGArgs        string
	TNGPid         int
	TNGRebootCount int
	TNGStartTime   time.Time
	TNGMessage     string

	Msg   string
	Lock  bool
	Index int // index at worker array, same GPU
	GPU   int // run at which gpu, number [0-GPU_COUNT), value equal to index % GPU_COUNT

	logger          *log.Logger
	logWorkerTicker *time.Ticker
	logTaskTicker   *time.Ticker

	SIG_RESTART chan bool
	cancel      context.CancelFunc
	ctx         context.Context
}

var (
	workers = make([]*Worker, 0, GPU_COUNT* 10)
	id2idx  = map[int64]int{}
)

func initProcess() {
	for i := 0; i < cap(workers); i++ {
		workers = append(workers, nil)
	}

	for i := 0; i < len(workers); i++ {
		sub := &Worker{}
		sub.Index = i
		sub.GPU = sub.Index % GPU_COUNT
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
			w.killOnvif()
			w.killTNG()
			w.Lock = false
			return
		case <-w.SIG_RESTART:
			w.logger.Printf("[SIGNAL] RECEIVED RESTART")
			w.killOnvif()
			w.killTNG()
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

	w.OnvifArgs = fmt.Sprintf("exec onvifProxy %d %s %s", w.Channel, localNet.IP, w.Task.ONVIF_IP)
	w.logger.Printf("[EXEC]: %s", w.OnvifArgs)

	cmd := exec.Command("sh", "-c", w.OnvifArgs)
	_ = cmd.Start()
	w.OnvifPid = cmd.Process.Pid
	go func() {
		_, err := cmd.Process.Wait()
		w.logger.Printf("onvifProxy crash: %+v", err)
		select {
		case <-w.ctx.Done():
			w.logger.Printf("[SIGNAL] onvifProxy GoRoutine Received cancel, return")
			return
		default:
			w.logger.Printf("[SIGNAL] onvifProxy send RESTART")
			w.SIG_RESTART <- true
		}
	}()
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
	go w.log()
}

func (w *Worker) killOnvif() {
	defer func() {
		w.OnvifPid = 0
	}()
	if w.OnvifPid != 0 {
		w.logger.Printf("[onvif] kill pid: %d", w.OnvifPid)
		_ = syscall.Kill(w.OnvifPid, syscall.SIGKILL)
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
