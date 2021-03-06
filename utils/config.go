package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// Connection configuration
type Connection struct {
	APIEndpoint string
	CACert      string
	Cert        string
	Key         string
	Insecure    bool
}

// Config for kdeploy
type Config struct {
	Connection Connection
	Path       string
}

var cachedConfig *Config

// GetConfig returns cached config
func GetConfig() (*Config, error) {

	if cachedConfig != nil {
		return cachedConfig, nil
	}

	return nil, errors.New("Configuration must be initialized")
	//return fmt.Errorf("Configuration must be initialized", nil)
}

// PreReadConfig reads configuration before full arguments
func PreReadConfig() *Config {

	config, _ := ReadConfig()

	// if config is nil fill an structure with empty values
	if config == nil {
		config = &Config{
			Connection: Connection{},
		}
	}

	if os.Getenv("KUBERNETES_CA_CERT") != "" {
		config.Connection.CACert = os.Getenv("KUBERNETES_CA_CERT")
	}
	if os.Getenv("KUBERNETES_CLIENT_CERT") != "" {
		config.Connection.Cert = os.Getenv("KUBERNETES_CLIENT_CERT")
	}
	if os.Getenv("KUBERNETES_CLIENT_KEY") != "" {
		config.Connection.Key = os.Getenv("KUBERNETES_CLIENT_KEY")
	}
	if os.Getenv("KUBERNETES_ENDPOINT") != "" {
		config.Connection.APIEndpoint = os.Getenv("KUBERNETES_ENDPOINT")
	}

	return config
}

// func CachedConfig() (*Config, error) {
// 	count := 0
// 	var config Config
//
// 	if os.Getenv("KUBERNETES_CA_CERT") != "" {
// 		config.Connection.CACert = os.Getenv("KUBERNETES_CA_CERT")
// 		count = count + 1
// 	}
// 	if os.Getenv("KUBERNETES_CLIENT_CERT") != "" {
// 		config.Connection.Cert = os.Getenv("KUBERNETES_CLIENT_CERT")
// 		count = count + 1
// 	}
// 	if os.Getenv("KUBERNETES_CLIENT_KEY") != "" {
// 		config.Connection.Key = os.Getenv("KUBERNETES_CLIENT_KEY")
// 		count = count + 1
// 	}
// 	if os.Getenv("KUBERNETES_ENDPOINT") != "" {
// 		config.Connection.APIEndpoint = os.Getenv("KUBERNETES_ENDPOINT")
// 		count = count + 1
// 	}
// 	// TODO delete
// 	// if os.Getenv("KDEPLOY_CONFIG") != "" {
// 	// 	config.Path = os.Getenv("KDEPLOY_CONFIG")
// 	// 	count = count + 1
// 	// }
// 	if os.Getenv("KUBECONFIG") != "" {
// 		config.Path = os.Getenv("KUBECONFIG")
// 		count = count + 1
// 	}
//
// 	if count == 5 {
// 		return &config, nil
// 	}
// 	return nil, errors.New("Please check that all parameters are set in your configuration")
//
// }

func ReadConfig() (*Config, error) {

	if os.Getenv("KUBECONFIG") != "" {
		cfg, err := readKubeConfigFrom(os.Getenv("KUBECONFIG"))
		if err != nil {
			return cfg, nil
		}
	}

	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("Couldn't get home dir for current user: %s", err.Error())
	}

	cfg, err := readKubeConfigFrom(filepath.Join(home, ".kube/config"))
	if err == nil {
		return cfg, nil
	}

	return nil, nil
}

// readKubeConfigFrom reads kubeconfig and use it to fill a kdeploy config
func readKubeConfigFrom(path string) (*Config, error) {
	log.Debugf("readKubeConfigFrom %s", path)
	// Read kubeconfig
	if !FileExists(path) {
		return nil, fmt.Errorf("Could not locate file %s", path)
	}

	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var kc KubeConfig

	err = yaml.Unmarshal(configBytes, &kc)
	if err != nil {
		return nil, err
	}

	// Create kdeploy config
	config := readConfigFromKubeConfig(&kc)

	// add path
	config.Path = path

	log.Debugf("returning config from kubeconfig: %+v", config)
	return config, nil
}

