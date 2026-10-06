package backend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neoduck0/hestia/src/fsutils"
)

// GroupState describes a group's link state without prescribing display text.
type GroupState uint8

const (
	// GroupNotLinked means no mappings in a non-empty group are fully linked.
	GroupNotLinked GroupState = iota
	// GroupPartiallyLinked means some, but not all, mappings are fully linked.
	GroupPartiallyLinked
	// GroupLinked means every mapping in a non-empty group is fully linked.
	GroupLinked
	// GroupEmpty means the group has no mappings.
	GroupEmpty
)

// GroupStatus describes how many mappings in a group are currently linked.
// A matching symlink or copy counts as linked, regardless of the operation
// used to create it. A directory mapping counts once, and only when all its
// files are linked. Empty directories do not count as linked.
type GroupStatus struct {
	Name   string
	Linked int
	Total  int
}

// State returns the group state based only on the number of fully linked
// mappings. Groups with no mappings are empty. Callers choose how to display it.
func (s GroupStatus) State() GroupState {
	if s.Total == 0 {
		return GroupEmpty
	}

	if s.Linked == 0 {
		return GroupNotLinked
	}

	if s.Total > 0 && s.Linked == s.Total {
		return GroupLinked
	}

	return GroupPartiallyLinked
}

// Status returns the status of every group in mappings-file order without
// changing files. It uses the same mappings validation as other operations,
// including requiring sources to exist.
func (p *Project) Status(s Settings) ([]GroupStatus, error) {
	if err := p.readMappingsFile(s); err != nil {
		return nil, err
	}

	statuses := make([]GroupStatus, 0, len(p.groups))
	for i := range p.groups {
		g := &p.groups[i]
		if err := g.inspectState(p); err != nil {
			return nil, err
		}
		statuses = append(statuses, GroupStatus{
			Name:   g.name,
			Linked: g.linkedMappings,
			Total:  g.totalMappings,
		})
	}
	return statuses, nil
}

// inspectState checks the group's mappings and stores only group-level counts.
// Failed inspection leaves all counts at -1 rather than a partial snapshot.
func (g *group) inspectState(p *Project) error {
	g.linkedMappings, g.totalMappings = -1, -1

	linkedMappings := 0
	for _, m := range g.mappings {
		linked, err := m.isLinked(p)
		if err != nil {
			return fmt.Errorf("group %q, mapping %q -> %q: %w", g.name, m.src, m.dst, err)
		}
		if linked {
			linkedMappings++
		}
	}
	g.linkedMappings, g.totalMappings = linkedMappings, len(g.mappings)
	return nil
}

// isLinked reports whether the mapping is fully linked, without storing state
// on it. It checks the file paths used by link, accepting either a symlink or
// a matching copy. Empty directories and directories with any unlinked files
// are not linked. Checking stops at the first unlinked file.
func (m *mapping) isLinked(p *Project) (bool, error) {
	src, err := m.absSrc(p)
	if err != nil {
		return false, err
	}

	dst, err := m.absDst(p)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	files := []string{""}
	if info.IsDir() {
		files, err = fsutils.FindDirFiles(src)
		if err != nil {
			return false, err
		}
	}

	for _, file := range files {
		fileSrc, fileDst := filepath.Join(src, file), filepath.Join(dst, file)
		linked, err := fsutils.IsSymlinkTo(fileDst, fileSrc)
		if err != nil {
			return false, err
		}
		if !linked {
			linked, err = fsutils.IsCopyOf(fileDst, fileSrc)
			if err != nil {
				return false, err
			}
		}
		if !linked {
			return false, nil
		}
	}
	return len(files) > 0, nil
}
