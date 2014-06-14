package util

import (
	"crypto/rand"
	"encoding/base32"
	"math/big"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func RandomVersion() string {
	ver := make([]byte, 8)
	_, err := rand.Read(ver)
	if err != nil {
		panic(err)
	}
	return "0.0.0-" + strings.TrimRight(base32.StdEncoding.EncodeToString(ver), "=")
}

func DefaultIpIface() string {
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			ifaceMatches := regexp.MustCompile(`.*interface: ([^\n ]+).*`).FindStringSubmatch(line)
			if ifaceMatches != nil {
				return ifaceMatches[1]
			}
		}
	}
	return "lo0"
}

func DefaultWorldVersion() string {
	ver, err := exec.Command("uname", "-r").Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(ver))
}

func RootDir() string {
	rootDir := os.Getenv("LIGHTJAIL_ROOT")
	if rootDir == "" {
		rootDir = "/usr/local/lj"
	}
	return rootDir
}

func Hostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return host
}

func FreeIPAddr() string {
	// TODO: make this an actual thing
	i, _ := rand.Int(rand.Reader, big.NewInt(255))
	return "127.0.0." + i.String()
}
