package upgrade

import (
	"os"

	"github.com/codegangsta/cli"
)

// Flags builds a spec of the flags available for the command
func Flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "attribute, a",
			Usage:  "Attribute List",
			EnvVar: "KDEPLOY_ATTRIBUTE",
		},
		cli.StringFlag{
			Name:   "kubeware, k",
			Usage:  "Kubeware path",
			EnvVar: "KDEPLOY_KUBEWARE",
		},
		cli.StringFlag{
			Name:   "namespace, n",
			Usage:  "Namespace which to deploy Kubeware",
			Value:  "default",
			EnvVar: "KDEPLOY_NAMESPACE",
		},
		cli.BoolFlag{
			Name:   "dry-run, d",
			Usage:  "Dry Run of Deploy used for debugging options",
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

	os.Setenv("KDEPLOY_NAMESPACE", c.String("namespace"))

	return nil
}
