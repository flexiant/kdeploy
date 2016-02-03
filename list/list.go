package list

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

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
	serviceList, err := kubernetes.GetServices()
	utils.CheckError(err)
	// Get all controllers to extract their kubeware labels
	controllersList, err := kubernetes.GetControllers()
	utils.CheckError(err)
	// build the list to be printed
	kubeList := models.BuildKubeList(serviceList, controllersList)

	if len(kubeList) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)

		if c.Bool("all") || (!c.Bool("services") && !c.Bool("controllers")) {
			fmt.Fprintln(w, "KUBEWARE\tNAMESPACE\tVERSION\tSVC\tRC\tUP\tFQDN\r")
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
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d%%\t%s\n", kubewareName, kubeware.GetNamespace(), kubeware.GetVersion(), len(kubeware.Services), len(kubeware.Controllers), up, strings.Join(fqdns, ","))
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d%%\t%d\n", kubewareName, kubeware.GetNamespace(), kubeware.GetVersion(), len(kubeware.Services), up, len(kubeware.Controllers))
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
			for kubewareName, kubeware := range kubeList {
				for _, service := range kubeware.Services {
					if service["ExternalFQDN"] != nil {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", kubewareName, service["Namespace"], service["Name"], service["ClusterIP"], service["ExternalFQDN"])
					} else {
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", kubewareName, service["Namespace"], service["Name"], service["ClusterIP"])
					}

				}
			}
		}
		if c.Bool("all") {
			fmt.Fprintf(w, "\n")
		}
		if c.Bool("all") || c.Bool("controllers") {
			fmt.Fprintln(w, "KUBEWARE\tNAMESPACE\tRC\tREPLICAS\tUP\r")
			for kubewareName, kubeware := range kubeList {
				for _, controller := range kubeware.Controllers {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d%%\n", kubewareName, controller["Namespace"], controller["Name"], controller["Replicas"], controller["Up"])
				}
			}
		}
		w.Flush()
	} else {
		fmt.Printf("No Kubeware's deployed")
	}
}
