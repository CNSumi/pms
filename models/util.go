package models

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	contentRegex = regexp.MustCompile(`addr: (.*)`)
)

func isIPV4Addr(ip string) bool {
	if len(ip) < 7 || len(ip) > 15 {
		return false
	}
	fields := strings.Split(ip, ".")
	if len(fields) != 4 {
		return false
	}
	for _, sub := range fields {
		tmp, err := strconv.ParseUint(sub, 10, 64)
		if err != nil || tmp > 255 {
			return false
		}
	}
	return true
}

func trans2dot2ff64(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return num
}

func inSet_string(needle string, haystack []string) bool {
	for _, item := range haystack {
		if needle == item {
			return true
		}
	}
	return false
}

func GetRTSPAddress(host, user, password string) (string, error) {
	if user == "" || password == "" {
		return "", fmt.Errorf("empty user or password")
	}
	if !isIPV4Addr(host) {
		return "", fmt.Errorf("invalid ip(%s)", host)
	}

	content, _ := execCommand("getrtsp", host, user, password)
	content = strings.TrimSpace(content)
	matches := contentRegex.FindStringSubmatch(content)
	log.Printf("matches: %+v", matches)
	if len(matches) == 2 {
		return matches[1], nil
	}

	return content, fmt.Errorf("unknown error")
}

func execCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	b, err := cmd.Output()
	return string(b), err
}

