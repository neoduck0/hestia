package backend

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func Init() error {
	err := os.Mkdir(hestiaDirName, defaultDirPerm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	file, err := os.OpenFile(filepath.Join(hestiaDirName, mappingsFileName), os.O_RDONLY|os.O_CREATE, defaultFilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Infof("initialized %s project: %v", projectName, filepath.Dir(file.Name()))
	return nil
}
