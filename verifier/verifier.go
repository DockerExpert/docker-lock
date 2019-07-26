package verifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/michaelperel/docker-lock/generator"
	"github.com/michaelperel/docker-lock/registry"
	"io/ioutil"
)

type Verifier struct {
	*generator.Generator
}

func New(cmdLineArgs []string) (*Verifier, error) {
	var lockfileFlag string
	command := flag.NewFlagSet("verify", flag.ExitOnError)
	command.StringVar(&lockfileFlag, "o", "docker-lock.json", "Path to Lockfile from current directory.")
	command.Parse(cmdLineArgs)
	if lockfileFlag == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
	outByt, err := ioutil.ReadFile(lockfileFlag)
	if err != nil {
		return nil, err
	}
	var output generator.Output
	if err := json.Unmarshal(outByt, &output); err != nil {
		return nil, err
	}
	return &Verifier{Generator: output.Generator}, nil
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
	// TODO: No longer correct logic
	var lockfileImages, verificationImages []generator.Image
	if err := json.Unmarshal(lockfileBytes, &lockfileImages); err != nil {
		return err
	}
	if err := json.Unmarshal(verificationBytes, &verificationImages); err != nil {
		return err
	}
	if len(lockfileImages) != len(verificationImages) {
		return fmt.Errorf("Lockfile has %d images. Verification found %d images.", len(lockfileImages), len(verificationImages))
	}
	for i, _ := range lockfileImages {
		if lockfileImages[i] != verificationImages[i] {
			return fmt.Errorf("Lockfile has image %+v. Verification has image %+v.", lockfileImages[i], verificationImages[i])
		}
	}
	return errors.New("Existing lockfile does not match newly generated lockfile.")
}
