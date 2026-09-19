// Package backend implements Hestia projects: reading and writing the
// mappings file, managing groups of mappings, and linking mapped sources to
// their destinations.
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
	// projectName is the user-facing name of the tool.
	projectName = "Hestia"

	// hestiaDirName is the name of the directory that marks a project root.
	hestiaDirName = ".hestia"

	// defaultDirPerm and defaultFilePerm are the modes, before umask, used
	// when creating directories and files.
	defaultDirPerm  os.FileMode = 0o777
	defaultFilePerm os.FileMode = 0o666
)

// Project is a Hestia project rooted at a .hestia directory. Its groups are
// populated from the mappings file by each operation.
type Project struct {
	root string

	groups []group
}

// findGroupIndex returns the index of the group named groupName in p.groups,
// or -1 if there is none.
func (p *Project) findGroupIndex(groupName string) int {
	for i := range p.groups {
		if p.groups[i].name == groupName {
			return i
		}
	}
	return -1
}

// NewProject returns an empty Project. The project root is located when an
// operation is run.
func NewProject() Project {
	return Project{}
}

// Settings controls how project operations behave.
type Settings struct {
	// DryRun reports what would be linked without touching the filesystem.
	DryRun bool

	// DefaultOp is the operation used for mappings when ForceOp is unset.
	DefaultOp Op
	// ForceOp, if non-empty, overrides DefaultOp for every mapping.
	ForceOp Op

	// NoPortable stores paths as given instead of collapsing the home
	// directory to "~".
	NoPortable bool
}

// NewSettings returns Settings with DefaultOp set to the package default.
func NewSettings() Settings {
	return Settings{
		DefaultOp: DefaultOp,
	}
}

// Op is the operation used to place a source file at its destination.
type Op string

const (
	// DefaultOp is the operation used when none is specified.
	DefaultOp Op = OpSymlink

	// OpSymlink places a symlink to the source at the destination.
	OpSymlink Op = "symlink"
	// OpCopy places a copy of the source at the destination.
	OpCopy Op = "copy"
)

// SetOp stores op in v, returning an error if op is not a known Op.
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

// findHestiaDir searches the working directory and its ancestors for a
// .hestia directory and sets p.root to the first one found.
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
				log.Debugf("found hestia directory: %v", p.root)
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

// group is a named set of mappings that are linked together.
type group struct {
	name     string
	mappings []*mapping
}

// link links every mapping in g, stopping at the first error.
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

// addMapping appends m to g after checking that its destination is not
// already mapped anywhere in p.
func (g *group) addMapping(p *Project, m *mapping) error {
	if err := verifyDst(p, m.dst); err != nil {
		return err
	}

	g.mappings = append(g.mappings, m)

	return nil
}

// newGroup returns an empty group with the given name.
func newGroup(name string) group {
	return group{name: name}
}

// mapping pairs a source path with a destination path and the operation
// used to place it. Relative paths are resolved against the directory
// containing the project's .hestia directory.
type mapping struct {
	src string
	dst string
	op  Op
}

// absSrc returns the absolute path of m's source.
func (m *mapping) absSrc(p *Project) (string, error) {
	return fsutils.ExpandPath(m.src, filepath.Dir(p.root))
}

// absDst returns the absolute path of m's destination.
func (m *mapping) absDst(p *Project) (string, error) {
	return fsutils.ExpandPath(m.dst, filepath.Dir(p.root))
}

// link places m's source at its destination using m.op, creating parent
// directories as needed. A directory source is linked file by file. Existing
// symlinks that already point to the source, and existing copies that already
// match it, are left untouched.
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
			err = fsutils.SymlinkFile(resolvedSrc, resolvedDst)
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

// newMapping returns a mapping from src to dst after checking that src
// exists. The operation is s.ForceOp if set, otherwise s.DefaultOp.
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

// verifySrc returns an error if src does not exist.
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

// verifyDst returns an error if dst is already the destination of a mapping
// in p.
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
