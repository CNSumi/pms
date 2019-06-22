package main

import (
"fmt"
"math/rand"
"os"
)

func main() {
	args := os.Args
	if len(args) >= 4 {
		fmt.Printf("exec success:rtsp://%s/%d\n", args[1], rand.Uint64())
		return
	}
	fmt.Printf("err args")
}
