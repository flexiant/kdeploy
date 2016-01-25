package rest

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/flexiant/kdeploy/config"
)

type RestService struct {
	client   *http.Client
	endpoint string
}

func NewRestService(config config.Config) (*RestService, error) {
	client, err := httpClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client: %v", err)
	}
	return &RestService{client, extractEndpoint(config)}, nil
}

func extractEndpoint(config config.Config) string {
	endpoint := config.Connection.APIEndpoint
	if strings.HasSuffix(endpoint, "/") {
		return endpoint
	}
	return endpoint + "/"
}

func httpClient(config config.Config) (*http.Client, error) {
	// load client certificate
	cert, err := tls.LoadX509KeyPair(config.Connection.Cert, config.Connection.Key)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair: %v", err)
	}
	// load CA file to verify server
	caPool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile(config.Connection.CACert)
	if err != nil {
		return nil, fmt.Errorf("could not load CA file: %v", err)
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

func (r *RestService) Post(path string, json []byte) ([]byte, int, error) {
	output := strings.NewReader(string(json))
	log.Printf("debug: post request path: %s , body: \n%s\n", r.endpoint+path, string(prettyprint(json)))
	response, err := r.client.Post(r.endpoint+path, "application/json", output)
	if err != nil {
		return nil, -1, fmt.Errorf("error on http request (POST %v): %v", r.endpoint+path, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("error reading http request body: %v", err)
	}

	return body, response.StatusCode, err
}

func (r *RestService) Delete(path string) ([]byte, int, error) {
	request, err := http.NewRequest("DELETE", r.endpoint+path, nil)
	if err != nil {
		return nil, -1, fmt.Errorf("error creating http request (DELETE %v): %v", r.endpoint+path, err)
	}
	response, err := r.client.Do(request)
	if err != nil {
		return nil, -1, fmt.Errorf("error executing http request (DELETE %v): %v", r.endpoint+path, err)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("error reading http request body: %v", err)
	}

	return body, response.StatusCode, nil
}

func (r *RestService) Get(path string) ([]byte, int, error) {
	response, err := r.client.Get(r.endpoint + path)
	if err != nil {
		return nil, -1, fmt.Errorf("error on http request (GET %v): %v", r.endpoint+path, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("error reading http request body: %v", err)
	}

	return body, response.StatusCode, nil
}
