package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func (p *Project) writeMappingsFile() error {
	if err := p.findProjectRoot(); err != nil {
		return err
	}

	mappingsPath := filepath.Join(p.root, mappingsFileName)
	log.Debugf("writing mappings to: %v", mappingsPath)
	mappingsInfo, err := os.Stat(mappingsPath)
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(p.root, mappingsFileName+".*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Chmod(mappingsInfo.Mode().Perm()); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return err
	}

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(p.groups); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return err
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return err
	}

	if err := os.Rename(tempPath, mappingsPath); err != nil {
		_ = os.Remove(tempPath)
		return err
	}

	return nil
}

func (p *Project) readMappingsFile(s Settings) error {
	if err := p.findProjectRoot(); err != nil {
		return err
	}

	filePath := filepath.Join(p.root, mappingsFileName)
	log.Debugf("reading mappings from: %v", filePath)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	p.groups = []*group{}

	var groups []*group
	if err := json.Unmarshal(fileBytes, &groups); err != nil {
		return fmt.Errorf("%s: %w", mappingsFileName, err)
	}

	for _, g := range groups {
		if g == nil {
			return fmt.Errorf("%s: null group", mappingsFileName)
		}
		if err := validateGroupName(g.Name); err != nil {
			return fmt.Errorf("%s: %w", mappingsFileName, err)
		}
		if p.findGroupIndex(g.Name) != -1 {
			return fmt.Errorf("%s: duplicate group: %s", mappingsFileName, g.Name)
		}

		currentGroup := newGroup(g.Name)
		p.groups = append(p.groups, currentGroup)
		for _, m := range g.Mappings {
			if m == nil || strings.TrimSpace(m.Src) == "" || strings.TrimSpace(m.Dst) == "" {
				return fmt.Errorf("%s: blank source or destination in group: %s", mappingsFileName, g.Name)
			}
			mapping, err := newMapping(p, s, m.Src, m.Dst)
			if err != nil {
				return fmt.Errorf("%s: %w", mappingsFileName, err)
			}
			if err := currentGroup.addMapping(p, mapping); err != nil {
				return fmt.Errorf("%s: %w", mappingsFileName, err)
			}
		}
	}

	return nil
}
