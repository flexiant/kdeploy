package main

import (
	"fmt"

	"github.com/mafraba/digger"
	"github.com/mafraba/kdeploy/kubeclient"
	"github.com/mafraba/kdeploy/metadata"
)

func main() {
	md := metadata.ParseMetadata("./samples/")
	defaults, err := md.AttributeDefaults()
	if err != nil {
		panic(err)
	}
	// build attributes merging "role list" to defaults
	attributes := buildAttributes(defaults)
	// get list of services and parse each one
	servicesSpecs, err := md.ParseServices(attributes)
	if err != nil {
		panic(err)
	}
	// get list of replica controllers and parse each one
	controllersSpecs, err := md.ParseControllers(attributes)
	if err != nil {
		panic(err)
	}

	// get services just to check API availability
	// getServices()

	// create each of the services
	err = createServices(servicesSpecs)
	if err != nil {
		panic(err)
	}
	// create each of the controllers
	err = createControllers(controllersSpecs)
	if err != nil {
		panic(err)
	}
}

func getServices() {
	kube, _ := kubeclient.NewKubeClient()
	services, _ := kube.GetServices()
	fmt.Println("services: ")
	fmt.Println(services)
}

func buildAttributes(defaults digger.Digger) digger.Digger {
	roleList := `
	{
		"rc" : {
			"frontend" : {
				"number" : 3,
				"container": {
					"name": "contname",
					"image": "somerepo/someimage",
					"version": "latest",
					"port": 8080
				}
			}
		}
	}
	`
	roleListDigger, err := digger.NewJSONDigger([]byte(roleList))
	if err != nil {
		panic(err)
	}
	attributes, err := digger.NewMultiDigger(
		roleListDigger,
		defaults,
	)
	if err != nil {
		panic(err)
	}

	return attributes
}

func createServices(svcSpecs []string) error {
	kube, err := kubeclient.NewKubeClient()
	if err != nil {
		return fmt.Errorf("error creating kube client: %v", err)
	}
	for _, spec := range svcSpecs {
		_, err = kube.CreateService("default", []byte(spec))
		if err != nil {
			return fmt.Errorf("error creating services: %v", err)
		}
	}
	return nil
}

func createControllers(rcSpecs []string) error {
	kube, err := kubeclient.NewKubeClient()
	if err != nil {
		return fmt.Errorf("error creating kube client: %v", err)
	}
	for _, spec := range rcSpecs {
		_, err = kube.CreateReplicaController("default", []byte(spec))
		if err != nil {
			return fmt.Errorf("error creating replication controllers: %v", err)
		}
	}
	return nil
}
