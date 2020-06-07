package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	webfriend "github.com/ghetzel/go-webfriend"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/server"
)

func main() {
	app := cli.NewApp()
	app.Name = `webfriend`
	app.Usage = webfriend.Slogan
	app.Version = webfriend.Version
	app.EnableBashCompletion = true
	app.ArgsUsage = `[FILENAME]`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `info`,
			EnvVar: `LOGLEVEL`,
		},
		cli.BoolFlag{
			Name:   `debug, D`,
			Usage:  `Whether to open the browser in a non-headless mode for debugging purposes.`,
			EnvVar: `WEBFRIEND_DEBUG`,
		},
		cli.BoolFlag{
			Name:  `interactive, I`,
			Usage: `Start Webfriend in an interactive Friendscript shell.`,
		},
		cli.BoolFlag{
			Name:   `server, S`,
			Usage:  `Whether to run the Webfriend Debugging Server`,
			EnvVar: `WEBFRIEND_SERVER`,
		},
		cli.StringFlag{
			Name:   `address, a`,
			Usage:  `If running the Webfriend Debugging Server, this specifies the [address]:port to listen on.`,
			Value:  `:19222`,
			EnvVar: `WEBFRIEND_SERVER_ADDR`,
		},
		cli.BoolFlag{
			Name:  `print-vars, P`,
			Usage: `Print the final state of all variables upon script completion.`,
		},
		cli.StringSliceFlag{
			Name:  `var, V`,
			Usage: `Set one or more variables ([deeply.nested.]key=value) before executing the script.`,
		},
		cli.IntFlag{
			Name:   `remote-debugging-port, R`,
			Usage:  `Explicitly provide the port number for the DevTools protocol.`,
			EnvVar: `WEBFRIEND_REMOTE_DEBUG_PORT`,
			Value:  browser.DefaultDebuggingPort,
		},
		cli.StringFlag{
			Name:   `remote-debugging-address, r`,
			Usage:  `If given, Webfriend will connect to an already-running DevTools instance instead of starting its own browser.`,
			EnvVar: `WEBFRIEND_REMOTE_DEBUG_ADDR`,
		},
		cli.StringFlag{
			Name:  `execute, e`,
			Usage: `Execute the given argument as a Friendscript in the connected session, then exit.`,
		},
		cli.StringFlag{
			Name:   `container, C`,
			Usage:  `If provided, Webfriend will launch and monitor a Chrome session inside of this Docker container.`,
			EnvVar: `WEBFRIEND_CONTAINER`,
		},
		cli.StringFlag{
			Name:   `container-engine, X`,
			Usage:  `Specifies the container runtime to utilize: docker, kubernetes`,
			Value:  browser.DefaultContainerRuntime,
			EnvVar: `WEBFRIEND_CONTAINER_ENGINE`,
		},
		cli.StringFlag{
			Name:   `container-memory`,
			Usage:  `Specify the amount of memory allocated to the container.`,
			Value:  browser.DefaultContainerMemory,
			EnvVar: `WEBFRIEND_CONTAINER_MEMORY`,
		},
		cli.StringFlag{
			Name:   `container-shm-size`,
			Usage:  `Specify the amount of shared memory allocated to the container.`,
			Value:  browser.DefaultContainerSharedMemory,
			EnvVar: `WEBFRIEND_CONTAINER_SHMSIZE`,
		},
		cli.StringFlag{
			Name:   `container-name`,
			Usage:  `Explicitly provide the container name.`,
			EnvVar: `WEBFRIEND_CONTAINER_NAME`,
		},
		cli.StringFlag{
			Name:   `container-hostname`,
			Usage:  `Explicitly set the hostname that will be visible inside the container.`,
			EnvVar: `WEBFRIEND_CONTAINER_HOSTNAME`,
		},
		cli.StringSliceFlag{
			Name:  `container-volume`,
			Usage: `Provide additional volumes to expose to the container; specified as: --container-volume=/outer/path:/inner/path[:ro]`,
		},
		cli.StringSliceFlag{
			Name:  `container-port`,
			Usage: `Provide additional ports to expose from the container; specified as: --container-port=OUTERPORT:INNERPORT`,
		},
	}

	var chrome *browser.Browser

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))
		return nil
	}

	app.Action = func(c *cli.Context) {
		defer browser.StopAllActiveBrowsers()

		log.Debugf("Starting %s %s", c.App.Name, c.App.Version)
		chrome = browser.NewBrowser()
		chrome.Headless = !c.Bool(`debug`)
		chrome.HideScrollbars = true
		chrome.RemoteDebuggingPort = c.Int(`remote-debugging-port`)
		chrome.RemoteAddress = c.String(`remote-debugging-address`)

		if i := c.String(`container`); i != `` {
			switch rt := c.String(`container-engine`); rt {
			case `docker`:
				chrome.Container = browser.NewDockerContainer(``)
			case `kubernetes`:
				chrome.Container = browser.NewKubernetesContainer()
			default:
				log.Fatalf("invalid container engine %q", rt)
				return
			}

			if cfg := chrome.Container.Config(); cfg != nil {
				cfg.Name = c.String(`container-name`)
				cfg.Hostname = c.String(`container-hostname`)
				cfg.ImageName = i
				cfg.Volumes = c.StringSlice(`container-volume`)
				cfg.Memory = c.String(`container-memory`)
				cfg.SharedMemory = c.String(`container-shm-size`)
				cfg.Ports = c.StringSlice(`container-port`)
			}
		}

		if err := chrome.Launch(); err == nil {
			// evaluate Friendscript / run the REPL
			var script = webfriend.NewEnvironment(chrome)
			var wferr error

			// pre-populate initial variables
			for _, pair := range c.StringSlice(`var`) {
				k, v := stringutil.SplitPair(pair, `=`)
				script.Set(k, typeutil.Auto(v))
			}

			for {
				if c.Bool(`server`) {
					wferr = server.NewServer(script).ListenAndServe(c.String(`address`))
				} else if c.Bool(`interactive`) {
					if scope, err := script.REPL(); err == nil {
						fmt.Println(scope)
					} else {
						wferr = fmt.Errorf("runtime error: %v", err)
						break
					}
				} else {
					var input io.Reader

					if e := c.String(`execute`); e != `` {
						input = bytes.NewBufferString(e)
					} else if c.NArg() > 0 {
						var filename = c.Args().First()

						switch filename {
						case `-`:
							input = os.Stdin
						default:
							if file, err := os.Open(c.Args().First()); err == nil {
								log.Debugf("Friendscript being read from file %s", file.Name())
								input = file
							} else {
								wferr = fmt.Errorf("file error: %v", err)
								break
							}
						}
					} else {
						break
					}

					if scope, err := script.EvaluateReader(input); err == nil {
						if c.Bool(`print-vars`) {
							fmt.Println(scope)
						}
					} else {
						wferr = fmt.Errorf("runtime error: %v", err)
					}
				}

				break
			}

			if wferr != nil {
				log.Critical(wferr)
			}
		} else {
			log.Criticalf("could not launch browser: %v", err)
		}

		browser.StopAllActiveBrowsers()
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
