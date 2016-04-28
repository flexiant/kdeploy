package fetchers

import (
	"testing"
)

func TestGithubCanHandle(t *testing.T) {
	uri := "http://github.com/user/repo"
	githubFetcher := &GithubFetcher{}
	canHandle := githubFetcher.CanHandle(uri)
	if !canHandle {
		t.Fatalf("GithubFetcher should handle '%s'", uri)
	}
}

func TestGithubCanNotHandleInvalidURLs(t *testing.T) {
	uri := "anything but an url"
	githubFetcher := &GithubFetcher{}
	canHandle := githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri)
	}
}

func TestGithubCanNotHandleNonGithubURLs(t *testing.T) {
	uri := "http://grijan.com/asdf/qwr"
	githubFetcher := &GithubFetcher{}
	canHandle := githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri)
	}
}

func TestGithubCanNotHandleIncorrectGithubURLs(t *testing.T) {
	var uri string
	var canHandle bool

	// We test here that the path should have user and repo, no more no less
	githubFetcher := &GithubFetcher{}

	uri = "http://github.com/user"
	canHandle = githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri)
	}

	uri = "http://github.com/user/repo/extrapath"
	canHandle = githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri)
	}
}

func TestFetch(t *testing.T) {
	githubFetcher := &GithubFetcher{}

	uri := "https://github.com/flexiant/kubeware-spark"
	localPath, err := githubFetcher.Fetch(uri)
	if err != nil || localPath == "" {
		t.Fatalf("GithubFetcher could not resolve '%s'", uri)
	}
	t.Logf("-> %s", localPath)
}

func TestFetchWithInvalidURL(t *testing.T) {
	githubFetcher := &GithubFetcher{}

	uri := "https://qwlidg.com"
	localPath, err := githubFetcher.Fetch(uri)
	if err == nil {
		t.Fatalf("GithubFetcher should have returned error for '%s'", uri)
		t.Logf("-> %s", localPath)
	}
}
