//nolint
package test

import (
	"os"
	"path"
	"runtime"
)

// REFACTOR me

// importing _ "github.com/trento-project/agent/test" in tests would set the cwd to the root of the repo
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
