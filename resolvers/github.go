package resolvers

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// GithubResolver downloads and extracts a kubeware from a github repo
type GithubResolver struct{}

// CanHandle tells if the URL can be handled by this resolver
func (gh *GithubResolver) CanHandle(uri *url.URL) bool {
	if uri.Host == "github.com" {
		path := strings.Split(uri.Path, "/")
		if len(path) != 3 {
			return false
		}
		return true
	}
	return false
}

// Resolve downloads and extract archive/master zip from github repo into
// a temporal local directory, and returns its path
func (gh *GithubResolver) Resolve(uri *url.URL) (string, error) {
	if !gh.CanHandle(uri) {
		return "", fmt.Errorf("URL can't be handled by GithubResolver: '%s'", uri.String())
	}
	path := strings.Split(uri.Path, "/")
	kubewareName := path[2]
	newPath := append([]string{""}, path[1], path[2], "archive", "master.zip")

	kubewareURL := *uri
	kubewareURL.Path = strings.Join(newPath, "/")
	client, err := webservice.NewSimpleWebClient(kubewareURL.String())
	if err != nil {
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "kdeploy")
	if err != nil {
		return "", err
	}

	zipFileLocation, err := client.GetFile(kubewareURL.Path, tmpDir)
	if err != nil {
		return "", err
	}

	err = utils.Unzip(zipFileLocation, tmpDir)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s-master/", tmpDir, kubewareName), nil
}
