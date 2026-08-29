package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	mappingsFileName = "mappings.conf"
)

func (p *Project) writeMappingsFile() error {
	if err := p.findHestiaDir(); err != nil {
		return err
	}

	mappingsPath := filepath.Join(p.root, mappingsFileName)
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
		tempFile.Close()
		os.Remove(tempPath)
		return err
	}

	for _, g := range p.groups {
		if _, err := tempFile.WriteString("[" + g.name + "]\n"); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return err
		}

		for _, m := range g.mappings {
			line := fmt.Sprintf("%q -> %q\n", m.src, m.dst)
			if _, err := tempFile.WriteString(line); err != nil {
				tempFile.Close()
				os.Remove(tempPath)
				return err
			}
		}
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := os.Rename(tempPath, mappingsPath); err != nil {
		os.Remove(tempPath)
		return err
	}

	return nil
}

func (p *Project) readMappingsFile(s Settings) error {
	if err := p.findHestiaDir(); err != nil {
		return err
	}

	filePath := filepath.Join(p.root, mappingsFileName)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	p.groups = []group{}

	var currentGroup *group
	for i, line := range strings.Split(string(fileBytes), "\n") {
		lineNumber := i + 1

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "[") {
			groupName, err := parseGroupLine(trimmed)
			if err != nil {
				return fmt.Errorf("%s:%d: %w", mappingsFileName, lineNumber, err)
			}
			if p.findGroupIndex(groupName) != -1 {
				return fmt.Errorf("%s:%d: duplicate group: %s", mappingsFileName, lineNumber, groupName)
			}

			p.groups = append(p.groups, newGroup(groupName))
			currentGroup = &p.groups[len(p.groups)-1]
			continue
		}

		src, dst, err := parseMappingLine(trimmed)
		if err != nil {
			return fmt.Errorf("%s:%d: %w", mappingsFileName, lineNumber, err)
		}
		if currentGroup == nil {
			return fmt.Errorf("%s:%d: mapping is without a group: %s", mappingsFileName, lineNumber, trimmed)
		}

		mapping, err := newMapping(p, s, src, dst)
		if err != nil {
			return fmt.Errorf("%s:%d: %w", mappingsFileName, lineNumber, err)
		}
		if err := currentGroup.addMapping(p, mapping); err != nil {
			return fmt.Errorf("%s:%d: %w", mappingsFileName, lineNumber, err)
		}
	}

	return nil
}

func parseGroupLine(line string) (string, error) {
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return "", fmt.Errorf("malformed group line: %s", line)
	}

	name := strings.TrimSpace(line[1 : len(line)-1])
	if name == "" || strings.ContainsAny(name, "[]") {
		return "", fmt.Errorf("malformed group name: %s", line)
	}

	return name, nil
}

func parseMappingLine(line string) (string, string, error) {
	var fields [2]string

	i := 0
	for f := range fields {
		fieldName := "source"
		if f == 1 {
			fieldName = "destination"
		}

		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}

		if f == 1 {
			if !strings.HasPrefix(line[i:], "->") {
				return "", "", fmt.Errorf("missing \"->\" separator: %s", line)
			}

			i += 2
			for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
				i++
			}
		}

		if i < len(line) && line[i] == '"' {
			end := i + 1
			for end < len(line) {
				if line[end] == '\\' {
					end += 2
					continue
				}
				if line[end] == '"' {
					break
				}
				end++
			}
			if end >= len(line) {
				return "", "", fmt.Errorf("unterminated quoted value: %s", line)
			}

			value, err := strconv.Unquote(line[i : end+1])
			if err != nil {
				return "", "", fmt.Errorf("malformed quoted value: %s", line[i:end+1])
			}
			if strings.TrimSpace(value) == "" {
				return "", "", fmt.Errorf("blank %s value: %s", fieldName, line)
			}

			fields[f] = value
			i = end + 1
			continue
		}

		start := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' && !strings.HasPrefix(line[i:], "->") {
			i++
		}
		if i == start {
			return "", "", fmt.Errorf("missing %s value: %s", fieldName, line)
		}

		fields[f] = line[start:i]
	}

	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i != len(line) {
		return "", "", fmt.Errorf("unexpected text after mapping: %s", line)
	}

	return fields[0], fields[1], nil
}
