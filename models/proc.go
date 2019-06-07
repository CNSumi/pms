package models

import (
	"log"
	"os/exec"
)

func startTask(t *Task) {
	cmd := exec.Command("TNGVideoTool")
	err := cmd.Start()
	log.Printf("process.start err: %+v", err)
	log.Printf("process.start cmd: %+v", cmd)

}