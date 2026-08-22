package backend

import (
	"os"
	"path/filepath"
	"strings"
)

func (p *Project) Add(s Settings, groupName, src, dst string) error {
	if err := p.readMappings(s); err != nil {
		return err
	}

	var chosenGroup *group

	for i := range p.groups {
		if p.groups[i].name == groupName {
			chosenGroup = &p.groups[i]
			break
		}
	}

	if chosenGroup == nil {
		p.groups = append(p.groups, newGroup(groupName))
		chosenGroup = &p.groups[len(p.groups)-1]
	}

	newSrc := src
	newDst := dst
	if !s.SkipPortable {
		var err error
		newSrc, err = portablePath(src)
		if err != nil {
			return err
		}
		newDst, err = portablePath(dst)
		if err != nil {
			return err
		}
	}

	newMapping, err := newMapping(p, s, newSrc, newDst)
	if err != nil {
		return err
	}

	if err := chosenGroup.addMapping(p, newMapping); err != nil {
		return err
	}

	if err := p.writeMappings(); err != nil {
		return err
	}

	return nil
}

func portablePath(s string) (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if s == userHome {
		return "~", nil
	}

	prefix := filepath.Clean(userHome) + "/"

	newPath := s
	if strings.HasPrefix(s, prefix) {
		newPath = strings.Replace(s, prefix, "~/", 1)
	}

	return newPath, nil
}
