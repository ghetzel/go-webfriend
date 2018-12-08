package main

import (
	"encoding/json"
	"os"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
	webfriend "github.com/ghetzel/go-webfriend"
)

func main() {
	log.SetLevelString(sliceutil.OrString(os.Getenv(`LOGLEVEL`), `info`))

	filename := `documentation.json`

	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	if out, err := os.Create(filename); err == nil {
		docs := webfriend.NewEnvironment(nil).Documentation()
		enc := json.NewEncoder(out)
		enc.SetIndent(``, `  `)

		if err := enc.Encode(docs); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}
