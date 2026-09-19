// Package fsutils provides filesystem helpers for atomically placing files
// and symlinks and for resolving home-relative paths.
package fsutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindDirFiles returns the paths, relative to dir, of every non-directory
// entry under dir.
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

// CopyFile copies src to dst, preserving its permissions. If src is a
// symlink, the link itself is copied rather than its target. dst is replaced
// atomically via a temporary file in the same directory.
func CopyFile(src, dst string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}

		return SymlinkFile(target, dst)
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

// SymlinkFile atomically creates or replaces dst with a symlink to target.
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

// SetSymlinkTarget atomically repoints the existing symlink src to target.
// It returns an error if src is not a symlink.
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

// IsSymlinkTo reports whether path is a symlink whose target is exactly
// target. A missing path is not an error.
func IsSymlinkTo(path, target string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	current, err := os.Readlink(path)
	if err != nil {
		return false, err
	}

	return current == target, nil
}

// IsCopyOf reports whether path already matches what CopyFile would produce
// from src: a symlink with the same target if src is a symlink, otherwise a
// regular file with the same permissions and contents. A missing path is not
// an error.
func IsCopyOf(path, src string) (bool, error) {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return false, err
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return false, err
		}

		return IsSymlinkTo(path, target)
	}

	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if !info.Mode().IsRegular() ||
		info.Mode().Perm() != srcInfo.Mode().Perm() ||
		info.Size() != srcInfo.Size() {
		return false, nil
	}

	return sameContents(path, src)
}

// sameContents reports whether the files at a and b have identical contents.
func sameContents(a, b string) (bool, error) {
	fileA, err := os.Open(a)
	if err != nil {
		return false, err
	}
	defer fileA.Close()

	fileB, err := os.Open(b)
	if err != nil {
		return false, err
	}
	defer fileB.Close()

	bufA := make([]byte, 32*1024)
	bufB := make([]byte, 32*1024)
	for {
		nA, errA := io.ReadFull(fileA, bufA)
		nB, errB := io.ReadFull(fileB, bufB)
		if !bytes.Equal(bufA[:nA], bufB[:nB]) {
			return false, nil
		}

		endA := errors.Is(errA, io.EOF) || errors.Is(errA, io.ErrUnexpectedEOF)
		endB := errors.Is(errB, io.EOF) || errors.Is(errB, io.ErrUnexpectedEOF)
		if errA != nil && !endA {
			return false, errA
		}
		if errB != nil && !endB {
			return false, errB
		}
		if endA || endB {
			return endA && endB, nil
		}
	}
}

// CollapsePath replaces a leading home directory in p with "~". It returns p
// unchanged if p is not under the home directory or the home directory is
// unknown.
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

// DecollapsePath expands a leading "~" or "~/" in p to the home directory.
// Other paths, including "~user" forms, are returned unchanged.
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

// ExpandPath returns p as a clean absolute path. A leading "~" is expanded
// and relative paths are joined to root, or to the working directory if root
// is empty.
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
