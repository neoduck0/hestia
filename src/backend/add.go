package backend

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

	newMapping, err := newMapping(p, s, src, dst)
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
