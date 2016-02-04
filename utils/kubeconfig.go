package utils

// KubeConfig holds the information needed to build connect to remote kubernetes clusters as a given user
type KubeConfig struct {
	Kind           string        `yaml:"kind,omitempty"`
	APIVersion     string        `yaml:"apiVersion,omitempty"`
	Clusters       []ClusterWrap `yaml:"clusters,omitempty"`
	Users          []UserWrap    `yaml:"users,omitempty"`
	Contexts       []ContextWrap `yaml:"contexts,omitempty"`
	CurrentContext string        `yaml:"current-context,omitempty"`
}

type ClusterWrap struct {
	Cluster Cluster `yaml:"cluster,omitempty"`
	Name    string  `yaml:"name,omitempty"`
}

type Cluster struct {
	Server                string `yaml:"server,omitempty"`
	APIVersion            string `yaml:"apiVersion,omitempty"`
	InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify,omitempty"`
	CertificateAuthority  string `yaml:"certificate-authority,omitempty"`
}

type UserWrap struct {
	User User   `yaml:"user,omitempty"`
	Name string `yaml:"name,omitempty"`
}

type ContextWrap struct {
	Context Context `yaml:"context,omitempty"`
	Name    string  `yaml:"name,omitempty"`
}

// User contains information that describes identity information.  This is use to tell the kubernetes cluster who you are.
type User struct {
	ClientCertificate string `yaml:"client-certificate,omitempty"`
	ClientKey         string `yaml:"client-key,omitempty"`
}

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes cluster), a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}
