package lock

import "io/ioutil"

func WriteFile(lockfile string, lockfileBytes []byte) {
	err := ioutil.WriteFile(lockfile, lockfileBytes, 0644)
	if err != nil {
		panic(err)
	}
}
