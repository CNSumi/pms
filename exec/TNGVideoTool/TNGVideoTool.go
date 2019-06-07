package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)
/**
用来模拟转码程序
 */

/**
 -hide_banner
-loglevel warning
-stimeout 3000000
-rtsp_transport tcp
-hwaccel cuvid
-vcodec hevc_cuvid
-hwaccel_device 0
-gpu 0
-i rtsp://admin:admin12345@192.168.1.12
-f rtsp
-rtsp_transport tcp
-g 250
-b:v 500k
-zerolatency 1
-profile:v high
-vcodec h264_nvenc
-gpu 0
rtsp://192.168.1.227/main

参数说明:
-hide_banner   								固定参数
-loglevel warning 						    日志打印等级   debug  info warning failed  quiet(禁止输出)
-stimeout 5000000							拉流超时时间 单位微秒
-rtsp_transport tcp 						固定参数,拉流协议,可以是udp  or  tcp(使用tcp)
-hwaccel cuvid 								固定参数
-vcodec hevc_cuvid  						源rtsp编码类型,该参数指定解码器(h264_cuvid hevc_cuvid)
-hwaccel_device 0							解码使用的GPU编号,有8个GPU,编号0-7(用作负载均衡)
-gpu 0										解码使用的GPU编号,有8个GPU,编号0-7(用作负载均衡)
-i rtsp://admin:admin12345@192.168.1.12     源rtsp地址,admin:admin12345代表用户名和密码@之后代表url
-f rtsp 									固定参数
-rtsp_transport tcp  						推流协议,根据参数设置,udp or tcp
-g 250 										gop值   fps*配置参数中gop秒数
-b:v 500k   								目标视频码率
-zerolatency 1 								固定参数
-vcodec h264_nvenc 							编码器类型(h264_nvenc  hevc_nvenc)
-profile:v high 							编码器等级(h264有三个  baseline main high   hevc有一个main)
-gpu 0										解码使用的GPU编号,有8个GPU,编号0-7(用作负载均衡)
-acodec aac									固定参数 代表音频使用aac编码(如果该源rtsp中没有音频,会报警告,但无影响)
-b:a   24k									固定参数 代表音频码率 24k
rtsp://127.0.0.1/1							目标rtsp地址,目标地址都是推到本机,本机搭载流媒体服务器,url规则为:rtsp://localhost/channel
 */

func main() {
	fs := flag.NewFlagSet("TNGVideoTool", flag.ContinueOnError)

	p_gpu := fs.Uint("gpu", 0, "gpu.no, load balance")
	p_rand := fs.Uint("rand", 10, "rand of crash")
	p_prefix := fs.String("prefix", "0", "prefix")
	fs.Parse(os.Args[1:])

	prefix := *p_prefix
	num := *p_rand
	log.SetPrefix(fmt.Sprintf("[%s][%d]", prefix, *p_gpu))

	for i := uint64(0); i < 1024; i = (i + 1) % 1024 {
		log.Printf("live %d", i)
		if rand.Intn(int(num)) == 0 {
			log.Fatalf("random crash")
		}
		time.Sleep(time.Second)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

