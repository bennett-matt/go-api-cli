/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	// "github.com/bennett-matt/go-api-cli/templates"
	"github.com/bennett-matt/go-api-cli/templates"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [project name]",
	Short: "Generate a new REST API project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		createProjectStructure(projectName)
		fmt.Printf("✅ REST API project '%s' created successfully!\n", projectName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func createProjectStructure(projectName string) {
	directories := []string{
		projectName,
		projectName + "/cmd",
		projectName + "/cmd/api",
		projectName + "/internal",
		projectName + "/internal/data",
		projectName + "/internal/validator",
		projectName + "/migrations",
	}

	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("❌ Failed to create directory %s: %v\n", dir, err)
		}
	}

	err := fs.WalkDir(templates.EmbeddedTemplates, "structure", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relpath := strings.TrimPrefix(strings.TrimSuffix(path, ".tmpl"), "structure/")
		outputPath := filepath.Join(projectName, relpath)
		if relpath == "migrations/init_db_setup.sql" {
			split := strings.Split(relpath, "/")
			outputPath = fmt.Sprintf("%s/migrations/%s_%s", projectName, time.Now().Format("20060102150405"), split[1])
		}

		return generateFileFromTemplate(path, outputPath, map[string]string{
			"projectName":    projectName,
			"envProjectName": toEnvVarName(projectName),
			"dbProjectName":  toDBName(projectName),
		})
	})
	if err != nil {
		log.Fatalln("❌ Error reading embedded files:", err)
	}

	cmd := exec.Command("go", "mod", "init", projectName)
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to initialize go module: %v\n", err)
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to run go mod tidy: %v\n", err)
	}

	cmd = exec.Command("go", "get", "github.com/ogen-go/ogen/cmd/ogen")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to get ogen dependency: %v\n", err)
	}

	cmd = exec.Command("go", "get", "github.com/ogen-go/ogen")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to get ogen dependency: %v\n", err)
	}

	cmd = exec.Command("make", "openapi-generate")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to generate files for OpenAPI: %v\n", err)
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to run go mod tidy: %v\n", err)
	}
}

func toEnvVarName(input string) string {
	input = strings.ToUpper(input)
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ReplaceAll(input, "-", "_")
	re := regexp.MustCompile(`[^A-Z0-9_]`)
	input = re.ReplaceAllString(input, "")
	return input
}

func toDBName(input string) string {
	input = strings.ToLower(input)
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ReplaceAll(input, "-", "_")
	re := regexp.MustCompile(`[^a-z0-9_]`)
	input = re.ReplaceAllString(input, "")
	input = regexp.MustCompile(`^[^a-z]+`).ReplaceAllString(input, "")
	input = regexp.MustCompile(`_+`).ReplaceAllString(input, "_")
	input = strings.Trim(input, "_")
	return input
}

func generateFileFromTemplate(templatePath, outputPath string, vars map[string]string) error {
	tmpl, err := template.ParseFS(templates.EmbeddedTemplates, templatePath)
	if err != nil {
		log.Fatalf("❌ Template parse error: %v\n", err)
	}

	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, vars)
}
