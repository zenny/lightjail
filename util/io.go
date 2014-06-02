package util

import (
	"io/ioutil"
)

func MustReadFile(path string) []byte {
	txt, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return txt
}

func MustReadFileStr(path string) string {
	return string(MustReadFile(path))
}

func MustTempDir(prefix string) string {
	mountPoint, err := ioutil.TempDir("", prefix)
	if err != nil {
		panic(err)
	}
	return mountPoint
}
