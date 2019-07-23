package lock

import (
	"bytes"
	"fmt"
	"github.com/michaelperel/docker-lock/options"
	"io/ioutil"
	"os"
)

func Verify(options options.Options) (bool, string) {
	existing := getLockfile(options.Lockfile)
	generated := Generate(options)
	equal := bytes.Equal(existing, generated)
	var reason string
	if !equal {
		reason = "TODO"
	}
	return equal, reason
}

func getLockfile(lockfile string) []byte {
	existing, err := ioutil.ReadFile(lockfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return existing
}
