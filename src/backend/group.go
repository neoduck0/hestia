package backend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
)

// AddGroup creates an empty group named groupName and saves the mappings
// file.
func (p *Project) AddGroup(s Settings, groupName string) error {
	if err := p.readMappingsFile(s); err != nil {
		return err
	}

	if _, err := p.createGroup(groupName); err != nil {
		return err
	}

	if err := p.writeMappingsFile(); err != nil {
		return err
	}

	log.Infof("group added: %v", groupName)

	return nil
}

// RenameGroup renames the group oldName to newName and saves the mappings
// file. It returns an error if oldName does not exist or newName is taken.
func (p *Project) RenameGroup(s Settings, oldName, newName string) error {
	if err := validateGroupName(newName); err != nil {
		return err
	}

	if err := p.readMappingsFile(s); err != nil {
		return err
	}

	groupIndex := p.findGroupIndex(oldName)
	if groupIndex == -1 {
		return fmt.Errorf("group does not exist: %s", oldName)
	}

	if oldName == newName {
		return nil
	}

	if p.findGroupIndex(newName) != -1 {
		return fmt.Errorf("group already exists: %s", newName)
	}

	log.Debugf("renaming group: %v -> %v", oldName, newName)
	p.groups[groupIndex].name = newName

	if err := p.writeMappingsFile(); err != nil {
		return err
	}

	log.Infof("group renamed: %v -> %v", oldName, newName)

	return nil
}

// createGroup validates name, appends a new empty group to p, and returns a
// pointer to it. The pointer is invalidated if p.groups is reallocated.
func (p *Project) createGroup(name string) (*group, error) {
	if err := validateGroupName(name); err != nil {
		return nil, err
	}

	if p.findGroupIndex(name) != -1 {
		return nil, fmt.Errorf("group already exists: %s", name)
	}

	log.Debugf("creating group: %v", name)
	p.groups = append(p.groups, newGroup(name))

	return &p.groups[len(p.groups)-1], nil
}

// validateGroupName returns an error if name is blank, has surrounding
// whitespace, or contains brackets or line breaks.
func validateGroupName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("group name is blank")
	}

	if name != strings.TrimSpace(name) || strings.ContainsAny(name, "[]\n\r") {
		return fmt.Errorf("invalid group name: %q", name)
	}

	return nil
}
