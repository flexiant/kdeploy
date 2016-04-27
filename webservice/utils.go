package webservice

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/flexiant/kdeploy/utils"
)

// FetchKubeFromURL fetches a remote kubeware, extracts it locally, and returns its local path
func FetchKubeFromURL(kubeURL string) (string, error) {
	kubewareURL, err := url.Parse(kubeURL)
	if err != nil {
		return "", err
	}

	if kubewareURL.Host != "github.com" {
		return "", errors.New("We currently only support Github urls")
	}

	path := strings.Split(kubewareURL.Path, "/")
	if len(path) != 3 {
		return "", fmt.Errorf("Github URL should have the following format: 'https://github.com/<user>/<repo_name>'")
	}
	kubewareName := path[2]

	newPath := append([]string{""}, path[1], path[2], "archive", "master.zip")

	kubewareURL.Path = strings.Join(newPath, "/")

	client, err := NewSimpleWebClient(kubewareURL.String())
	utils.CheckError(err)

	tmpDir, err := ioutil.TempDir("", "kdeploy")
	utils.CheckError(err)

	zipFileLocation, err := client.GetFile(kubewareURL.Path, tmpDir)
	utils.CheckError(err)

	err = utils.Unzip(zipFileLocation, tmpDir)
	utils.CheckError(err)

	return fmt.Sprintf("%s/%s-master/", tmpDir, kubewareName), nil
}
