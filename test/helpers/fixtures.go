package helpers

import (
	"path"
	"runtime"
)

var fixturesFolder = "" // nolint:gochecknoglobals

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("error recovering caller information in test helper")
	}
	fixturesFolder = path.Join(path.Dir(filename), "../fixtures")
}

func GetFixturePath(name string) string {
	return path.Join(fixturesFolder, name)
}
