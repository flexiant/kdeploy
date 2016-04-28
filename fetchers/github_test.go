package fetchers

import (
	"net/url"
	"testing"
)

func TestGithubCanHandle(t *testing.T) {
	uri, _ := url.Parse("http://github.com/user/repo")
	githubFetcher := &GithubFetcher{}
	canHandle := githubFetcher.CanHandle(uri)
	if !canHandle {
		t.Fatalf("GithubFetcher should handle '%s'", uri.String())
	}
}

func TestGithubCanNotHandleInvalidURLs(t *testing.T) {
	uri, _ := url.Parse("anything but an url")
	githubFetcher := &GithubFetcher{}
	canHandle := githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri.String())
	}
}

func TestGithubCanNotHandleNonGithubURLs(t *testing.T) {
	uri, _ := url.Parse("http://grijan.com/asdf/qwr")
	githubFetcher := &GithubFetcher{}
	canHandle := githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri.String())
	}
}

func TestGithubCanNotHandleIncorrectGithubURLs(t *testing.T) {
	var uri *url.URL
	var canHandle bool

	// We test here that the path should have user and repo, no more no less
	githubFetcher := &GithubFetcher{}

	uri, _ = url.Parse("http://github.com/user")
	canHandle = githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri.String())
	}

	uri, _ = url.Parse("http://github.com/user/repo/extrapath")
	canHandle = githubFetcher.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubFetcher should not handle '%s'", uri.String())
	}
}

func TestFetch(t *testing.T) {
	githubFetcher := &GithubFetcher{}

	uri, _ := url.Parse("https://github.com/flexiant/kubeware-spark")
	localPath, err := githubFetcher.Fetch(uri)
	if err != nil || localPath == "" {
		t.Fatalf("GithubFetcher could not resolve '%s'", uri.String())
	}
	t.Logf("-> %s", localPath)
}

func TestFetchWithInvalidURL(t *testing.T) {
	githubFetcher := &GithubFetcher{}

	uri, _ := url.Parse("https://qwlidg.com")
	localPath, err := githubFetcher.Fetch(uri)
	if err == nil {
		t.Fatalf("GithubFetcher should have returned error for '%s'", uri.String())
		t.Logf("-> %s", localPath)
	}
}
