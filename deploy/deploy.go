package deploy

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/digger"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdDeploy implements the 'deploy' command
func CmdDeploy(c *cli.Context) {
	// utils.CheckRequiredFlags(c, []string{"attribute", "kubeware", "namespace"})
	localKubePath, err := webservice.FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	metadata := template.ParseMetadata(localKubePath)
	defaults, err := metadata.AttributeDefaults()
	utils.CheckError(err)
	// build attributes merging "role list" to defaults
	log.Debugf("Building attributes")
	attributes := template.BuildAttributes(c.String("attribute"), defaults)
	// get list of services and parse each one
	log.Debugf("Parsing services")
	servicesSpecs, err := metadata.ParseServices(attributes)
	utils.CheckError(err)
	// get list of replica controllers and parse each one
	log.Debugf("Parsing controllers")
	controllersSpecs, err := metadata.ParseControllers(attributes)
	utils.CheckError(err)
	// set version labels on RCs
	err = setKubeVersionOnRCs(controllersSpecs)
	utils.CheckError(err)
	// creates Kubernetes client
	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)
	// create each of the services
	log.Debugf("Creating services")
	err = kubernetes.CreateServices(utils.Values(servicesSpecs))
	utils.CheckError(err)
	// create each of the controllers
	log.Debugf("Creating controllers")
	err = kubernetes.CreateReplicaControllers(utils.Values(controllersSpecs))
	utils.CheckError(err)

	fmt.Printf("Kubeware %s from %s has been deployed", metadata.Name, os.Getenv("KDEPLOY_KUBEWARE"))
}

func setKubeVersionOnRCs(rcs map[string]string) error {
	for name, rcjson := range rcs {
		var rc map[string]interface{}
		err := json.Unmarshal([]byte(rcjson), &rc)
		if err != nil {
			return fmt.Errorf("could not unmarshal rc '%s': %v", name, err)
		}
		err = setKubeVersionOnRC(rc)
		if err != nil {
			return err
		}
		newJSON, err := json.Marshal(rc)
		if err != nil {
			log.Debugf("could not marshal modified rc: %v", err)
			return fmt.Errorf("could not marshal modified rc: %v", err)
		}
		rcs[name] = string(newJSON)
	}
	return nil
}

func setKubeVersionOnRC(rc map[string]interface{}) error {
	kv, err := extractKubeVersion(rc)
	if err != nil {
		return err
	}
	// set at pod template
	path := []string{"spec", "template", "metadata", "labels"}
	m := rc
	for _, s := range path {
		if m[s] == nil {
			m[s] = map[string]interface{}{}
		}
		m = m[s].(map[string]interface{})
	}
	m["kubeware"] = kv
	// set at label selector
	path = []string{"spec", "selector"}
	m = rc
	for _, s := range path {
		if m[s] == nil {
			m[s] = map[string]interface{}{}
		}
		m = m[s].(map[string]interface{})
	}
	m["kubeware"] = kv

	return nil
}

func extractKubeVersion(rc map[string]interface{}) (string, error) {
	d, err := digger.NewMapDigger(rc)
	if err != nil {
		return "", err
	}
	k, err := d.GetString("metadata/labels/kubeware")
	if err != nil {
		return "", err
	}
	v, err := d.GetString("metadata/labels/kubeware-version")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", k, v), nil
}
