package vagrant

import (
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/cjlapao/common-go/helper"
)

type VagrantFile struct {
	ctx  basecontext.ApiContext
	path string
	Root *VagrantConfigBlock
}

func NewVagrantFile(ctx basecontext.ApiContext, filePath string) *VagrantFile {
	result := &VagrantFile{
		ctx:  ctx,
		path: filePath,
		Root: &VagrantConfigBlock{
			Content:  make([]string, 0),
			Children: make([]*VagrantConfigBlock, 0),
		},
	}

	result.Root.SetHeader("Vagrant.configure(\"2\") do |config|")

	return result
}

func LoadVagrantFile(ctx basecontext.ApiContext, filePath string) (*VagrantFile, error) {
	if !helper.FileExists(filePath) {
		return nil, errors.Newf("Vagrant file %v does not exist", filePath)
	}

	result := &VagrantFile{
		ctx:  ctx,
		path: filePath,
	}

	contentBytes, err := helper.ReadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(contentBytes)

	_, rootBlock := result.extractConfigBlock(content, 0)

	result.Root = rootBlock

	return result, nil
}

func (s *VagrantFile) Save() error {
	if s.path == "" {
		return errors.New("Cannot save vagrant file without a path")
	}

	return helper.WriteToFile(s.Root.String(), s.path)
}

func (s *VagrantFile) String() string {
	return s.Root.String()
}

func (s *VagrantFile) Refresh() {
	content := s.Root.String()
	_, rootBlock := s.extractConfigBlock(content, 0)
	s.Root = rootBlock
}

func (s *VagrantFile) GetConfigBlock(name string) []*VagrantConfigBlock {
	result := make([]*VagrantConfigBlock, 0)
	if strings.EqualFold(s.Root.Name, name) {
		result = append(result, s.Root)
	}

	for _, child := range s.Root.Children {
		if strings.EqualFold(child.Name, name) {
			result = append(result, child)
		}
	}

	return result
}

func (s *VagrantFile) GetConfigBlockByTypeName(typeName string) []*VagrantConfigBlock {
	result := make([]*VagrantConfigBlock, 0)
	if strings.EqualFold(s.Root.Type, typeName) {
		result = append(result, s.Root)
	}

	for _, child := range s.Root.Children {
		if strings.EqualFold(child.Type, typeName) {
			result = append(result, child)
		}
	}

	return result
}

func (s *VagrantFile) extractConfigBlock(content string, nestingLevel int) (string, *VagrantConfigBlock) {
	result := VagrantConfigBlock{
		Name:     "",
		Content:  make([]string, 0),
		Children: make([]*VagrantConfigBlock, 0),
	}

	inBlock := false
	lines := strings.Split(content, "\n")
	index := 0
	for {
		if len(lines) == 0 {
			break
		}

		trimmed := strings.TrimSpace(lines[index])

		if strings.Contains(trimmed, " do ") {
			if inBlock {
				content, child := s.extractConfigBlock(strings.Join(lines, "\n"), nestingLevel+1)
				result.Children = append(result.Children, child)
				lines = strings.Split(content, "\n")
				continue
			}

			parts := strings.Split(trimmed, "|")
			if len(parts) > 1 {
				result.VariableName = strings.TrimSpace(parts[1])
			} else {
				result.VariableName = strings.ReplaceAll(strings.TrimSpace(trimmed), "\"", "")
			}
			typeParts := strings.Split(parts[0], " ")
			if len(typeParts) > 1 {
				result.Type = strings.TrimSpace(typeParts[0])
				if strings.TrimSpace(typeParts[1]) != "do" {
					result.Name = strings.ReplaceAll(strings.TrimSpace(typeParts[1]), "\"", "")
				}
			}
			inBlock = true
			result.SetHeader(trimmed)
			result.Indent = nestingLevel
			lines = append(lines[:index], lines[index+1:]...)
			continue
		} else if trimmed == "end" {
			lines = append(lines[:index], lines[index+1:]...)
			return strings.Join(lines, "\n"), &result
		}

		if inBlock && trimmed != "" {
			result.Content = append(result.Content, trimmed)
		}

		lines = append(lines[:index], lines[index+1:]...)
	}

	return strings.Join(lines, "\n"), &result
}
