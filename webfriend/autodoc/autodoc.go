package main

import (
	"encoding/gob"
	"os"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-webfriend"
)

func main() {
	app := cli.NewApp()
	app.Name = `webfriend-autodoc`
	app.EnableBashCompletion = false

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `debug`,
			EnvVar: `LOGLEVEL`,
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))
		return nil
	}

	app.Action = func(c *cli.Context) {
		if out, err := os.Create(`autodoc.gob`); err == nil {
			docs := webfriend.NewEnvironment(nil).Documentation()

			if err := gob.NewEncoder(out).Encode(docs); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	app.Run(os.Args)
}
