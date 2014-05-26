package util

import (
	"github.com/myfreeweb/gosemver"
	"io/ioutil"
	"log"
)

func FindMaxVersionFile(dir, constraint string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Name()
	}

	versions, err := gosemver.ParseVersions(filenames)
	if err != nil {
		log.Fatal(err)
	}

	maxv, err := gosemver.FindMax(versions, constraint)
	if err != nil {
		log.Fatal(err)
	}
	return maxv.String()
}
