package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const dtoTemplate = `package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

{{.Border}}
// ||         {{.DTOName}} Request         ||
{{.Border}}

type {{.DTOName}}Request struct {
}

func New{{.DTOName}}Request() *{{.DTOName}}Request {
	return &{{.DTOName}}Request{}
}

func (l *{{.DTOName}}Request) GetValue() *{{.DTOName}}Request {
	return l
}

func (s *{{.DTOName}}Request) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", err.Field()))
		}
	}
	return msgs, nil
}

{{.Border}}
// ||         {{.DTOName}} Response        ||
{{.Border}}

type {{.DTOName}}Response struct {
}

func New{{.DTOName}}Response() *{{.DTOName}}Response {
	return &{{.DTOName}}Response{}
}

func (l *{{.DTOName}}Response) GetValue() *{{.DTOName}}Response {
	return l
}

func (l *{{.DTOName}}Response) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		default:
			msgs = append(msgs, err.Field()+" is invalid")
		}
	}
	return msgs, nil
}
`

type TemplateData struct {
	PackageName string
	DTOName     string
	Border      string
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter file name (e.g., login.go): ")
	fileName, _ := reader.ReadString('\n')
	fileName = strings.TrimSpace(fileName)

	fmt.Print("Enter folder name for dto (e.g., auth): ")
	relativePath, _ := reader.ReadString('\n')
	relativePath = strings.TrimSpace(relativePath)

	fmt.Print("Enter DTO name (e.g., Login): ")
	dtoName, _ := reader.ReadString('\n')
	dtoName = strings.TrimSpace(dtoName)

	// ✅ Resolve full relative path from .tools (so "../" points to root)
	fullRelativePath := filepath.Join("api/", relativePath, "dto")

	// ✅ Check if path exists
	if _, err := os.Stat(fullRelativePath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: The relative path '%s' does not exist. Please create it first.\n", fullRelativePath)
		return
	}

	packageName := getPackageName(relativePath)
	fullFilePath := filepath.Join(fullRelativePath, fileName)

	// Override existing file if it exists
	if _, err := os.Stat(fullFilePath); err == nil {
		fmt.Printf("⚠️  Warning: The file '%s' already exists. It will be overridden.\n", fullFilePath)
		fmt.Print("Do you want to continue? (y/n): ")
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(confirmation)

		if confirmation != "y" {
			fmt.Println("❌ Operation cancelled.")
			return
		}
	}

	tmpl, err := template.New("dto").Parse(dtoTemplate)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(fullFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data := TemplateData{
		PackageName: packageName,
		DTOName:     dtoName,
		Border:      buildBorder(dtoName),
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("✅ DTO file generated: %s\n", fullFilePath)
}

func getPackageName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func buildBorder(dtoName string) string {
	maxContentLen := len(dtoName) + len(" Response")
	totalWidth := maxContentLen + 20
	return "// " + strings.Repeat("=", totalWidth)
}
