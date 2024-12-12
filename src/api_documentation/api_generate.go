package main

//go:generate go run api_generate.go
import (
	"log"

	pkg "github.com/Parallels/prl-devops-service/api_documentation/package"
)

func main() {
	doc := pkg.NewApiDocument()
	doc.Content = "# API Documentation\n\nThis document describes the REST API for the service.\n\n"
	_, err := doc.Process()
	if err != nil {
		log.Fatalf("Failed to generate API Document: %v", err)
	}

	log.Println("API Document generated successfully!")
}
