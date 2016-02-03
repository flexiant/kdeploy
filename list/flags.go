package list

import "github.com/codegangsta/cli"

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
