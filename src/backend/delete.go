package backend

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/log"
)

func (p *Project) Delete(s Settings, groupName string) error {
	if err := p.readMappingsFile(s); err != nil {
		return err
	}

	groupIndex := p.findGroupIndex(groupName)
	if groupIndex == -1 {
		return fmt.Errorf("group does not exist: %s", groupName)
	}

	p.groups = slices.Delete(p.groups, groupIndex, groupIndex+1)

	if err := p.writeMappingsFile(); err != nil {
		return err
	}

	log.Infof("group deleted: %v", groupName)

	return nil
}
