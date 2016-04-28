package fetchers

// KubewareFetchers objects will handle the download and extraction of a kubeware into a local directory. There
// will be different implementations to manage different kinds of supported sources (e.g. github repo, local dir)
type KubewareFetchers interface {
	CanHandle(kpath string) bool
	Fetch(kpath string) (string, error)
}

// allFetchers is the collection of available fetchers
var allFetchers = []KubewareFetchers{
	&GithubFetcher{},
	&LocaldirFetcher{},
}

// Fetch the indicated kubeware
func Fetch(kpath string) (string, error) {
	var localKubePath string
	var err error

	for _, fetcher := range allFetchers {
		if fetcher.CanHandle(kpath) {
			localKubePath, err = fetcher.Fetch(kpath)
			if err != nil {
				return "", err
			}
		}
	}

	return localKubePath, nil
}
