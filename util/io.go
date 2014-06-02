package util

import (
	"io/ioutil"
)

func MustReadFileStr(path string) string {
	txt, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(txt)
}

func MustTempDir(prefix string) string {
	mountPoint, err := ioutil.TempDir("", prefix)
	if err != nil {
		panic(err)
	}
	return mountPoint
}
