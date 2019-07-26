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

func New(g *generator.Generator) (*Verifier, error) {
	if g.Lockfile == "" {
		return nil, errors.New("Lockfile cannot be empty.")
	}
	return &Verifier{Generator: g}, nil
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
