package main

import (
	"os"

	"github.com/campadrenalin/contentgremlin/app"
)

func main() {
	app.Main(nil, true, os.Stderr)
}
