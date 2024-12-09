package pkg

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Parallels/prl-devops-service/api_documentation/package/comment_processor"
	"github.com/Parallels/prl-devops-service/api_documentation/package/examples_processor"
	"github.com/Parallels/prl-devops-service/api_documentation/package/helpers"
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
	"gopkg.in/yaml.v3"
)

const (
	emptyCategory = "General"
)

type ApiDocument struct {
	Layout             string             `yaml:"layout"`
	Title              string             `yaml:"title"`
	DefaultHost        string             `yaml:"default_host,omitempty"`
	ApiPrefix          string             `yaml:"api_prefix,omitempty"`
	OutputFolder       string             `yaml:"-"`
	ExportCategories   bool               `yaml:"-"`
	IsCategoryDocument bool               `yaml:"is_category_document,omitempty"`
	Categories         []*models.Category `yaml:"categories,omitempty"`
	Content            string             `yaml:"-"`
	Endpoints          []*models.Endpoint `yaml:"endpoints"`
}

func NewApiDocument() *ApiDocument {
	doc := &ApiDocument{
		Layout:           "api",
		Title:            "API Documentation",
		Endpoints:        []*models.Endpoint{},
		DefaultHost:      "http://localhost",
		OutputFolder:     "../../docs/rest-api",
		ExportCategories: true,
		Categories:       []*models.Category{},
		ApiPrefix:        "/api",
	}
	doc.Categories = append(doc.Categories, &models.Category{
		Name:      "General",
		Path:      "general",
		Endpoints: []models.CategoryEndpoint{},
	})

	return doc
}

func (d *ApiDocument) String() (string, error) {
	output := "---\n"
	for i, endpoint := range d.Endpoints {
		if len(endpoint.Content) > 0 {
			markdown := strings.Join(endpoint.Content, "\n")
			d.Endpoints[i].MarkdownContent = markdown
		}
	}
	categoriesWithEndpoints := make([]*models.Category, 0)
	for _, category := range d.Categories {
		if len(category.Endpoints) == 0 {
			continue
		}
		categoriesWithEndpoints = append(categoriesWithEndpoints, category)
	}
	d.Categories = categoriesWithEndpoints

	content, err := yaml.Marshal(d)
	if err != nil {
		return "", err
	}
	output += string(content)
	output += "\n---"
	if d.Content != "" {
		output += "\n"
		output += d.Content
		output += "\n"
	}

	return output, nil
}

func (d *ApiDocument) Process() (*ApiDocument, error) {
	endpoints, err := d.extractBlocks()
	if err != nil {
		return nil, err
	}

	d.Endpoints = append(d.Endpoints, endpoints...)
	if err := d.Save(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *ApiDocument) addCategory(category string) *models.Category {
	for _, documentCategory := range d.Categories {
		if strings.EqualFold(documentCategory.Name, category) {
			return documentCategory
		}
	}

	newCategory := models.Category{
		Name:      category,
		Path:      helpers.NormalizeString(category),
		Endpoints: []models.CategoryEndpoint{},
	}
	d.Categories = append(d.Categories, &newCategory)

	return &newCategory
}

func (d *ApiDocument) getCategoryEndpoints(category string) []*models.Endpoint {
	var categoryEndpoints []*models.Endpoint
	for _, endpoint := range d.Endpoints {
		if strings.EqualFold(endpoint.Category, category) {
			categoryEndpoints = append(categoryEndpoints, endpoint)
		}
	}

	return categoryEndpoints
}

func (d *ApiDocument) Save() error {
	// Creating the folder if it does not exist
	if _, err := os.Stat(d.OutputFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(d.OutputFolder, os.ModePerm); err != nil {
			return err
		}
	}

	indexFilename := filepath.Join(d.OutputFolder, "index.md")
	data, err := d.String()
	if err != nil {
		return err
	}
	if err := os.WriteFile(indexFilename, []byte(data), 0o644); err != nil {
		return err
	}

	if d.ExportCategories && !d.IsCategoryDocument {
		for _, category := range d.Categories {
			categoryEndpoints := d.getCategoryEndpoints(category.Name)
			if len(categoryEndpoints) == 0 {
				continue
			}

			categoryFolder := filepath.Join(d.OutputFolder, helpers.NormalizeString(category.Name))
			categoryDocument := NewApiDocument()
			categoryDocument.Title = category.Name
			categoryDocument.Endpoints = categoryEndpoints
			categoryDocument.DefaultHost = d.DefaultHost
			categoryDocument.ApiPrefix = d.ApiPrefix
			categoryDocument.OutputFolder = categoryFolder
			categoryDocument.ExportCategories = false
			categoryDocument.IsCategoryDocument = true
			categoryDocument.Categories = d.Categories
			categoryDocument.Content = "# " + category.Name + " endpoints \n\n This document contains the endpoints for the " + category.Name + " category.\n\n"
			categoryDocument.Save()
		}
	}

	return nil
}

func (d *ApiDocument) extractBlocks() ([]*models.Endpoint, error) {
	var endpoints []*models.Endpoint
	root := "../"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".go") || !strings.Contains(info.Name(), "main.go") {
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}

			scanner := bufio.NewScanner(file)
			var buffer bytes.Buffer

			inCommentBlock := false

			// Parse the comments
			for scanner.Scan() {
				line := scanner.Text()
				rgx := regexp.MustCompile(`^//[\s\t]*@`)
				if rgx.MatchString(line) {
					buffer.WriteString(line + "\n")
					inCommentBlock = true
				} else if inCommentBlock && strings.TrimSpace(line) == "" {
					// Process the accumulated comment block
					endpoint := d.parseComments(buffer.String(), path)
					if endpoint.Path == "" {
						continue
					}
					if endpoint.Category == "" {
						endpoint.Category = emptyCategory
					}
					endpoint.CategoryPath = helpers.NormalizeString(endpoint.Category)

					cat := d.addCategory(endpoint.Category)
					cat.AddEndpoint(endpoint)
					endpoints = append(endpoints, &endpoint)
					buffer.Reset()
					inCommentBlock = false
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

func (d *ApiDocument) parseComments(commentBlock string, fileName string) models.Endpoint {
	var endpoint models.Endpoint
	lines := strings.Split(commentBlock, "\n")
	commentProcessor := comment_processor.NewCommentProcessor()
	examples_processor := examples_processor.NewExamplesProcessor()
	for _, line := range lines {
		if err := commentProcessor.ProcessComment(&endpoint, fileName, line); err != nil {
			log.Fatal(err)
		}
	}

	endpoint.RequiresAuth = true
	endpoint.ApiPrefix = d.ApiPrefix
	endpoint.HostUrl = d.DefaultHost
	if err := examples_processor.Process(&endpoint); err != nil {
		log.Fatal(err)
	}

	return endpoint
}
