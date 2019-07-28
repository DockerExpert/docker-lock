package verifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/michaelperel/docker-lock/generator"
	"github.com/michaelperel/docker-lock/registry"
	"io/ioutil"
)

type Verifier struct {
	*generator.Generator
}

func New(flags *Flags) (*Verifier, error) {
	lockfileByt, err := ioutil.ReadFile(flags.Outfile)
	if err != nil {
		return nil, err
	}
	var lockfile generator.Lockfile
	if err := json.Unmarshal(lockfileByt, &lockfile); err != nil {
		return nil, err
	}
	return &Verifier{Generator: lockfile.Generator}, nil
}

func (v *Verifier) VerifyLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := ioutil.ReadFile(v.Outfile)
	if err != nil {
		return err
	}
	verificationBytes, err := v.GenerateLockfileBytes(wrapper)
	if err != nil {
		return err
	}
	if bytes.Equal(lockfileBytes, verificationBytes) {
		return nil
	}
	var existingLockfile, verificationlockfile generator.Lockfile
	if err := json.Unmarshal(lockfileBytes, &existingLockfile); err != nil {
		return err
	}
	if err := json.Unmarshal(verificationBytes, &verificationlockfile); err != nil {
		return err
	}
	if len(existingLockfile.Images) != len(verificationlockfile.Images) {
		return fmt.Errorf("Existing lockfile has %d images. Verification found %d images.", len(existingLockfile.Images), len(verificationlockfile.Images))
	}
	for i, _ := range existingLockfile.Images {
		if existingLockfile.Images[i] != verificationlockfile.Images[i] {
			return fmt.Errorf("Existing lockfile has image %+v. Verification has image %+v.", existingLockfile.Images[i], verificationlockfile.Images[i])
		}
	}
	return errors.New("Existing lockfile does not match newly generated lockfile.")
}
