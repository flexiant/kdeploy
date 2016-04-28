package fetchers

// KubewareFetchers objects will handle the download and extraction of a kubeware into a local directory. There
// will be different implementations to manage different kinds of supported sources (e.g. github repo, local dir)
type KubewareFetchers interface {
	CanHandle(kpath string) bool
	Fetch(kpath string) (string, error)
}

// AllFetchers is the collection of available fetchers
var AllFetchers = []KubewareFetchers{
	&GithubFetcher{},
	&LocaldirFetcher{},
}
