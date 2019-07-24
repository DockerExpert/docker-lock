package registry

type Wrapper interface {
	GetDigest() (string, error)
}
