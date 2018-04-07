package main

import (
	"io"
	"os"
	"os/signal"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-webfriend"
	"github.com/ghetzel/go-webfriend/browser"
)

func main() {
	app := cli.NewApp()
	app.Name = `webfriend`
	app.Usage = `Your friendly friend in web browser automation.`
	app.Version = `1.0.0`
	app.EnableBashCompletion = false

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `debug`,
			EnvVar: `LOGLEVEL`,
		},
		cli.BoolFlag{
			Name:  `debug, D`,
			Usage: `Whether to open the browser in a non-headless mode for debugging purposes.`,
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))
		return nil
	}

	app.Action = func(c *cli.Context) {
		log.Infof("Starting %s %s", c.App.Name, c.App.Version)
		browser := browser.NewBrowser()
		browser.Headless = !c.Bool(`debug`)

		if err := browser.Launch(); err == nil {
			defer handleSignals(func() {
				browser.Stop()
			})

			script := webfriend.NewEnvironment(browser)
			var input io.Reader

			if c.NArg() > 0 {
				if file, err := os.Open(c.Args().First()); err == nil {
					input = file
				} else {
					log.Fatalf("file error: %v", err)
				}
			} else if fileutil.IsTerminal() {
				input = os.Stdin
			} else {
				log.Fatal("Must specify a file to execute or Friendscript via standard input")
			}

			if scope, err := script.EvaluateReader(input); err == nil {
				log.Debugf("Final scope: %v", scope)
				log.Infof("Done")
				browser.Stop()
				os.Exit(0)
			} else {
				log.Fatalf("runtime error: %v", err)
			}
		} else {
			log.Fatalf("could not launch browser: %v", err)
		}
	}

	app.Run(os.Args)
}

func handleSignals(handler func()) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for _ = range signalChan {
		handler()
		break
	}

	os.Exit(0)
}
