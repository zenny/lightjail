package util

import (
	"io/ioutil"
	"log"
	"sort"
	"strings"
)

func FindPrefixedFile(dir, prefix string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	names := []string{}
	for _, file := range files {
		fname := file.Name()
		if strings.HasPrefix(fname, prefix) {
			names = append(names, fname)
		}
	}
	sort.Strings(names)
	return names[len(names)-1]
}
