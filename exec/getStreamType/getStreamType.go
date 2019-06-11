package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(3)
	switch r {
	case 0:
		fmt.Printf("stream type h264\n")
	case 1:
		fmt.Printf("stream type hevc\n")
	default:
		fmt.Printf("stream type not\n")
	}
}
