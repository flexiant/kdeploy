package main

import (
	"os"
  "flag"

	"github.com/golang/glog"
	"github.com/codegangsta/cli"
	"github.com/flexiant/kdeploy/delete"
	"github.com/flexiant/kdeploy/deploy"
	"github.com/flexiant/kdeploy/list"
	"github.com/flexiant/kdeploy/show"
	"github.com/flexiant/kdeploy/upgrade"
	"github.com/flexiant/kdeploy/utils"
)

func cmdNotFound(c *cli.Context, command string) {
	glog.Fatalf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
}

func prepareFlags(c *cli.Context) error {
	v :=  c.Int("verbosity")
  flag.Parse()
  flag.Lookup("logtostderr").Value.Set("true")
  flag.Lookup("v").Value.Set(string(v))

	// Initialize cached config
	utils.InitializeConfig(c)

	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "kdeploy"
	app.Author = "Concerto Contributors"
	app.Email = "https://github.com/flexiant/kdeploy"

	app.Usage = "Deploys Kubeware in kubernetes clusters"
	app.Version = utils.VERSION
	app.CommandNotFound = cmdNotFound
	app.Before = prepareFlags

	config := utils.PreReadConfig()
	// TODO we shouldn't be pre-reading, since it fills default values for flags.
	// If someone uses a config file as cmd argument, some of the values could be
	// overwritten

	app.Flags = []cli.Flag{
		cli.IntFlag{
      EnvVar: "KDEPLOY_VERBOSITY",
			Name:  "verbosity, V",
      Value: 1,
			Usage: "Verbosity level",
		},
		cli.BoolFlag{
			Name:  "insecure",
			Usage: "Do not validate server certificates",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_CA_CERT",
			Name:   "ca-cert",
			Value:  config.Connection.CACert,
			Usage:  "CA to verify remote connections",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_CLIENT_CERT",
			Name:   "client-cert",
			Value:  config.Connection.Cert,
			Usage:  "Client cert to use for Kubernetes",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_CLIENT_KEY",
			Name:   "client-key",
			Value:  config.Connection.Key,
			Usage:  "Private key used in client Kubernetes auth",
		},
		cli.StringFlag{
			EnvVar: "KUBERNETES_ENDPOINT",
			Name:   "kubernetes-endpoint",
			Value:  config.Connection.APIEndpoint,
			Usage:  "Kubernetes Endpoint",
		},
		cli.StringFlag{
			EnvVar: "KUBECONFIG",
			Name:   "kubeconfig",
			Value:  config.Path,
			Usage:  "Kubeconfig client file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "deploy",
			Usage:  "Deploys a Kubeware",
			Before: deploy.PrepareFlags,
			Action: deploy.CmdDeploy,
			Flags:  deploy.Flags(),
		},
		{
			Name:   "delete",
			Usage:  "Deletes a Kubeware",
			Before: delete.PrepareFlags,
			Action: delete.CmdDelete,
			Flags:  delete.Flags(),
		},
		{
			Name:   "list",
			Usage:  "List's Kubewares deployed",
			Before: list.PrepareFlags,
			Action: list.CmdList,
			Flags:  list.Flags(),
		},
		{
			Name:   "show",
			Usage:  "Shows how a Kubeware would be created once resolved with the indicated attributes",
			Before: show.PrepareFlags,
			Action: show.CmdShow,
			Flags:  show.Flags(),
		},
		{
			Name:   "upgrade",
			Usage:  "Upgrades a Kubeware to a new version",
			Before: upgrade.PrepareFlags,
			Action: upgrade.CmdUpgrade,
			Flags:  upgrade.Flags(),
		},
	}

	app.Run(os.Args)
}