// readConfigFromKubeConfig extract Kubeconfig fields to build a Kdeploy config structure
func readConfigFromKubeConfig(kc *KubeConfig) *Config {
	log.Debugf("readConfigFromKubeConfig %+v", kc)
	// currentContext := kc.CurrentContext
	var userID, clusterID string
	config := &Config{}

	for _, context := range kc.Contexts {
		if context.Name == kc.CurrentContext {
			userID = context.Context.User
			clusterID = context.Context.Cluster
			break
		}
	}

	if userID != "" {
		for _, user := range kc.Users {
			if user.Name == userID {
				config.Connection.Cert = user.User.ClientCertificate
				config.Connection.Key = user.User.ClientKey
				break
			}

		}
	}

	if clusterID != "" {
		for _, cluster := range kc.Clusters {
			if cluster.Name == clusterID {
				config.Connection.APIEndpoint = cluster.Cluster.Server
				config.Connection.CACert = cluster.Cluster.CertificateAuthority
				config.Connection.Insecure = cluster.Cluster.InsecureSkipTLSVerify
			}
		}
	}

	return config
}

func ReadConfigFrom(path string) (*Config, error) {

	cfgFile, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if FileExists(cfgFile) {
		configBytes, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return nil, err
		}
		var cfg Config
		err = yaml.Unmarshal(configBytes, &cfg)
		if err != nil {
			return nil, err
		}

		if os.Getenv("KUBERNETES_CA_CERT") != "" {
			cfg.Connection.CACert = os.Getenv("KUBERNETES_CA_CERT")
		}
		if os.Getenv("KUBERNETES_CLIENT_CERT") != "" {
			cfg.Connection.Cert = os.Getenv("KUBERNETES_CLIENT_CERT")
		}
		if os.Getenv("KUBERNETES_CLIENT_KEY") != "" {
			cfg.Connection.Key = os.Getenv("KUBERNETES_CLIENT_KEY")
		}
		if os.Getenv("KUBERNETES_ENDPOINT") != "" {
			cfg.Connection.APIEndpoint = os.Getenv("KUBERNETES_ENDPOINT")
		}

		cfg.Path = cfgFile
		return &cfg, nil
	}
	return nil, errors.New(fmt.Sprintf("Could not locate file %s", path))
}

func InitializeConfig(c *cli.Context) error {
	log.Debug("InitializeConfig")
	// overwrite with environment/arguments vars
	if overwKdeployConfig := c.String("kdeploy-config"); overwKdeployConfig != "" {
		cachedConfig.Path = overwKdeployConfig
		cachedConfig, _ = ReadConfigFrom(overwKdeployConfig)
	} else {
		// read from path
		cachedConfig, _ = ReadConfig()
	}

	// if config is nil fill an structure with empty values
	if cachedConfig == nil {
		cachedConfig = &Config{
			Connection: Connection{},
		}
	}

	parameters := false

	// overwrite with environment/arguments vars
	if overwCaCert := c.String("ca-cert"); overwCaCert != "" {
		cachedConfig.Connection.CACert = overwCaCert
	}

	// overwrite with environment/arguments vars
	if overwClientCert := c.String("client-cert"); overwClientCert != "" {
		cachedConfig.Connection.Cert = overwClientCert
	}

	// overwrite with environment/arguments vars
	if overwClientKey := c.String("client-key"); overwClientKey != "" {
		cachedConfig.Connection.Key = overwClientKey
	}

	// overwrite with environment/arguments vars
	if overwKubernetesEndpoint := c.String("kubernetes-endpoint"); overwKubernetesEndpoint != "" {
		cachedConfig.Connection.APIEndpoint = overwKubernetesEndpoint
	}

	// insecure connection flag
	cachedConfig.Connection.Insecure = cachedConfig.Connection.Insecure || c.Bool("insecure")

	if cachedConfig.Connection.APIEndpoint == "" {
		log.Warn("Please use parameter --kubernetes-endpoint")
		parameters = true
	}
	if cachedConfig.Connection.Cert == "" {
		log.Warn("Please use parameter --client-cert")
		parameters = true
	}
	if cachedConfig.Connection.Key == "" {
		log.Warn("Please use parameter --client-key")
		parameters = true
	}

	if parameters {
		fmt.Printf("\n")
		cli.ShowCommandHelp(c, c.Command.Name)
		os.Exit(2)
	}

	os.Setenv("KUBERNETES_CA_CERT", cachedConfig.Connection.CACert)
	os.Setenv("KUBERNETES_CLIENT_CERT", cachedConfig.Connection.Cert)
	os.Setenv("KUBERNETES_CLIENT_KEY", cachedConfig.Connection.Key)
	os.Setenv("KUBERNETES_ENDPOINT", cachedConfig.Connection.APIEndpoint)
	os.Setenv("KUBECONFIG", cachedConfig.Path)

	// TODO delete
	// os.Setenv("KDEPLOY_CONFIG", config.Path)

	return nil
}
