package main

import (
	"os"
	"os/signal"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-webfriend"
	"github.com/ghetzel/go-webfriend/browser"
	// "github.com/ghetzel/go-webfriend"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger(`webfriend`)

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
	}

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))

		if level, err := logging.LogLevel(c.String(`log-level`)); err == nil {
			logging.SetLevel(level, ``)
		} else {
			return err
		}

		return nil
	}

	app.Action = func(c *cli.Context) {
		log.Infof("Starting %s %s", c.App.Name, c.App.Version)

		if browser, err := browser.Start(); err == nil {
			defer handleSignals(func() {
				browser.Stop()
			})

			script := webfriend.NewEnvironment(browser)

			if file, err := os.Open(c.Args().First()); err == nil {
				if scope, err := script.EvaluateReader(file); err == nil {
					log.Debugf("Final scope: %v", scope)
					log.Infof("Done")
					browser.Stop()
					os.Exit(0)
				} else {
					log.Fatalf("runtime error: %v", err)
				}
			} else {
				log.Fatalf("file error: %v", err)
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
