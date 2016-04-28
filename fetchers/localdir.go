package fetchers

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocaldirFetcher copies a kubeware from a local url
type LocaldirFetcher struct{}

// CanHandle tells if the path can be handled by this resolver
func (gh *LocaldirFetcher) CanHandle(kpath string) bool {
	abs, err := filepath.Abs(kpath)
	if err != nil || abs == "" {
		return false
	}
	metadata := filepath.Join(abs, "metadata.yaml")
	if _, err := os.Stat(metadata); os.IsNotExist(err) {
		return false
	}
	return true
}

// Fetch simply returns its absolute path
func (gh *LocaldirFetcher) Fetch(kpath string) (string, error) {
	if !gh.CanHandle(kpath) {
		return "", fmt.Errorf("URL can't be handled by LocaldirFetcher: '%s'", kpath)
	}
	return filepath.Abs(kpath)
}
