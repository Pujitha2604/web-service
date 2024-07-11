package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Endpoint struct {
	Method string
	Path   string
	Result string
}

type NewmanReport struct {
	Run Run `json:"run"`
}

type Run struct {
	Executions []Execution `json:"executions"`
}

type Execution struct {
	Item     Item     `json:"item"`
	Response Response `json:"response"`
}

type Item struct {
	Name    string  `json:"name"`
	Request Request `json:"request"`
}

type Request struct {
	Method string `json:"method"`
	URL    URL    `json:"url"`
}

type URL struct {
	Path []string `json:"path"`
}

type Response struct {
	Code int `json:"code"`
}

func analyzeFileForAPIEndpoints(rootDir string) map[string]Endpoint {
	endpoints := make(map[string]Endpoint)

	// Define a function to process each Go file
	processFile := func(filename string, info os.FileInfo, err error) error {
		// Check if the file is a Go source file
		if !info.IsDir() && strings.HasSuffix(filename, ".go") {

			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
			if err != nil {
				log.Printf("Error parsing file %s: %v\n", filename, err)
				return nil
			}

			// Inspect all function declarations
			for _, decl := range node.Decls {
				if fdecl, ok := decl.(*ast.FuncDecl); ok {
					// Check if there are comments associated with the function
					if fdecl.Doc != nil {
						var method, path string
						// Iterate through comments associated with the function
						for _, comment := range fdecl.Doc.List {
							text := strings.TrimSpace(comment.Text)
							if strings.HasPrefix(text, "//@Method:") {
								method = strings.TrimSpace(strings.TrimPrefix(text, "//@Method:"))
							} else if strings.HasPrefix(text, "//@Route:") {
								path = strings.TrimSpace(strings.TrimPrefix(text, "//@Route:"))
							}
						}
						// If both method and path are found, add endpoint to map
						if method != "" && path != "" {
							if !strings.HasPrefix(path, "/") {
								path = "/" + path
							}
							endpoints[path] = Endpoint{
								Method: method,
								Path:   path,
								Result: "Not Compared", // Default value before comparison
							}
						}
					}
				}
			}
		}
		return nil
	}

	// Walk through all files in the handlers directory
	err := filepath.Walk(rootDir, processFile)
	if err != nil {
		log.Printf("Error walking directory %s: %v\n", rootDir, err)
	}

	return endpoints
}

func AnalyzeEndpoints(rootDir string, newmanReportPath string) {
	allEndpoints := make(map[string]Endpoint)
	handlerFiles, err := getAllGoFiles(rootDir)
	if err != nil {
		log.Fatalf("Error retrieving handler files: %v", err)
	}

	for _, file := range handlerFiles {
		endpoints := analyzeFileForAPIEndpoints(file)
		for endpoint, endpointDetails := range endpoints {
			allEndpoints[endpoint] = Endpoint{
				Method: endpointDetails.Method,
				Path:   endpoint,
				Result: "Not Compared", // Default value before comparison
			}
		}
	}

	newmanEndpoints, err := parseNewmanReport(newmanReportPath)
	if err != nil {
		log.Fatalf("Error parsing Newman report: %v", err)
	}

	// Rest of the function remains unchanged
	for endpoint, details := range allEndpoints {
		for newmanEndpoint, newmanStatus := range newmanEndpoints {
			if matchEndpoint(endpoint, newmanEndpoint) {
				if newmanStatus == 200 {
					allEndpoints[endpoint] = Endpoint{
						Method: details.Method,
						Path:   details.Path,
						Result: "Success",
					}
				} else {
					allEndpoints[endpoint] = Endpoint{
						Method: details.Method,
						Path:   details.Path,
						Result: "Failure",
					}
				}
				break
			} else {
				allEndpoints[endpoint] = Endpoint{
					Method: details.Method,
					Path:   details.Path,
					Result: "Not Covered",
				}
			}
		}
	}

	printEndpointsTable(allEndpoints, newmanReportPath)
}

func printEndpointsTable(allEndpoints map[string]Endpoint, newmanReportPath string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "METHOD", "PATH", "RESULT", "SOURCE"})

	count := 0
	for _, details := range allEndpoints {
		count++
		method := details.Method
		source := ""
		if details.Result != "Not Covered" {
			source = newmanReportPath
		}
		t.AppendRow(table.Row{count, method, details.Path, details.Result, source})
	}

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Format.Header = text.FormatDefault
	t.Render()
}

func getAllGoFiles(dir string) ([]string, error) {
	var goFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %v", dir, err)
	}
	return goFiles, nil
}

func parseNewmanReport(reportPath string) (map[string]int, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("error reading Newman report file: %v", err)
	}

	var report NewmanReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("error parsing Newman report JSON: %v", err)
	}

	endpoints := make(map[string]int)
	for _, exec := range report.Run.Executions {
		path := exec.Item.Request.URL.Path
		if len(path) > 0 && path[0] == "employee" && len(path) > 1 && path[1] != "" {
			path = path[:1]
		}
		pathStr := "/" + strings.Join(path, "/")
		endpoints[pathStr] = exec.Response.Code
	}

	return endpoints, nil
}

func matchEndpoint(handlerEndpoint, newmanEndpoint string) bool {
	pattern := "^" + regexp.QuoteMeta(handlerEndpoint) + "$"
	matched, _ := regexp.MatchString(pattern, newmanEndpoint)
	return matched
}

func RunNewman(collectionFile string, reportPath string) error {
	cmd := exec.Command("newman", "run", collectionFile, "--reporters", "json", "--reporter-json-export", reportPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running Newman: %v", err)
	}
	return nil
}

func main() {
	// Check if root directory is provided as command-line argument
	if len(os.Args) < 2 {
		log.Fatal("Root directory not provided. Usage: go run main.go <rootDir>")
	}

	rootDir := os.Args[1]
	collectionFile := filepath.Join(rootDir, "collection.json")
	newmanReportPath := "newman-report.json"

	// Run Newman and generate report
	if err := RunNewman(collectionFile, newmanReportPath); err != nil {
		log.Fatalf("Error running Newman: %v", err)
	}

	// Analyze endpoints based on generated Newman report
	AnalyzeEndpoints(rootDir, newmanReportPath)
}
