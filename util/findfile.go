package util

import (
	"io/ioutil"

        "github.com/myfreeweb/gosemver"
)

func FindMaxVersionFile(dir, constraint string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Name()
	}

	versions := gosemver.MustParseVersions(filenames)
	maxv := gosemver.MustFindMax(versions, constraint)
	return maxv.String()
}
