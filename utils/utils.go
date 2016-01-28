package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/kdeploy/webservice"
)

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	CheckError(err)
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		CheckError(err)
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			CheckError(err)
			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			CheckError(err)
			defer f.Close()

			_, err = io.Copy(f, rc)
			CheckError(err)
		}
	}
	return nil
}

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
	kubewareName := path[2]

	newPath := append([]string{""}, path[1], path[2], "archive", "master.zip")

	kubewareURL.Path = strings.Join(newPath, "/")

	client, err := webservice.NewSimpleWebClient(kubewareURL.String())
	CheckError(err)

	tmpDir, err := ioutil.TempDir("", "kdeploy")
	CheckError(err)

	zipFileLocation, err := client.GetFile(kubewareURL.Path, tmpDir)
	CheckError(err)

	err = Unzip(zipFileLocation, tmpDir)
	CheckError(err)

	return fmt.Sprintf("%s/%s-master/", tmpDir, kubewareName), nil
}
