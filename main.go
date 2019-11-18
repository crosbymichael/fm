package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "fm"
	app.Version = "1"
	app.Description = "cli file manager"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in the logs",
		},
	}
	app.Before = func(clix *cli.Context) error {
		if clix.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	app.Commands = []cli.Command{
		bulkMoveCommand,
		dupsCommand,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
