package resolvers

import (
	"net/url"
	"testing"
)

func TestGithubCanHandle(t *testing.T) {
	uri, _ := url.Parse("http://github.com/user/repo")
	githubResolver := &GithubResolver{}
	canHandle := githubResolver.CanHandle(uri)
	if !canHandle {
		t.Fatalf("GithubResolver should handle '%s'", uri.String())
	}
}

func TestGithubCanNotHandleInvalidURLs(t *testing.T) {
	uri, _ := url.Parse("anything but an url")
	githubResolver := &GithubResolver{}
	canHandle := githubResolver.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubResolver should not handle '%s'", uri.String())
	}
}

func TestGithubCanNotHandleNonGithubURLs(t *testing.T) {
	uri, _ := url.Parse("http://grijan.com/asdf/qwr")
	githubResolver := &GithubResolver{}
	canHandle := githubResolver.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubResolver should not handle '%s'", uri.String())
	}
}

func TestGithubCanNotHandleIncorrectGithubURLs(t *testing.T) {
	var uri *url.URL
	var canHandle bool

	// We test here that the path should have user and repo, no more no less
	githubResolver := &GithubResolver{}

	uri, _ = url.Parse("http://github.com/user")
	canHandle = githubResolver.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubResolver should not handle '%s'", uri.String())
	}

	uri, _ = url.Parse("http://github.com/user/repo/extrapath")
	canHandle = githubResolver.CanHandle(uri)
	if canHandle {
		t.Fatalf("GithubResolver should not handle '%s'", uri.String())
	}
}

func TestResolve(t *testing.T) {
	githubResolver := &GithubResolver{}

	uri, _ := url.Parse("https://github.com/flexiant/kubeware-spark")
	localPath, err := githubResolver.Resolve(uri)
	if err != nil || localPath == "" {
		t.Fatalf("GithubResolver could not resolve '%s'", uri.String())
	}
	t.Logf("-> %s", localPath)
}

func TestResolveWithInvalidURL(t *testing.T) {
	githubResolver := &GithubResolver{}

	uri, _ := url.Parse("https://qwlidg.com")
	localPath, err := githubResolver.Resolve(uri)
	if err == nil {
		t.Fatalf("GithubResolver should have returned error for '%s'", uri.String())
		t.Logf("-> %s", localPath)
	}
}
