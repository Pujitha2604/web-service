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
	"strings"
)

// Endpoint represents an API endpoint with method and path
type Endpoint struct {
	Method string
	Path   string
	Result string
}

// NewmanReport represents the structure of Newman collection run report
type NewmanReport struct {
	Run struct {
		Executions []struct {
			Item struct {
				Name    string `json:"name"`
				Request struct {
					Method string `json:"method"`
					URL    struct {
						Path []string `json:"path"`
					} `json:"url"`
				} `json:"request"`
			} `json:"item"`
			Response struct {
				Code int `json:"code"`
			} `json:"response"`
		} `json:"executions"`
	} `json:"run"`
}

// parseNewmanReport parses the Newman report and extracts endpoint status codes
func parseNewmanReport(reportPath string) (map[string]int, error) {
	// Read JSON file
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("error reading Newman report file: %v", err)
	}

	// Unmarshal JSON data
	var report NewmanReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("error parsing Newman report JSON: %v", err)
	}

	// Extract endpoint status codes
	endpoints := make(map[string]int)
	for _, exec := range report.Run.Executions {
		path := "/" + strings.Join(exec.Item.Request.URL.Path, "/")
		endpoints[path] = exec.Response.Code
	}

	return endpoints, nil
}

// AnalyzeEndpoints analyzes API endpoints in the specified directory
func AnalyzeEndpoints(rootDir string, newmanReportPath string) {
	// Track API endpoints across all handler files
	allEndpoints := make(map[string]Endpoint)

	// Analyze each handler file
	handlerFiles := []string{
		filepath.Join(rootDir, "handlers", "employeebyID.go"),
		filepath.Join(rootDir, "handlers", "getEmployees.go"),
		filepath.Join(rootDir, "handlers", "register.go"),
	}

	for _, file := range handlerFiles {
		endpoints := analyzeFileForAPIEndpoints(file)
		for endpoint, method := range endpoints {
			allEndpoints[endpoint] = Endpoint{Method: method, Path: endpoint, Result: "Not Covered"}
		}
	}

	// Parse Newman report
	newmanEndpoints, err := parseNewmanReport(newmanReportPath)
	if err != nil {
		log.Fatalf("Error parsing Newman report: %v", err)
	}

	// Compare handler endpoints with Newman endpoints
	for endpoint, details := range allEndpoints {
		// Check if the endpoint exists in Newman report
		if newmanStatus, exists := newmanEndpoints[details.Path]; exists {
			// Determine the result based on the HTTP status code
			if newmanStatus == 200 {
				allEndpoints[endpoint] = Endpoint{Method: details.Method, Path: details.Path, Result: "Success"}
			} else {
				allEndpoints[endpoint] = Endpoint{Method: details.Method, Path: details.Path, Result: "Failure"}
			}
		} else {
			// Check for a more flexible comparison like ignoring trailing slashes
			normalizedEndpoint := strings.TrimSuffix(endpoint, "/")
			if newmanStatus, exists := newmanEndpoints[normalizedEndpoint]; exists && newmanStatus == 200 {
				allEndpoints[endpoint] = Endpoint{Method: details.Method, Path: details.Path, Result: "Success"}
			} else {
				allEndpoints[endpoint] = Endpoint{Method: details.Method, Path: details.Path, Result: "Not Covered"}
			}
		}
	}

	// Print the report
	fmt.Println("| # | METHOD | PATH                        | RESULT      | SOURCE                |")
	fmt.Println("|---|--------|-----------------------------|-------------|------------------------|")

	count := 0
	for _, details := range allEndpoints {
		count++
		method := details.Method
		switch details.Path {
		case "/register":
			method = "POST"
		case "/employees", "/employee/6687b5635181f93273da46f1":
			method = "GET"
		}

		source := ""
		if details.Result != "Not Covered" {
			source = newmanReportPath
		}

		fmt.Printf("| %d | %s | %-28s | %-11s | %-22s |\n", count, method, details.Path, details.Result, source)
	}
}

// analyzeFileForAPIEndpoints analyzes a single Go file for API endpoints
func analyzeFileForAPIEndpoints(filename string) map[string]string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s: %v\n", filename, err)
		return nil
	}

	endpoints := make(map[string]string)

	// Inspect the AST
	ast.Inspect(node, func(n ast.Node) bool {
		// Check if the node is a function declaration
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			// Check if it's an HTTP handler registration
			if isHTTPHandler(funcDecl) {
				// Extract the endpoint path
				endpointPath := extractEndpointPath(funcDecl)
				if endpointPath != "" {
					endpoints[endpointPath] = funcDecl.Name.Name
				}
			}
		}
		return true
	})

	return endpoints
}

// isHTTPHandler checks if a FuncDecl is an HTTP handler registration
func isHTTPHandler(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Name == nil || funcDecl.Name.Name == "" {
		return false
	}
	// Check if the function name matches any of your handler functions
	switch funcDecl.Name.Name {
	case "RegisterEmployee", "EmployeeById", "Employees":
		return true
	default:
		return false
	}
}

// extractEndpointPath extracts the endpoint path from a FuncDecl
func extractEndpointPath(funcDecl *ast.FuncDecl) string {
	switch funcDecl.Name.Name {
	case "RegisterEmployee":
		return "/register"
	case "EmployeeById":
		return "/employee/6687b5635181f93273da46f1"
	case "Employees":
		return "/employees"
	default:
		return ""
	}
}

// RunNewman runs Newman with the specified collection file and generates a report
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
	rootDir := "C:/Users/Rekanto/Desktop/employee-service"
	collectionFile := filepath.Join(rootDir, "collection.json")
	newmanReportPath := "newman-report.json"

	// Run Newman and generate report
	if err := RunNewman(collectionFile, newmanReportPath); err != nil {
		log.Fatalf("Error running Newman: %v", err)
	}

	// Analyze endpoints based on generated Newman report
	AnalyzeEndpoints(rootDir, newmanReportPath)
}
