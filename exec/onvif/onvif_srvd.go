package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"
)

/**
2.onvif程序使用实例(该程序自带守护进程)
./onvif_srvd
--ifs   	eth0
--port 		8001
--pid_file  /tmp/8001.pid
--scope onvif://www.onvif.org/name/TestDev
--scope onvif://www.onvif.org/Profile/S
--name RTSPSever
--width 1920
--height 1080
--url rtsp://%s:554/unicast
--type JPEG"

参数说明:
--ifs eth0  								 选择网卡  "enth0"为网卡名称,根据实际需要修改
--port 		8001							 使用端口  规则为8000+channel
--pid_file  /tmp/8001.pid 					 守护进程使用的pidfile  规则 8000+channel+.pid
											 注:同名文件如果存在会导致程序启动失败
--scope onvif://www.onvif.org/name/RTSPSever 固定参数
--scope onvif://www.onvif.org/Profile/S 	 固定参数
--name RTSPSever 							 固定参数
--width 800 								 目标rtsp流的分辨率  宽
--height 600 								 目标rtsp流的分辨率  高
--url rtsp://127.0.0.1/1				 	 目标rtsp地址
--type JPEG"								 如果编码器使用h264,则用H264;使用hevc则用JPEG
*/
func main() {
	f := flag.NewFlagSet("onvif", flag.ContinueOnError)

	p_ifs := f.String("ifs", "etho", "选择网卡  enth0为网卡名称,根据实际需要修改")
	p_port := f.Uint("port", 9001, "使用端口  规则为9000+channel")
	p_pid_file := f.String("pid_file", "/tmp/9001.pid", "守护进程使用的pidfile  规则 9000+channel+.pid")
	p_scope := f.String("scope", "onvif://www.onvif.org/name/RTSPSever", "固定参数")
	// scope
	p_name := f.String("name", "RTSPSever", "服务名")
	p_width := f.Uint("width", 1920, "目标rtsp流的分辨率  宽")
	p_height := f.Uint("height", 1080, "目标rtsp流的分辨率  高")
	p_url := f.String("url", "rtsp://127.0.0.1/1", "目标rtsp地址")
	p_type := f.String("type", "JPEG", "如果编码器使用h264,则用H264;使用hevc则用JPEG")
	f.Parse(os.Args[1:])

	for {
		log.Printf("ifs: %s\n", *p_ifs)
		log.Printf("ifs: %d\n", *p_port)
		log.Printf("ifs: %s\n", *p_pid_file)
		log.Printf("ifs: %s\n", *p_scope)
		log.Printf("ifs: %s\n", *p_name)
		log.Printf("ifs: %d\n", *p_width)
		log.Printf("ifs: %d\n", *p_height)
		log.Printf("ifs: %s\n", *p_url)
		log.Printf("ifs: %s\n", *p_type)
		log.Printf("----------------------------------")
		time.Sleep(time.Second * time.Duration(rand.Intn(5)))
	}
}
