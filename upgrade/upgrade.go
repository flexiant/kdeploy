package upgrade

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/template"
	"github.com/flexiant/kdeploy/upgrade/strategies"
	"github.com/flexiant/kdeploy/utils"
	"github.com/flexiant/kdeploy/webservice"
	"github.com/hashicorp/go-version"
)

// CmdUpgrade implements 'upgrade' command
func CmdUpgrade(c *cli.Context) {
	localKubePath, err := webservice.FetchKubeFromURL(os.Getenv("KDEPLOY_KUBEWARE"))
	utils.CheckError(err)

	log.Debugf("Going to parse kubeware in %s", localKubePath)

	md := template.ParseMetadata(localKubePath)
	utils.CheckError(err)

	kubernetes, err := webservice.NewKubeClient()
	utils.CheckError(err)

	namespace := os.Getenv("KDEPLOY_NAMESPACE")
	// labelSelector := fmt.Sprintf("kubeware=%s,kubeware-version=%s", md.Name, md.Version)

	// Check if kubeware already installed, error if it's not
	v, err := kubernetes.FindDeployedKubewareVersion(namespace, md.Name)
	utils.CheckError(err)
	log.Infof("Found version %s of %s.%s", v, namespace, md.Name)

	// Check if equal or newer version already exists, error if so
	deployedVersion, err := version.NewVersion(v)
	utils.CheckError(err)
	upgradeVersion, err := version.NewVersion(md.Version)
	utils.CheckError(err)
	if upgradeVersion.LessThan(deployedVersion) {
		log.Fatalf("Can not upgrade to version '%s' since version '%s' is already deployed", md.Version, v)
	}

	// build attributes merging "role list" to defaults
	log.Debugf("Building attributes")
	defaults, err := md.AttributeDefaults()
	utils.CheckError(err)
	attributes := template.BuildAttributes(c.String("attribute"), defaults)

	// get services and parse each one
	log.Debugf("Parsing services")
	servicesSpecs, err := md.ParseServices(attributes)
	utils.CheckError(err)

	// get replica controllers and parse each one
	log.Debugf("Parsing controllers")
	controllersSpecs, err := md.ParseControllers(attributes)
	utils.CheckError(err)

	// upgStrategy := upgradeStrategies.RecreateAllStrategy(kubernetes)
	upgStrategy := upgradeStrategies.RollRcPatchSvcStrategy(kubernetes, 1)
	upgStrategy.Upgrade(namespace, servicesSpecs, controllersSpecs)
}
