package app

import (
	"log"
	"os"
	"path"

	"github.com/campadrenalin/contentgremlin/cgdb"
)

func InitDirectory(dirpath string, logger *log.Logger) error {
	logger.Printf("Attempting to initialize in %s...\n", dirpath)

	// Set up top-level directory
	mode := os.ModeDir + 0740
	err := os.MkdirAll(dirpath, mode)
	if err != nil {
		return err
	}

	// Set up database
	db, err := cgdb.Open(dirpath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Init(); err != nil {
		return err
	}

	// Set up "contents" subdir
	err = os.Mkdir(path.Join(dirpath, "contents"), mode)
	if err == nil {
		logger.Println("Successfully initialized CG dir in " + dirpath)
	}
	return err
}
