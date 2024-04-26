package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

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
		cli.DurationFlag{
			Name:  `start-wait-time, W`,
			Usage: `The amount of time that Webfriend should wait for the browser to startup before assuming it never will and killing it.`,
			Value: browser.DefaultStartWait,
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
		cli.DurationFlag{
			Name:  `retrieve-timeout`,
			Usage: `Specifies the timeout for retrieving runnable scripts from remote sources (e.g.: HTTP)`,
			Value: 30 * time.Second,
		},
	}

	var chrome *browser.Browser

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))
		return nil
	}

	app.Action = func(c *cli.Context) {
		go handleSignals(browser.StopAllActiveBrowsers)
		defer browser.StopAllActiveBrowsers()

		log.Infof("Starting %s %s...", c.App.Name, c.App.Version)
		chrome = browser.NewBrowser()
		chrome.Headless = !c.Bool(`debug`)
		chrome.HideScrollbars = true
		chrome.RemoteDebuggingPort = c.Int(`remote-debugging-port`)
		chrome.RemoteAddress = c.String(`remote-debugging-address`)
		chrome.StartWait = c.Duration(`start-wait-time`)

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
						var scriptpath = c.Args().First()
						var scheme, _ = stringutil.SplitPair(scriptpath, `:`)
						scheme = strings.ToLower(scheme)

						switch scheme {
						case `-`:
							input = os.Stdin
						case `http`, `https`:
							http.DefaultClient.Timeout = c.Duration(`retrieve-timeout`)

							if res, err := http.DefaultClient.Get(scriptpath); err == nil {
								if res.StatusCode < 300 {
									defer res.Body.Close()
									input = res.Body
								} else {
									wferr = fmt.Errorf("remote error: request failed with HTTP %s", res.Status)
									break
								}
							} else {
								wferr = fmt.Errorf("remote error: %v", err)
								break
							}
						default:
							if file, err := os.Open(scriptpath); err == nil {
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
