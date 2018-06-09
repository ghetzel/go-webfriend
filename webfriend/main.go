package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/log"
	webfriend "github.com/ghetzel/go-webfriend"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/server"
)

func main() {
	app := cli.NewApp()
	app.Name = `webfriend`
	app.Usage = webfriend.Slogan
	app.Version = webfriend.Version
	app.EnableBashCompletion = false

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `info`,
			EnvVar: `LOGLEVEL`,
		},
		cli.BoolFlag{
			Name:  `debug, D`,
			Usage: `Whether to open the browser in a non-headless mode for debugging purposes.`,
		},
		cli.BoolFlag{
			Name:  `interactive, I`,
			Usage: `Start Webfriend in an interactive Friendscript shell.`,
		},
		cli.BoolFlag{
			Name:  `server, S`,
			Usage: `Whether to run the Webfriend Debugging Server`,
		},
		cli.StringFlag{
			Name:  `address, a`,
			Usage: `If running the Webfriend Debugging Server, this specifies the [address]:port to listen on.`,
			Value: `:19222`,
		},
	}

	var chrome *browser.Browser

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))
		return nil
	}

	app.After = func(c *cli.Context) error {
		if chrome != nil {
			return chrome.Stop()
		} else {
			return nil
		}
	}

	app.Action = func(c *cli.Context) {
		log.Infof("Starting %s %s", c.App.Name, c.App.Version)
		chrome = browser.NewBrowser()
		chrome.Headless = !c.Bool(`debug`)

		if err := chrome.Launch(); err == nil {
			exiterr := make(chan error)
			defer chrome.Stop()

			defer func() {
				if r := recover(); r != nil {
					log.Criticalf("Emergency Stop: %v", r)
					chrome.Stop()
					os.Exit(127)
				}
			}()

			// if Chrome exits before we do, cleanup and quit
			go func() {
				err := chrome.Wait()
				exiterr <- err
			}()

			// evaluate Friendscript / run the REPL
			script := webfriend.NewEnvironment(chrome)

			go func() {
				if c.Bool(`server`) {
					exiterr <- server.NewServer(script).ListenAndServe(c.String(`address`))
					return
				} else if c.Bool(`interactive`) {
					if scope, err := script.REPL(); err == nil {
						fmt.Println(scope)
						exiterr <- nil
					} else {
						exiterr <- fmt.Errorf("runtime error: %v", err)
					}
				} else {
					var input io.Reader

					if c.NArg() > 0 {
						filename := c.Args().First()

						switch filename {
						case `-`:
							input = os.Stdin
						default:
							if file, err := os.Open(c.Args().First()); err == nil {
								log.Debug("Friendscript being read from file %s", file.Name())
								input = file
							} else {
								exiterr <- fmt.Errorf("file error: %v", err)
								return
							}
						}
					} else {
						exiterr <- fmt.Errorf("Must specify a Friendscript filename to execute.")
						return
					}

					if scope, err := script.EvaluateReader(input); err == nil {
						fmt.Println(scope)
						exiterr <- nil
					} else {
						exiterr <- fmt.Errorf("runtime error: %v", err)
					}
				}
			}()

			select {
			case err := <-exiterr:
				if err != nil {
					log.Error(err)
				}
			}
		} else {
			log.Criticalf("could not launch browser: %v", err)
			return
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
