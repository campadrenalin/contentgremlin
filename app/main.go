package app

import (
	"io"
	"log"

	"github.com/docopt/docopt-go"
)

var version = "contentgremlin 0.0.1"
var usage = `contentgremlin - Service for Federated Media Hosting

Usage:
    contentgremlin init [ <dirpath> ]
    contentgremlin -h | --help
    contentgremlin --version
`

func Main(argv []string, exit bool, output io.Writer) {
	logger := log.New(output, "contentgremlin: ", 0)
	args, err := docopt.Parse(usage, argv, true, version, false, exit)
	if err != nil {
		logger.Fatal(err)
	}
	if args["init"].(bool) {
		dirpath, ok := args["<dirpath>"].(string)
		if !ok {
			dirpath = "."
		}
		if err = InitDirectory(dirpath, logger); err != nil {
			logger.Fatal(err)
		}
	}
	logger.Printf("Arguments: %v\n", args)
}
