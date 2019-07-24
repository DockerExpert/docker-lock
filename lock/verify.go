package lock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func VerifyLockfile(options Options) (bool, string) {
	lockfileBytes := readLockfile(options.Lockfile)
	verificationBytes := generateLockfileBytes(options)
	equal := bytes.Equal(lockfileBytes, verificationBytes)
	var reason string
	if !equal {
		var lockfileImages, verificationImages []image
		if err := json.Unmarshal(lockfileBytes, &lockfileImages); err != nil {
			panic(err)
		}
		if err := json.Unmarshal(verificationBytes, &verificationImages); err != nil {
			panic(err)
		}
		if len(lockfileImages) != len(verificationImages) {
			reason = fmt.Sprintf("Got %d images. Want %d images.", len(lockfileImages), len(verificationImages))
			return equal, reason
		}
		for i, _ := range lockfileImages {
			if lockfileImages[i] != verificationImages[i] {
				reason = fmt.Sprintf("Got %+v. Want %+v.", lockfileImages[i], verificationImages[i])
				return equal, reason
			}
		}
	}
	reason = fmt.Sprintf("Regenerated same bytes as in file: '%s'\n", options.Lockfile)
	return equal, reason
}

func readLockfile(lockfile string) []byte {
	existing, err := ioutil.ReadFile(lockfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return existing
}
