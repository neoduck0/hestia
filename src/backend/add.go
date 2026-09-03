package backend

import (
	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/fsutils"
)

func (p *Project) Add(s Settings, groupName, src, dst string) error {
	if err := p.readMappingsFile(s); err != nil {
		return err
	}

	var chosenGroup *group
	if groupIndex := p.findGroupIndex(groupName); groupIndex != -1 {
		chosenGroup = &p.groups[groupIndex]
	} else {
		log.Debugf("creating group: %v", groupName)
		p.groups = append(p.groups, newGroup(groupName))
		chosenGroup = &p.groups[len(p.groups)-1]
	}

	newSrc := src
	newDst := dst
	if !s.NoPortable {
		newSrc = fsutils.CollapsePath(src)
		newDst = fsutils.CollapsePath(dst)
	}

	newMapping, err := newMapping(p, s, newSrc, newDst)
	if err != nil {
		return err
	}

	log.Debugf("adding mapping to %v: %v -> %v", groupName, newSrc, newDst)
	if err := chosenGroup.addMapping(p, newMapping); err != nil {
		return err
	}

	if err := p.writeMappingsFile(); err != nil {
		return err
	}

	return nil
}
