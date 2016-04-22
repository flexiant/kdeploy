package list

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/models"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
)

// CmdList implements 'list' command
func CmdList(c *cli.Context) {
	var fqdns []string
	up := 0
	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)
	// Get all services to extract their kubeware labels
	log.Debug("Get all services to extract their kubeware labels ...")
	serviceList, err := kubernetes.GetServices()
	utils.CheckError(err)
	// Get all controllers to extract their kubeware labels
	log.Debug("Get all controllers to extract their kubeware labels ...")
	controllersList, err := kubernetes.GetControllers()
	utils.CheckError(err)
	// build the list to be printed
	kubeList := models.BuildKubeList(serviceList, controllersList)

	if len(kubeList) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)

		if c.Bool("all") || (!c.Bool("services") && !c.Bool("controllers")) {
			fmt.Fprintln(w, "KUBEWARE\tNAMESPACE\tVERSION\tSVC\tRC\tUP\tFQDN\r")
			for _, kubeware := range kubeList {
				for _, service := range kubeware.Services {
					if service.GetFQDN() != "" {
						fqdns = append(fqdns, service.GetFQDN())
					}
				}
				for _, controller := range kubeware.ReplicaControllers {
					up = up + controller.GetUpStats()
				}
				if len(kubeware.Services) > 0 {
					up = up / len(kubeware.Services)
				}

				if len(fqdns) > 0 {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d%%\t%s\n", kubeware.GetKube(), kubeware.GetNamespace(), kubeware.GetVersion(), len(kubeware.Services), len(kubeware.ReplicaControllers), up, strings.Join(fqdns, ","))
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d%%\t%d\n", kubeware.GetKube(), kubeware.GetNamespace(), kubeware.GetVersion(), len(kubeware.Services), up, len(kubeware.ReplicaControllers))
				}
				up = 0
				fqdns = []string{}
			}
		}
		if c.Bool("all") {
			fmt.Fprintf(w, "\n")
		}
		if c.Bool("all") || c.Bool("services") {
			fmt.Fprintln(w, "KUBEWARE\tNAMESPACE\tSVC\tINTERNAL IP\tFQDN\r")
			for _, kubeware := range kubeList {
				for _, service := range kubeware.Services {
					if service.GetFQDN() != "" {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", kubeware.GetKube(), service.GetNamespace(), service.GetName(), service.GetInternalIp(), service.GetFQDN())
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", kubeware.GetKube(), service.GetNamespace(), service.GetName(), service.GetInternalIp())
					}
				}
			}
		}
		if c.Bool("all") {
			fmt.Fprintf(w, "\n")
		}
		if c.Bool("all") || c.Bool("controllers") {
			fmt.Fprintln(w, "KUBEWARE\tNAMESPACE\tRC\tREPLICAS\tUP\r")
			for _, kubeware := range kubeList {
				for _, replicaController := range kubeware.ReplicaControllers {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d%%\n", kubeware.GetKube(), kubeware.GetNamespace(), replicaController.GetName(), replicaController.GetReplicas(), replicaController.GetUpStats())
				}
			}
		}
		w.Flush()
	} else {
		log.Infof("No Kubeware deployed")
	}
}
