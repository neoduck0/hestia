package backend

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

// Init initializes a project in the working directory, preserving existing files.
func Init() error {
	err := os.Mkdir(hestiaDirName, defaultDirPerm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	file, err := os.OpenFile(filepath.Join(hestiaDirName, mappingsFileName), os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultFilePerm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	if err == nil {
		defer func() {
			_ = file.Close()
		}()
		if _, err := file.WriteString("[]\n"); err != nil {
			return err
		}
	}

	log.Infof("initialized %s project: %v", projectName, hestiaDirName)
	return nil
}
