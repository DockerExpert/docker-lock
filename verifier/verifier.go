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

func New(cmdLineArgs []string) (*Verifier, error) {
	flags, err := newFlags(cmdLineArgs)
	if err != nil {
		return nil, err
	}
	resultByt, err := ioutil.ReadFile(flags.lockfile)
	if err != nil {
		return nil, err
	}
	var result generator.Result
	if err := json.Unmarshal(resultByt, &result); err != nil {
		return nil, err
	}
	return &Verifier{Generator: result.Generator}, nil
}

func (v *Verifier) VerifyLockfile(wrapper registry.Wrapper) error {
	lockfileBytes, err := ioutil.ReadFile(v.Lockfile)
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
	var lockfileResult, verificationResult generator.Result
	if err := json.Unmarshal(lockfileBytes, &lockfileResult); err != nil {
		return err
	}
	if err := json.Unmarshal(verificationBytes, &verificationResult); err != nil {
		return err
	}
	if len(lockfileResult.Images) != len(verificationResult.Images) {
		return fmt.Errorf("Lockfile has %d images. Verification found %d images.", len(lockfileResult.Images), len(verificationResult.Images))
	}
	for i, _ := range lockfileResult.Images {
		if lockfileResult.Images[i] != verificationResult.Images[i] {
			return fmt.Errorf("Lockfile has image %+v. Verification has image %+v.", lockfileResult.Images[i], verificationResult.Images[i])
		}
	}
	return errors.New("Existing lockfile does not match newly generated lockfile.")
}
