package show

import (
	"os"

	"github.com/codegangsta/cli"
)

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
	}
}

func PrepareFlags(c *cli.Context) error {
	if c.String("attribute") != "" {
		os.Setenv("KDEPLOY_ATTRIBUTE", c.String("attribute"))
	}

	if c.String("kubeware") != "" {
		os.Setenv("KDEPLOY_KUBEWARE", c.String("kubeware"))
	}

	return nil
}
