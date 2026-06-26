package engine

import "example.com/safe-rm/internal/config"

type Policy int

const (
	PolicySoftDelete Policy = iota
	PolicyPermanent
	PolicyDanger
	PolicyDangerPermanent
)

func Route(path string, cfg *config.Config) (Policy, error) {
	return PolicySoftDelete, nil
}

func SoftDelete(path string, cfg *config.Config, freeDesktop bool) error {
	return nil
}

func PermanentDelete(path string) error {
	return nil
}
