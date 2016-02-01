package list

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// Flags builds a spec of the flags available for the command
func Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "List Kubeware, SVC, RC",
		},
		cli.BoolFlag{
			Name:  "services, svc",
			Usage: "List Services",
		},
		cli.BoolFlag{
			Name:  "controllers, rc",
			Usage: "List Replica Controllers",
		},
	}
}

// PrepareFlags processes the flags
func PrepareFlags(c *cli.Context) error {
	return nil
}

// struct representing an item to be listed
type Kube struct {
	Services    []map[string]interface{}
	Controllers []map[string]interface{}
}

type controllerList struct {
	Items []controllerItem
}

type controllerItem struct {
	Metadata struct {
		Name   string
		Labels map[string]string
	}
	Spec struct {
		Replicas int
	}
	Status struct {
		Replicas int
	}
}

// struct for unmarshalling json representations of services
type serviceList struct {
	Items []serviceItem
}

type serviceItem struct {
	Metadata struct {
		CreationTimestamp string
		Name              string
		Labels            map[string]string
	}
	Spec struct {
		ClusterIP string
	}
	Status struct {
		LoadBalancer struct {
			Ingress []struct {
				Hostname string
			}
		}
	}
}

// CmdList implements 'list' command
func CmdList(c *cli.Context) {
	var fqdns []string
	up := 0
	// Get all services to extract their kubeware labels
	serviceList, err := getServices()
	utils.CheckError(err)
	// Get all controllers to extract their kubeware labels
	controllersList, err := getControllers()
	utils.CheckError(err)
	// build the list to be printed
	kubeList := buildKubeList(serviceList, controllersList)

	if len(kubeList) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 15, 1, 3, ' ', 0)

		if c.Bool("all") || (!c.Bool("services") && !c.Bool("controllers")) {
			fmt.Fprintln(w, "KUBEWARE\tSVC\tRC\tUP\tFQDN\r")
			for kubewareName, kubeware := range kubeList {
				for _, service := range kubeware.Services {
					if service["ExternalFQDN"] != nil {
						fqdns = append(fqdns, service["ExternalFQDN"].(string))
					}
				}
				for _, controller := range kubeware.Controllers {
					up = up + controller["Up"].(int)
				}
				if len(kubeware.Services) > 0 {
					up = up / len(kubeware.Services)
				}

				if len(fqdns) > 0 {
					fmt.Fprintf(w, "%s\t%d\t%d\t%d%%\t%s\n", kubewareName, len(kubeware.Services), len(kubeware.Controllers), up, strings.Join(fqdns, ","))
				} else {
					fmt.Fprintf(w, "%s\t%d\t%d%%\t%d\n", kubewareName, len(kubeware.Services), up, len(kubeware.Controllers))
				}

			}
		}
		if c.Bool("all") {
			fmt.Fprintf(w, "\n")
		}
		if c.Bool("all") || c.Bool("services") {
			fmt.Fprintln(w, "KUBEWARE\tSVC\tINTERNAL IP\tFQDN\r")
			for kubewareName, kubeware := range kubeList {
				for _, service := range kubeware.Services {
					if service["ExternalFQDN"] != nil {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", kubewareName, service["Name"], service["ClusterIP"], service["ExternalFQDN"])
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\n", kubewareName, service["Name"], service["ClusterIP"])
					}

				}
			}
		}
		if c.Bool("all") {
			fmt.Fprintf(w, "\n")
		}
		if c.Bool("all") || c.Bool("controllers") {
			fmt.Fprintln(w, "KUBEWARE\tRC\tREPLICAS\tUP\r")
			for kubewareName, kubeware := range kubeList {
				for _, controller := range kubeware.Controllers {
					fmt.Fprintf(w, "%s\t%s\t%d\t%d%%\n", kubewareName, controller["Name"], controller["Replicas"], controller["Up"])
				}
			}
		}
		w.Flush()
	} else {
		fmt.Printf("No Kubeware's deployed")
	}
}

func buildKubeList(svcList *serviceList, rcList *controllerList) map[string]Kube {
	kmap := map[string]Kube{}
	for _, service := range svcList.Items {
		if kubeName, ok := service.Metadata.Labels["kubeware"]; ok {
			// check if kube already in map
			if _, ok := kmap[kubeName]; !ok {
				// if not, create it
				kmap[kubeName] = Kube{}
			}
			// add the service to the kube's list of services
			kube := kmap[kubeName]
			kube.Services = append(kube.Services, buildServiceRecord(service))
			kmap[kubeName] = kube
		}
	}
	for _, controller := range rcList.Items {
		if kubeName, ok := controller.Metadata.Labels["kubeware"]; ok {
			// check if kube already in map
			if _, ok := kmap[kubeName]; !ok {
				// if not, create it
				kmap[kubeName] = Kube{}
			}
			// add the controller to the kube's list of controllers
			kube := kmap[kubeName]
			kube.Controllers = append(kube.Controllers, buildControllerRecord(controller))
			kmap[kubeName] = kube
		}
	}
	return kmap
}

func buildServiceRecord(service serviceItem) map[string]interface{} {
	sr := map[string]interface{}{}
	sr["Name"] = service.Metadata.Name
	sr["CreationDate"] = service.Metadata.CreationTimestamp
	sr["ClusterIP"] = service.Spec.ClusterIP
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		sr["ExternalFQDN"] = service.Status.LoadBalancer.Ingress[0].Hostname
	}
	return sr
}

func buildControllerRecord(controller controllerItem) map[string]interface{} {
	sr := map[string]interface{}{}
	sr["Name"] = controller.Metadata.Name
	sr["Replicas"] = controller.Status.Replicas
	sr["Up"] = (controller.Status.Replicas / controller.Spec.Replicas) * 100
	return sr
}

func getServices() (*serviceList, error) {
	kube, err := webservice.NewKubeClient()
	if err != nil {
		return nil, err
	}
	jsonServices, err := kube.GetServices(nil)
	if err != nil {
		return nil, err
	}
	return unmarshalServices(jsonServices)
}

func getControllers() (*controllerList, error) {
	kube, err := webservice.NewKubeClient()
	if err != nil {
		return nil, err
	}
	jsonControllers, err := kube.GetControllers(nil)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("%s", jsonControllers)
	return unmarshalControllers(jsonControllers)
}

func unmarshalControllers(jsonStr string) (*controllerList, error) {
	var rl controllerList
	err := json.Unmarshal([]byte(jsonStr), &rl)
	if err != nil {
		return nil, err
	}
	return &rl, nil
}

func unmarshalServices(jsonStr string) (*serviceList, error) {
	var sl serviceList
	err := json.Unmarshal([]byte(jsonStr), &sl)
	if err != nil {
		return nil, err
	}
	return &sl, nil
}
