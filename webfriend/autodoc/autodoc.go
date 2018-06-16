package main

import (
	"encoding/gob"
	"os"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
	webfriend "github.com/ghetzel/go-webfriend"
)

func main() {
	log.SetLevelString(sliceutil.OrString(os.Getenv(`LOGLEVEL`), `info`))

	filename := `documentation.gob`

	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	if out, err := os.Create(filename); err == nil {
		docs := webfriend.NewEnvironment(nil).Documentation()

		if err := gob.NewEncoder(out).Encode(docs); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}
