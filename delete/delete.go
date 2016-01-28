package delete

import (
	"encoding/json"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// Flags builds a spec of the flags available for the command
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

// PrepareFlags processes the flags
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
	localKubePath, err := webservice.FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	md := template.ParseMetadata(localKubePath)
	utils.CheckError(err)

	// get services which are currently deployed as part of the kube
	serviceNames, err := getDeployedServicesForKubeware(md)
	utils.CheckError(err)
	log.Debugf("Services: %v", serviceNames)

	// get controllers which are currently deployed as part of the kube
	controllerNames, err := getDeployedControllersForKubeware(md)
	utils.CheckError(err)
	log.Debugf("Controllers: %v", controllerNames)

	// delete them
	err = deleteServices(serviceNames)
	utils.CheckError(err)

	// delete them
	err = deleteControllers(controllerNames)
	utils.CheckError(err)

}

func deleteServices(serviceNames []string) error {
	kube, err := webservice.NewKubeClient()
	if err != nil {
		return err
	}
	for _, s := range serviceNames {
		err := kube.DeleteService("default", s)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteControllers(controllerNames []string) error {
	kube, err := webservice.NewKubeClient()
	if err != nil {
		return err
	}
	for _, c := range controllerNames {
		err := kube.DeleteReplicationController("default", c)
		if err != nil {
			return err
		}
	}
	return nil
}

func getDeployedServicesForKubeware(m template.Metadata) ([]string, error) {

	type ServiceList struct {
		Items []struct {
			Metadata struct {
				Name string
			}
		}
	}

	kube, err := webservice.NewKubeClient()
	utils.CheckError(err)

	labelSelector := map[string]string{
		"kubeware": m.Name,
	}
	servicesJSON, err := kube.GetServices(labelSelector)
	utils.CheckError(err)

	var svcList ServiceList
	err = json.Unmarshal([]byte(servicesJSON), &svcList)
	utils.CheckError(err)

	names := []string{}
	for _, s := range svcList.Items {
		names = append(names, s.Metadata.Name)
	}
	return names, nil
}

func getDeployedControllersForKubeware(m template.Metadata) ([]string, error) {

	type ControllersList struct {
		Items []struct {
			Metadata struct {
				Name string
			}
		}
	}

	kube, err := webservice.NewKubeClient()
	utils.CheckError(err)

	labelSelector := map[string]string{
		"kubeware": m.Name,
	}
	controllersJSON, err := kube.GetControllers(labelSelector)
	utils.CheckError(err)

	var rcList ControllersList
	err = json.Unmarshal([]byte(controllersJSON), &rcList)
	utils.CheckError(err)

	names := []string{}
	for _, c := range rcList.Items {
		names = append(names, c.Metadata.Name)
	}
	return names, nil
}
