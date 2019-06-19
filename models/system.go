package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	Unix_KB         = 1 << 10
	Unit_KB_float64 = float64(Unix_KB)
)

var (
	stat     = &SystemStat{}
	version  = beego.AppConfig.DefaultString("version", "1.0.0")
	localNet = &NetStat{}
	ipMaskGateRegex = regexp.MustCompile(`^\s+inet\s+(\d{1,3}(\.\d{1,3}){3})\s+netmask\s+(\d{1,3}(\.\d{1,3}){3})\s+broadcast\s+(\d{1,3}(\.\d{1,3}){3})`)
)

type NetStat struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Mask string `json:"mask"`
	Gate string `json:"gate"`
}

type SystemStat struct {
	CPUUsed    float64 `json:"cpu_used"`
	MemoryUsed float64 `json:"mem_used"`
	UpFlow     float64 `json:"upflow"`
	DownFlow   float64 `json:"downflow"`
}

func initSystem() {
	if err := initLocalNet(); err != nil {
		//log.Fatalf("initLocalNet fail: %+v", err)
	}

	go func() {
		for true {
			ms, err := mem.VirtualMemory()
			if err == nil && ms != nil {
				stat.MemoryUsed = trans2dot2ff64(ms.UsedPercent)
			}

			cus, err := cpu.Percent(0, false)
			if err == nil && cus != nil && len(cus) == 1 {
				stat.CPUUsed = trans2dot2ff64(cus[0])
			}

			t1 := time.Now().UnixNano()
			io1s, _ := net.IOCounters(false)
			time.Sleep(time.Millisecond * 1000)
			interval := float64((time.Now().UnixNano() - t1) / 1e9)
			io2s, _ := net.IOCounters(false)

			if io1s != nil && len(io1s) == 1 && io2s != nil && len(io2s) == 1 {
				stat.UpFlow = trans2dot2ff64(float64(io2s[0].BytesSent-io1s[0].BytesSent) / interval / Unit_KB_float64)
				stat.DownFlow = trans2dot2ff64(float64(io2s[0].BytesRecv-io1s[0].BytesRecv) / interval / Unit_KB_float64)
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}()
}

func GetSystemStat() *SystemStat {
	return stat
}

func AppVersion() string {
	return version
}

func initLocalNet() error {
	localNet.Name = beego.AppConfig.String("eth")
	if localNet.Name == "" {
		return fmt.Errorf("missing network interface name conf, (ex: eth = wlp3s0@./conf/app.conf)")
	}

	cmd := exec.Command("ifconfig", localNet.Name)
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("command(%s) exec fail: %+v", strings.Join(cmd.Args, " "), err)
	}
	content := string(b)
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return fmt.Errorf("ifconfig: interface %s  not exist", localNet.Name)
	}
	fields := ipMaskGateRegex.FindStringSubmatch(lines[1])
	if len(fields) <= 6 {
		return fmt.Errorf("unsupport os")
	}
	localNet.IP = fields[1]
	localNet.Mask = fields[3]
	localNet.Gate = fields[5]

	return nil
}

func LocalNet() *NetStat {
	return localNet
}

func Reboot() error {
	out, err := execCommand("/sbin/reboot")
	log.Printf("reboot.out: %s\n", string(out))
	return err
}

func SetNetwork(inet, mask, broadcast string) error {
	if !isIPV4Addr(inet) {
		return fmt.Errorf("not ipv4 addr: %s", inet)
	}
	if !isIPV4Addr(mask) {
		return fmt.Errorf("not ipv4 addr: %s", mask)
	}
	if !isIPV4Addr(broadcast) {
		return fmt.Errorf("not ipv4 addr: %s", broadcast)
	}
	return nil
}