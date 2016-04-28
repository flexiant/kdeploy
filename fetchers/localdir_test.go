package fetchers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLocaldirCanHandle(t *testing.T) {
	path := tempFakeKubeware()
	defer os.RemoveAll(path)

	githubFetcher := &LocaldirFetcher{}
	canHandle := githubFetcher.CanHandle(path)
	if !canHandle {
		t.Errorf("LocaldirFetcher should handle '%s'", path)
	}
}

func TestLocaldirCanHandleRel(t *testing.T) {
	abs := tempFakeKubeware()
	defer os.RemoveAll(abs)

	path, err := filepath.Rel(".", abs)
	t.Log(err)
	t.Log(path)
	t.Log(abs)
	githubFetcher := &LocaldirFetcher{}
	canHandle := githubFetcher.CanHandle(path)
	if !canHandle {
		t.Errorf("LocaldirFetcher should handle '%s'", path)
	}
}

func TestLocaldirCanNotHandleURLs(t *testing.T) {
	path := "https://something"
	githubFetcher := &LocaldirFetcher{}
	canHandle := githubFetcher.CanHandle(path)
	if canHandle {
		r, err := githubFetcher.Fetch(path)
		t.Log(r, err)
		t.Errorf("LocaldirFetcher should not handle '%s'", path)
	}
}

func TestLocaldirFetch(t *testing.T) {
	githubFetcher := &LocaldirFetcher{}
	kurl := tempFakeKubeware()
	defer os.RemoveAll(kurl)

	localPath, err := githubFetcher.Fetch(kurl)
	if err != nil || localPath == "" {
		t.Errorf("LocaldirFetcher could not resolve '%s'", kurl)
	}
	t.Logf("-> %s", localPath)
}

func TestLocaldirFetchRelativePath(t *testing.T) {
	githubFetcher := &LocaldirFetcher{}
	abs := tempFakeKubeware()
	defer os.RemoveAll(abs)

	path, _ := filepath.Rel(".", abs)
	localPath, err := githubFetcher.Fetch(path)
	if err != nil || localPath == "" {
		t.Errorf("LocaldirFetcher should resolve '%s'", path)
	}
	t.Logf("-> %s", localPath)
}

func TestLocaldirFetchNoMetadata(t *testing.T) {
	githubFetcher := &LocaldirFetcher{}
	kurl := tempEmptyDir()
	defer os.RemoveAll(kurl)

	localPath, err := githubFetcher.Fetch(kurl)
	if err == nil || localPath != "" {
		t.Errorf("LocaldirFetcher should not resolve '%s'", kurl)
	}
	t.Logf("-> %s", localPath)
}

func tempFakeKubeware() string {
	tmpDir, _ := ioutil.TempDir(".", "")
	metadataPath := filepath.Join(tmpDir, "metadata.yaml")
	ioutil.WriteFile(metadataPath, []byte{}, 700)
	return tmpDir
}

func tempEmptyDir() string {
	tmpDir, _ := ioutil.TempDir(".", "")
	return tmpDir
}
