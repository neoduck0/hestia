// Package fsutils provides file, symlink, and path helpers.
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

	"golang.org/x/sys/unix"
)

// FindDirFiles returns relative paths of non-directory entries under dir.
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

// CopyFile atomically replaces dst with a copy of src, preserving permissions
// and copying symlinks as links. The destination's parent directory must exist.
func CopyFile(src, dst string) error {
	return copyFile(src, dst, true)
}

// CopyFileNoReplace copies src to dst, returning fs.ErrExist if dst exists.
func CopyFileNoReplace(src, dst string) error {
	return copyFile(src, dst, false)
}

func copyFile(src, dst string, replace bool) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}

		return symlinkFile(target, dst, replace)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	tempDir, err := os.MkdirTemp(filepath.Dir(dst), ".hestia-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempPath := filepath.Join(tempDir, "file")

	dstFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}
	if err := dstFile.Chmod(srcInfo.Mode().Perm()); err != nil {
		_ = dstFile.Close()
		return err
	}

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		return err
	}

	if err = dstFile.Close(); err != nil {
		return err
	}

	return placeFile(tempPath, dst, replace)
}

// SymlinkFile atomically creates or replaces dst with a symlink to target.
// The destination's parent directory must exist.
func SymlinkFile(target, dst string) error {
	return symlinkFile(target, dst, true)
}

// SymlinkFileNoReplace creates a symlink, returning fs.ErrExist if dst exists.
func SymlinkFileNoReplace(target, dst string) error {
	return symlinkFile(target, dst, false)
}

func symlinkFile(target, dst string, replace bool) error {
	tempDir, err := os.MkdirTemp(filepath.Dir(dst), ".hestia-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempPath := filepath.Join(tempDir, "link")
	if err := os.Symlink(target, tempPath); err != nil {
		return err
	}

	return placeFile(tempPath, dst, replace)
}

// Without replacement, do not fall back to os.Rename: it could overwrite a
// destination created concurrently.
func placeFile(src, dst string, replace bool) error {
	if replace {
		return os.Rename(src, dst)
	}
	if err := unix.Renameat2(unix.AT_FDCWD, src, unix.AT_FDCWD, dst, unix.RENAME_NOREPLACE); err != nil {
		return &os.LinkError{Op: "renameat2", Old: src, New: dst, Err: err}
	}
	return nil
}

// SetSymlinkTarget atomically repoints the existing symlink src to target.
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
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	tempPath := filepath.Join(tempDir, "link")
	if err := os.Symlink(target, tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, src)
}

// IsSymlinkTo reports whether path is a symlink to target.
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

// IsCopyOf reports whether path matches a copy of src, including permissions.
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

func sameContents(a, b string) (bool, error) {
	fileA, err := os.Open(a)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = fileA.Close()
	}()

	fileB, err := os.Open(b)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = fileB.Close()
	}()

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

// CollapsePath replaces a leading home directory with "~".
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

// DecollapsePath expands a leading "~" or "~/" to the home directory.
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

// ExpandPath expands home-relative paths and resolves relative paths against
// an absolute root, or the working directory if root is empty.
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
