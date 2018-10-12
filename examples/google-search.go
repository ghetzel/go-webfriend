package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ghetzel/go-webfriend"
	browser "github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/commands/core"
)

const (
	googleUrl string = "https://google.com"
)

// PROTIP: Export the environment variable LOGLEVEL=debug before executing your program to see
//         every single little thing going on behind the scenes while Webfriend is executing.
//         If it's too much, "info", "warning", and "error" are also valid values.

func main() {
	var searchTerm string

	if len(os.Args) > 1 {
		searchTerm = strings.Join(os.Args[1:], ` `)
	} else {
		log.Fatalf("Must specify a search term")
	}

	chrome := browser.NewBrowser()

	// uncomment this to actually see the browser doing stuff
	// chrome.Headless = false

	if err := chrome.Launch(); err == nil {
		environment := webfriend.NewEnvironment(chrome)

		// call the high-level "go" command, which takes care of a BUNCH of
		// stuff for you; namely it blocks until the page loads or the
		// timeout is reached.
		if _, err := environment.Core.Go(googleUrl, &core.GoArgs{
			Timeout: 10 * time.Second,
		}); err != nil {
			log.Fatalf("failed to load %v: %v", googleUrl, err)
		}

		// type in the search term into the #lst-ib field (what the search input is called)
		// on google.com
		//
		// The "Enter" parameter will send an Enter/Return keypress (keycode 13, 0x0D).
		// This will perform the search.
		if _, err := environment.Core.Field(`#lst-ib`, &core.FieldArgs{
			Value: searchTerm,
			Enter: true,
		}); err != nil {
			log.Fatalf("failed to perform search: %v", err)
		}

		// find all search result links, spend up to 5 seconds looking for them.
		if resultLinks, err := environment.Core.Select(`h3.r a`, &core.SelectArgs{
			CanBeEmpty: true,
		}); err == nil {
			if len(resultLinks) > 0 {
				// loop through all search result links and print their name and URL
				for i, a := range resultLinks {
					fmt.Printf("[%02d] %s\n", i, a.Text())
					fmt.Printf("       %s\n", a.Attributes()[`href`])
				}

			} else {
				fmt.Printf("No results found for %q\n", searchTerm)
			}

		} else {
			log.Fatal(err)
		}
	}
}
