package backend

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/fsutils"
)

// Add maps src to dst in the group named groupName and saves the mappings
// file. If the group does not exist, it is created when create is true and an
// error is returned otherwise. Unless s.NoPortable is set, paths under the
// home directory are stored with a leading "~".
func (p *Project) Add(s Settings, groupName, src, dst string, create bool) error {
	if err := p.readMappingsFile(s); err != nil {
		return err
	}

	var chosenGroup *group
	if groupIndex := p.findGroupIndex(groupName); groupIndex != -1 {
		chosenGroup = &p.groups[groupIndex]
	} else if create {
		var err error
		if chosenGroup, err = p.createGroup(groupName); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("group does not exist: %s", groupName)
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
