package fetchers

import "net/url"

// KubewareFetchers objects will handle the download and extraction of a kubeware into a local directory. There
// will be different implementations to manage different kinds of supported sources (e.g. github repo, local dir)
type KubewareFetchers interface {
	CanHandle(url *url.URL) bool
	Fetch(url *url.URL) (string, error)
}

// AllFetchers is the collection of available fetchers
var AllFetchers = []KubewareFetchers{
	&GithubFetcher{},
}
