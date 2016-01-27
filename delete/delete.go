package delete

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

func Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "kubeware, k",
			Usage:  "Kubeware path",
			Value:  "https://github.com/flexiant/kubeware-guestbook",
			EnvVar: "KDEPLOY_KUBEWARE",
		},
		cli.BoolFlag{
			Name:   "dry-run, d",
			Usage:  "Dry Run of Deploy used for debuging options",
			EnvVar: "KDEPLOY_DRYRUN",
		},
	}
}

func PrepareFlags(c *cli.Context) error {
	if c.String("attribute") != "" {
		os.Setenv("KDEPLOY_ATTRIBUTE", c.String("attribute"))
	}

	if c.String("kubeware") != "" {
		os.Setenv("KDEPLOY_KUBEWARE", c.String("kubeware"))
	}

	if c.Bool("dry-run") {
		os.Setenv("KDEPLOY_DRYRUN", "1")
	}

	return nil
}

// CmdDelete implements 'delete' command
func CmdDelete(c *cli.Context) {
	localKubePath, err := FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	md := template.ParseMetadata(localKubePath)
	utils.CheckError(err)

	// get services which are currently deployed as part of the kube
	serviceNames, err := getDeployedServicesForKubeware(md)
	utils.CheckError(err)

	fmt.Println("Services: ")
	for _, s := range serviceNames {
		fmt.Println(" - " + s)
	}

	// get controllers which are currently deployed as part of the kube
	// controllers, err := getDeployedControllersForKubeware(md)
	// utils.CheckError(err)

	// delete them
	// err := deleteServices(services)
	// utils.CheckError(err)

	// delete them
	// err := deleteControllers(controllers)
	// utils.CheckError(err)

}

func getDeployedServicesForKubeware(m template.Metadata) ([]string, error) {
	kube, err := webservice.NewKubeClient()
	utils.CheckError(err)

	labelSelector := map[string]string{
		"kubeware": m.Name,
	}
	servicesJson, err := kube.GetServices(labelSelector)
	utils.CheckError(err)

	var svcList ServiceList
	err = json.Unmarshal([]byte(servicesJson), &svcList)
	utils.CheckError(err)

	names := []string{}
	for _, s := range svcList.Items {
		names = append(names, s.Metadata.Name)
	}
	return names, nil
}

type ServiceList struct {
	Items []Service
}

type Service struct {
	Metadata struct {
		Name string
	}
}

// func getDeployedServicesForKubeware(m Metadata) ([]string, error) {
// 	kube, err := webservice.NewKubeClient()
// 	if err != nil {
// 		return nil, err
// 	}
// 	labelSelector := map[string]string{"kubeware": m.Name}
// 	services, err := kube.GetServices(labelSelector)
// 	if err != nil {
// 		return nil, err
// 	}
// 	serviceNames := []string{}
// 	for _, s := range services {
// 		serviceNames = append(serviceNames, s["name"])
// 	}
// 	return serviceNames, nil
// }

func FetchKubeFromURL(kubeURL string) (string, error) {
	kubewareUrl, err := url.Parse(kubeURL)
	if err != nil {
		return "", err
	}

	if kubewareUrl.Host != "github.com" {
		return "", errors.New("We currently only support Github urls")
	}

	path := strings.Split(kubewareUrl.Path, "/")
	kubewareName := path[2]

	newPath := append([]string{""}, path[1], path[2], "archive", "master.zip")

	kubewareUrl.Path = strings.Join(newPath, "/")

	client, err := webservice.NewSimpleWebClient(kubewareUrl.String())
	utils.CheckError(err)

	tmpDir, err := ioutil.TempDir("", "kdeploy")
	utils.CheckError(err)

	zipFileLocation, err := client.GetFile(kubewareUrl.Path, tmpDir)
	utils.CheckError(err)

	err = utils.Unzip(zipFileLocation, tmpDir)
	utils.CheckError(err)

	return fmt.Sprintf("%s/%s-master/", tmpDir, kubewareName), nil
}
