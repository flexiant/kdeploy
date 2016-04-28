package resolvers

import "net/url"

// KubewareResolver objects will handle the download and extraction of a kubeware into a local directory. There
// will be different implementations to manage different kinds of supported sources (e.g. github repo, local dir)
type KubewareResolver interface {
	CanHandle(url *url.URL) bool
	Resolve(url *url.URL) (string, error)
}

// URLResolvers is the collection of available resolvers
var URLResolvers = []KubewareResolver{
	&GithubResolver{},
}
