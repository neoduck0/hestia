package backend

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/fsutils"
)

const (
	projectName = "Hestia"

	hestiaDirName = ".hestia"

	defaultDirPerm  os.FileMode = 0o777
	defaultFilePerm os.FileMode = 0o666
)

type Project struct {
	root string

	groups []group
}

func (p *Project) findGroupIndex(groupName string) int {
	for i := range p.groups {
		if p.groups[i].name == groupName {
			return i
		}
	}
	return -1
}

func NewProject() Project {
	return Project{}
}

type Settings struct {
	DryRun bool

	DefaultOp Op
	ForceOp   Op

	NoPortable bool
}

func NewSettings() Settings {
	return Settings{
		DefaultOp: DefaultOp,
	}
}

type Op string

const (
	DefaultOp Op = OpSymlink

	OpSymlink Op = "symlink"
	OpCopy    Op = "copy"
)

func SetOp(op Op, v *Op) error {
	switch op {
	case OpSymlink, OpCopy:
		*v = op
		return nil
	}
	return fmt.Errorf("invalid operation: %s", op)
}

func (s *Settings) SetDryRun(b bool) {
	s.DryRun = b
}

func (p *Project) findHestiaDir() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	dir := wd
	for {
		candidate := filepath.Join(dir, hestiaDirName)
		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				p.root = candidate
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

	return errors.New("hestia directory not found")
}

type group struct {
	name     string
	mappings []*mapping
}

func (g *group) link(p *Project, s Settings) error {
	log.Debugf("linking group: %v", g.name)
	for _, m := range g.mappings {
		err := m.link(p, s)
		if err != nil {
			return err
		}
	}
	log.Infof("linked group: %v", g.name)
	return nil
}

func (g *group) addMapping(p *Project, m *mapping) error {
	if err := verifyDst(p, m.dst); err != nil {
		return err
	}

	g.mappings = append(g.mappings, m)

	return nil
}

func newGroup(name string) group {
	return group{name: name}
}

type mapping struct {
	src string
	dst string
	op  Op
}

func (m *mapping) absSrc(p *Project) (string, error) {
	return fsutils.ExpandPath(m.src, filepath.Dir(p.root))
}

func (m *mapping) absDst(p *Project) (string, error) {
	return fsutils.ExpandPath(m.dst, filepath.Dir(p.root))
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
	} else {
		files = append(files, "")
	}

	if s.DryRun {
		log.Infof("dry run skipped: %v", m.dst)
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

		log.Debugf("making directories: %v", filepath.Dir(resolvedDst))
		err = os.MkdirAll(filepath.Dir(resolvedDst), defaultDirPerm)
		if err != nil {
			return err
		}

		switch m.op {
		case OpSymlink:
			err = fsutils.SymlinkFile(resolvedSrc, resolvedDst)
		case OpCopy:
			err = fsutils.CopyFile(resolvedSrc, resolvedDst)
		default:
			return fmt.Errorf("unknown operation: %s", m.op)
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
		src: src,
		dst: dst,
		op:  op,
	}, nil
}

func verifySrc(p *Project, src string) error {
	absSrc, err := fsutils.ExpandPath(src, filepath.Dir(p.root))
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

func verifyDst(p *Project, dst string) error {
	absDst, err := fsutils.ExpandPath(dst, filepath.Dir(p.root))
	if err != nil {
		return err
	}

	for _, g := range p.groups {
		for _, m := range g.mappings {
			d, err := m.absDst(p)
			if err != nil {
				return err
			}

			if d == absDst {
				return fmt.Errorf("destination is already mapped: %s", dst)
			}
		}
	}

	return nil
}
