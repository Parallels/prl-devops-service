package vagrant

import (
	"fmt"
	"strings"
)

type VagrantConfigBlock struct {
	Indent       int
	Type         string
	Name         string
	VariableName string
	Header       string
	Content      []string
	Children     []*VagrantConfigBlock
}

func (s *VagrantConfigBlock) NewBlock(typeName string, name string, varName string) *VagrantConfigBlock {
	result := &VagrantConfigBlock{
		Indent:       s.Indent + 1,
		Type:         typeName,
		Name:         name,
		VariableName: varName,
		Header:       fmt.Sprintf("%s \"%s\" do |%s|", typeName, name, varName),
		Content:      make([]string, 0),
		Children:     make([]*VagrantConfigBlock, 0),
	}

	s.Children = append(s.Children, result)
	return result
}

func (c *VagrantConfigBlock) String() string {
	result := ""

	result += fmt.Sprintf("%s%s\n", strings.Repeat("  ", c.Indent), c.Header)

	for _, line := range c.Content {
		lineIndent := c.Indent + 1
		result += fmt.Sprintf("%s%s\n", strings.Repeat("  ", lineIndent), line)
	}

	for _, child := range c.Children {
		result += "\n"
		result += child.String()
	}

	result += fmt.Sprintf("%s%s\n", strings.Repeat("  ", c.Indent), "end")
	return result
}

func (c *VagrantConfigBlock) SetHeader(header string) {
	parts := strings.Split(header, " do ")
	c.Header = header
	if len(parts) > 1 {
		typeParts := strings.Split(parts[0], " ")
		if strings.HasPrefix(typeParts[0], "Vagrant.configure") {
			c.Type = "Vagrant.configure"
			namePart := strings.Split(typeParts[0], "(")
			if len(namePart) > 1 {
				c.Name = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(namePart[1]), ")", ""), "\"", "")
			} else {
				c.Name = ""
			}
		} else {
			c.Type = strings.TrimSpace(typeParts[0])
			if len(typeParts) > 1 {
				name := strings.Join(typeParts[1:], " ")
				name = strings.ReplaceAll(name, "\"", "")
				c.Name = name
			}
		}
		varParts := strings.Split(parts[1], "|")
		if len(varParts) > 1 {
			c.VariableName = strings.TrimSpace(varParts[1])
		} else {
			c.VariableName = strings.Trim(varParts[0], "|")
		}
	}
}

func (c *VagrantConfigBlock) AddContent(content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	lines := strings.Split(content, "\n")
	c.Content = append(c.Content, lines...)
}

func (c *VagrantConfigBlock) GetContent() string {
	result := ""

	for _, line := range c.Content {
		result += fmt.Sprintf("%s\n", line)
	}

	for _, child := range c.Children {
		result += "\n"
		result += child.GetContent()
	}

	return result
}

func (c *VagrantConfigBlock) GetContentVariable(name string) string {
	name = fmt.Sprintf("%s.%s", c.VariableName, name)
	for _, line := range c.Content {
		parts := strings.Split(line, "=")
		if len(parts) > 1 && strings.EqualFold(strings.TrimSpace(parts[0]), name) {
			return strings.ReplaceAll(strings.TrimSpace(parts[1]), "\"", "")
		}
	}

	return ""
}

func (c *VagrantConfigBlock) SetContentVariable(name string, value string) {
	name = fmt.Sprintf("%s.%s", c.VariableName, name)
	for index, line := range c.Content {
		parts := strings.Split(line, "=")
		if len(parts) > 1 && strings.EqualFold(strings.TrimSpace(parts[0]), name) {
			c.Content[index] = fmt.Sprintf("%s = \"%s\"", name, value)
			return
		}
	}

	c.Content = append(c.Content, fmt.Sprintf("%s = \"%s\"", name, value))
}
