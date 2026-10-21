// Package backend manages Hestia projects, mappings, and file deployment.
package backend

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/internal/fsutils"
)

const (
	projectName      = "Hestia"
	mappingsFileName = ".hestia"

	defaultDirPerm  os.FileMode = 0o777
	defaultFilePerm os.FileMode = 0o666
)

// Project represents a Hestia project.
type Project struct {
	root string

	groups []*group
}

func (p *Project) findGroupIndex(groupName string) int {
	for i := range p.groups {
		if p.groups[i].Name == groupName {
			return i
		}
	}
	return -1
}

// NewProject returns an empty project.
func NewProject() Project {
	return Project{}
}

// Settings controls how project operations behave.
type Settings struct {
	// DryRun validates mappings without linking files.
	DryRun bool

	// Force allows replacing existing destinations.
	Force bool

	// DefaultOp is the operation used when no override is specified.
	DefaultOp Op
	// ForceOp overrides the operation for every mapping.
	ForceOp Op

	// NoPortable disables collapsing home-directory paths to "~".
	NoPortable bool
}

// NewSettings returns the default project settings.
func NewSettings() Settings {
	return Settings{
		DefaultOp: DefaultOp,
	}
}

// Op is a file deployment operation.
type Op string

const (
	// DefaultOp is the operation used when no override is specified.
	DefaultOp Op = OpSymlink

	// OpSymlink creates symlinks to source files.
	OpSymlink Op = "symlink"
	// OpCopy copies source files.
	OpCopy Op = "copy"
)

// SetOp stores a valid operation in v or returns an error.
func SetOp(op Op, v *Op) error {
	switch op {
	case OpSymlink, OpCopy:
		*v = op
		return nil
	}
	return fmt.Errorf("invalid operation: %s", op)
}

// SetDryRun sets s.DryRun to b.
func (s *Settings) SetDryRun(b bool) {
	s.DryRun = b
}

func (p *Project) findProjectRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	dir := wd
	for {
		candidate := filepath.Join(dir, mappingsFileName)
		info, err := os.Stat(candidate)
		if err == nil {
			if info.Mode().IsRegular() {
				p.root = dir
				log.Debugf("found hestia file: %v", candidate)
				return nil
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return errors.New("hestia project file (.hestia) not found")
}

type group struct {
	Name     string     `json:"name"`
	Mappings []*mapping `json:"mappings"`

	linkedMappings int
	totalMappings  int
}

func (g *group) link(p *Project, s Settings) error {
	log.Debugf("linking group: %v", g.Name)
	for _, m := range g.Mappings {
		err := m.link(p, s)
		if err != nil {
			return err
		}
	}
	log.Infof("linked group: %v", g.Name)
	return nil
}

func (g *group) addMapping(p *Project, m *mapping) error {
	exists, err := dstExists(p, m.Dst)
	if err != nil {
		return err
	}

	if exists {
		log.Warnf("another mapping has the same destination: %s", m.Dst)
	}

	g.Mappings = append(g.Mappings, m)

	return nil
}

func newGroup(name string) *group {
	return &group{
		Name:           name,
		Mappings:       []*mapping{},
		linkedMappings: -1,
		totalMappings:  -1,
	}
}

type mapping struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
	op  Op
}

func (m *mapping) absSrc(p *Project) (string, error) {
	return fsutils.ExpandPath(m.Src, p.root)
}

func (m *mapping) absDst(p *Project) (string, error) {
	return fsutils.ExpandPath(m.Dst, p.root)
}

func (m *mapping) link(p *Project, s Settings) error {
	absSrc, err := m.absSrc(p)
	if err != nil {
		return err
	}

	absDst, err := m.absDst(p)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(absSrc)
	if err != nil {
		return err
	}

	files := []string{}

	if fileInfo.IsDir() {
		files, err = fsutils.FindDirFiles(absSrc)
		if err != nil {
			return err
		}
		log.Debugf("discovered %d files in directory: %v", len(files), absSrc)
	} else {
		files = append(files, "")
	}

	if s.DryRun {
		log.Infof("dry run skipped: %v", m.Dst)
		return nil
	}

	for _, p := range files {
		resolvedSrc := ""
		resolvedDst := ""
		if !fileInfo.IsDir() {
			resolvedSrc = absSrc
			resolvedDst = absDst
		} else {
			resolvedSrc = filepath.Join(absSrc, p)
			resolvedDst = filepath.Join(absDst, p)
		}

		if _, err := os.Lstat(resolvedDst); err == nil {
			log.Infof("file exists in destination: %v", resolvedDst)
			if !s.Force {
				continue
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to check destination existence: %w", err)
		}

		log.Debugf("making directories: %v", filepath.Dir(resolvedDst))
		err = os.MkdirAll(filepath.Dir(resolvedDst), defaultDirPerm)
		if err != nil {
			return err
		}

		switch m.op {
		case OpSymlink:
			var linked bool
			linked, err = fsutils.IsSymlinkTo(resolvedDst, resolvedSrc)
			if err != nil {
				return err
			}
			if linked {
				log.Debugf("already symlinked, skipping: %v", resolvedDst)
				continue
			}

			log.Debugf("symlinking: %v -> %v", resolvedSrc, resolvedDst)
			if s.Force {
				err = fsutils.SymlinkFile(resolvedSrc, resolvedDst)
			} else {
				err = fsutils.SymlinkFileNoReplace(resolvedSrc, resolvedDst)
			}
		case OpCopy:
			var copied bool
			copied, err = fsutils.IsCopyOf(resolvedDst, resolvedSrc)
			if err != nil {
				return err
			}
			if copied {
				log.Debugf("already copied, skipping: %v", resolvedDst)
				continue
			}

			log.Debugf("copying: %v -> %v", resolvedSrc, resolvedDst)
			if s.Force {
				err = fsutils.CopyFile(resolvedSrc, resolvedDst)
			} else {
				err = fsutils.CopyFileNoReplace(resolvedSrc, resolvedDst)
			}
		default:
			return fmt.Errorf("unknown operation: %s", m.op)
		}
		if !s.Force && errors.Is(err, fs.ErrExist) {
			log.Infof("file exists in destination: %v", resolvedDst)
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func newMapping(p *Project, s Settings, src, dst string) (*mapping, error) {
	if err := verifySrc(p, src); err != nil {
		return nil, err
	}

	var op Op
	if s.ForceOp != "" {
		op = s.ForceOp
	} else {
		op = s.DefaultOp
	}

	return &mapping{
		Src: src,
		Dst: dst,
		op:  op,
	}, nil
}

func verifySrc(p *Project, src string) error {
	absSrc, err := fsutils.ExpandPath(src, p.root)
	if err != nil {
		return err
	}

	_, err = os.Stat(absSrc)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("source does not exist: %s", src)
		}
		return err
	}

	return nil
}

func dstExists(p *Project, dst string) (bool, error) {
	absDst, err := fsutils.ExpandPath(dst, p.root)
	if err != nil {
		return false, err
	}

	for _, g := range p.groups {
		for _, m := range g.Mappings {
			d, err := m.absDst(p)
			if err != nil {
				return false, err
			}

			if d == absDst {
				return true, nil
			}
		}
	}

	return false, nil
}
