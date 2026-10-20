package backend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neoduck0/hestia/internal/fsutils"
)

// GroupState describes a group's linking status.
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

// GroupStatus holds a group's name and linked/total mapping counts.
type GroupStatus struct {
	Name   string
	Linked int
	Total  int
}

// State returns the group's linking status.
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

// Status returns each group's linking status without changing files.
func (p *Project) Status(s Settings) ([]GroupStatus, error) {
	if err := p.readMappingsFile(s); err != nil {
		return nil, err
	}

	statuses := make([]GroupStatus, 0, len(p.groups))
	for _, g := range p.groups {
		if err := g.inspectState(p); err != nil {
			return nil, err
		}
		statuses = append(statuses, GroupStatus{
			Name:   g.Name,
			Linked: g.linkedMappings,
			Total:  g.totalMappings,
		})
	}
	return statuses, nil
}

func (g *group) inspectState(p *Project) error {
	g.linkedMappings, g.totalMappings = -1, -1

	linkedMappings := 0
	for _, m := range g.Mappings {
		linked, err := m.isLinked(p)
		if err != nil {
			return fmt.Errorf("group %q, mapping %q -> %q: %w", g.Name, m.Src, m.Dst, err)
		}
		if linked {
			linkedMappings++
		}
	}
	g.linkedMappings, g.totalMappings = linkedMappings, len(g.Mappings)
	return nil
}

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
