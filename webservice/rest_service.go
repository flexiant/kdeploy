package webservice

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/flexiant/kdeploy/utils"
)

type RestService struct {
	client   *http.Client
	endpoint *url.URL
}

func NewRestService(config utils.Config) (*RestService, error) {
	client, err := httpClient(config)
	if err != nil {
		return nil, err
	}

	endpoint, err := url.Parse(config.Connection.APIEndpoint)
	if err != nil {
		return nil, err
	}
	return &RestService{client, endpoint}, nil
}

func NewSimpleWebClient(httpUrl string) (*RestService, error) {
	parsedUrl, err := url.Parse(httpUrl)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}

	return &RestService{client, parsedUrl}, nil
}

func httpClient(config utils.Config) (*http.Client, error) {
	// load client certificate
	cert, err := tls.LoadX509KeyPair(config.Connection.Cert, config.Connection.Key)
	if err != nil {
		return nil, err
	}
	// load CA file to verify server
	caPool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile(config.Connection.CACert)
	if err != nil {
		return nil, err
	}
	caPool.AppendCertsFromPEM(severCert)
	// create a client with specific transport configurations
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      caPool,
			Certificates: []tls.Certificate{cert},
		},
	}
	client := &http.Client{Transport: transport}

	return client, nil
}

func prettyprint(b []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	return out.Bytes()
}

func (r *RestService) Post(urlPath string, json []byte) ([]byte, int, error) {
	r.endpoint.Path = urlPath
	output := strings.NewReader(string(json))

	if os.Getenv("KDEPLOY_DRYRUN") == "1" {
		log.Infof("Post request url: %s , body:\n%s", r.endpoint.String(), string(prettyprint(json)))
		return nil, 200, nil
	} else {
		log.Debugf("Post request url: %s , body:\n%s", r.endpoint.String(), string(prettyprint(json)))
	}

	response, err := r.client.Post(r.endpoint.String(), "application/json", output)
	if err != nil {
		return nil, -1, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, err
	}

	return body, response.StatusCode, err
}

func (r *RestService) Delete(urlPath string) ([]byte, int, error) {
	r.endpoint.Path = urlPath

	if os.Getenv("KDEPLOY_DRYRUN") == "1" {
		log.Infof("Delete request url: %s", r.endpoint.String())
		return nil, 200, nil
	} else {
		log.Debugf("Delete request url: %s", r.endpoint.String())
	}

	request, err := http.NewRequest("DELETE", r.endpoint.String(), nil)
	if err != nil {
		return nil, -1, err
	}
	response, err := r.client.Do(request)
	if err != nil {
		return nil, -1, err
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, err
	}

	return body, response.StatusCode, nil
}

func (r *RestService) Get(urlPath string) ([]byte, int, error) {
	r.endpoint.Path = urlPath

	if os.Getenv("KDEPLOY_DRYRUN") == "1" {
		log.Infof("Get request url: %s", r.endpoint.String())
		return nil, 200, nil
	} else {
		log.Debugf("Get request url: %s", r.endpoint.String())
	}

	response, err := r.client.Get(r.endpoint.String())
	if err != nil {
		return nil, -1, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, err
	}

	return body, response.StatusCode, nil
}

func (r *RestService) GetFile(urlPath string, directory string) (string, error) {
	r.endpoint.Path = urlPath

	if os.Getenv("KDEPLOY_DRYRUN") == "1" {
		log.Infof("Get file request url: %s destination: %s", r.endpoint.String(), directory)
	} else {
		log.Debugf("Get file request url: %s destination: %s", r.endpoint.String(), directory)
	}

	response, err := r.client.Get(r.endpoint.String())
	if err != nil {
		return "", err
	}

	if response.StatusCode >= 300 {
		return "", errors.New(fmt.Sprintf("Obtained %d response code for downloading file", response.StatusCode))
	}

	slice := strings.Split(urlPath, "/")
	fileName := slice[len(slice)-1:][0]

	filePath := path.Join(directory, fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer out.Close()
	io.Copy(out, response.Body)
	return filePath, nil
}
