package appinfo

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func prefix(appName, confGroup string) string {
	p := appName + "_"

	if confGroup != "" {
		p = appName + "__" + confGroup
	}

	return strings.ToUpper(strings.Replace(p, "-", "_", -1))
}

func writeToFile(filename string, v interface{}, marshal func(v interface{}) ([]byte, error)) error {
	bytes, _ := marshal(v)
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filename, bytes, os.ModePerm)
}
