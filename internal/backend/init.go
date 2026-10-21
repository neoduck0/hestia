package backend

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/charmbracelet/log"
)

// Init initializes a project in the working directory, preserving existing files.
func Init() error {
	file, err := os.OpenFile(mappingsFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultFilePerm)
	if errors.Is(err, fs.ErrExist) {
		info, err := os.Stat(mappingsFileName)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s must be a regular JSON file", mappingsFileName)
		}
	} else if err != nil {
		return err
	} else {
		defer func() {
			_ = file.Close()
		}()
		if _, err := file.WriteString("[]\n"); err != nil {
			return err
		}
	}

	log.Infof("initialized %s project: %v", projectName, mappingsFileName)
	return nil
}
