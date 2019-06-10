package main

import (
"fmt"
"math/rand"
"os"
)

func main() {
	args := os.Args
	if len(args) >= 4 {
		fmt.Printf("exec success:rtsp://%s:%s@%s/%d\n", args[2], args[3], args[1], rand.Uint64())
		return
	}
	fmt.Printf("err args")
}
