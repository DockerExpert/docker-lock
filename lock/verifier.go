package lock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/michaelperel/docker-lock/registry"
	"io/ioutil"
)

type Verifier struct {
	Generator
}

func (v *Verifier) VerifyLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := ioutil.ReadFile(v.Lockfile)
	if err != nil {
		return err
	}
	verificationBytes, err := v.generateLockfileBytes(wrapper)
	if err != nil {
		return err
	}
	if bytes.Equal(lockfileBytes, verificationBytes) {
		return nil
	}
	var lockfileImages, verificationImages []image
	if err := json.Unmarshal(lockfileBytes, &lockfileImages); err != nil {
		return err
	}
	if err := json.Unmarshal(verificationBytes, &verificationImages); err != nil {
		return err
	}
	if len(lockfileImages) != len(verificationImages) {
		return fmt.Errorf("Got %d images. Want %d images.", len(lockfileImages), len(verificationImages))
	}
	for i, _ := range lockfileImages {
		if lockfileImages[i] != verificationImages[i] {
			return fmt.Errorf("Got %+v. Want %+v.", lockfileImages[i], verificationImages[i])
		}
	}
	return errors.New("Existing lockfile does not match newly generated lockfile.")
}
