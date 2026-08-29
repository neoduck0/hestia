package backend

import (
	"fmt"
	"maps"
	"strings"

	"github.com/charmbracelet/log"
)

func (p *Project) Link(s Settings, selectedGroups map[string]struct{}, all, exclude bool) error {
	if err := p.readMappingsFile(s); err != nil {
		return err
	}

	if err := p.validateSelectedGroups(selectedGroups); err != nil {
		return err
	}

	if all {
		if err := p.linkAll(s); err != nil {
			return err
		}
	} else if exclude {
		if err := p.linkExclude(s, selectedGroups); err != nil {
			return err
		}
	} else {
		if err := p.linkInclude(s, selectedGroups); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) linkAll(s Settings) error {
	for _, group := range p.groups {
		err := group.link(p, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) linkInclude(s Settings, selectedGroups map[string]struct{}) error {
	for _, group := range p.groups {
		if _, ok := selectedGroups[group.name]; !ok {
			log.Debugf("group not included: %v", group.name)
			continue
		}

		err := group.link(p, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) linkExclude(s Settings, selectedGroups map[string]struct{}) error {
	for _, group := range p.groups {
		if _, ok := selectedGroups[group.name]; ok {
			log.Debugf("group excluded: %v", group.name)
			continue
		}

		err := group.link(p, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) validateSelectedGroups(selectedGroups map[string]struct{}) error {
	if len(selectedGroups) == 0 {
		return nil
	}

	notFound := maps.Clone(selectedGroups)

	for _, group := range p.groups {
		delete(notFound, group.name)
	}

	if len(notFound) == 0 {
		return nil
	}

	msg := strings.Builder{}
	for group := range notFound {
		if msg.String() != "" {
			msg.WriteRune(',')
		}
		msg.WriteString(group)
	}

	return fmt.Errorf("unknown groups: %s", msg.String())
}
