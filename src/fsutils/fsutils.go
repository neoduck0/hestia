package fsutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func FindDirFiles(dir string) ([]string, error) {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dir)
	}

	files := []string{}

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			relativePath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			files = append(files, relativePath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func CopyFile(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return copySymlink(src, dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	tempDir, err := os.MkdirTemp(filepath.Dir(dst), ".hestia-*.tmp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, "file")

	dstFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}
	if err := dstFile.Chmod(srcInfo.Mode().Perm()); err != nil {
		dstFile.Close()
		return err
	}

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return err
	}

	if err = dstFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, dst)
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}

	return SymlinkFile(target, dst)
}

func SymlinkFile(target, dst string) error {
	tempDir, err := os.MkdirTemp(filepath.Dir(dst), ".hestia-*.tmp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, "link")
	if err := os.Symlink(target, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, dst)
}

func SetSymlinkTarget(src, target string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if srcInfo.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("path is not a symlink: %s", src)
	}

	tempDir, err := os.MkdirTemp(filepath.Dir(src), ".hestia-*.tmp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	tempPath := filepath.Join(tempDir, "link")
	if err := os.Symlink(target, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, src)
}

func CollapsePath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}

	home = filepath.Clean(home)
	if p == home {
		return "~"
	}

	if strings.HasPrefix(p, home+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(p, home)
	}

	return p
}

func DecollapsePath(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if p == "~" {
		return home, nil
	}

	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:]), nil
	}

	return p, nil
}

func ExpandPath(p, root string) (string, error) {
	p, err := DecollapsePath(p)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}

	if root == "" {
		root, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Clean(filepath.Join(root, p)), nil
}
