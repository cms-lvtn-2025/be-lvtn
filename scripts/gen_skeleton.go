package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Data struct {
	PackagePath string
	ProtoName   string
	ServiceName string
	Port        string
}

type HandlerData struct {
	PackagePath string
	ServiceName string
}

type Method struct {
	Name         string
	RequestType  string
	ResponseType string
}

type EntityHandlerData struct {
	PackagePath string
	EntityName  string
	Methods     []Method
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: go run gen_skeleton.go <proto_name> <service_name> <port>")
	}

	protoName := os.Args[1]   // vd: academic
	serviceName := os.Args[2] // vd: AcademicService
	port := os.Args[3]        // vd: 50051

	packagePath := "thaily/proto/" + protoName

	// Parse proto file to get RPC methods
	protoFile := filepath.Join("proto", protoName, protoName+".proto")
	methods, err := parseProtoFile(protoFile)
	if err != nil {
		log.Fatalf("Failed to parse proto file: %v", err)
	}

	// Group methods by entity
	entityMethods := groupMethodsByEntity(methods)

	// Create directories
	serviceDir := filepath.Join("src", "service", protoName)
	handlerDir := filepath.Join(serviceDir, "handler")
	os.MkdirAll(handlerDir, 0755)
	os.MkdirAll("env", 0755)
	os.MkdirAll("docker", 0755)

	data := Data{
		PackagePath: packagePath,
		ProtoName:   protoName,
		ServiceName: serviceName,
		Port:        port,
	}

	// Generate main.go
	generateMain(serviceDir, data)

	// Generate handler/handler.go
	generateHandlerRoot(handlerDir, HandlerData{
		PackagePath: packagePath,
		ServiceName: serviceName,
	})

	// Generate entity handler files
	for entityName, methods := range entityMethods {
		generateEntityHandler(handlerDir, EntityHandlerData{
			PackagePath: packagePath,
			EntityName:  entityName,
			Methods:     methods,
		})
	}

	// Generate env file
	generateEnvFile(protoName, data)

	// Generate Dockerfile
	generateDockerfile(protoName, data)

	log.Printf("Generated skeleton for %s service\n", serviceName)
}

func parseProtoFile(filename string) ([]Method, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var methods []Method
	scanner := bufio.NewScanner(file)
	rpcRegex := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*(\w+)\s*\)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := rpcRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			methods = append(methods, Method{
				Name:         matches[1],
				RequestType:  matches[2],
				ResponseType: matches[3],
			})
		}
	}

	return methods, scanner.Err()
}

func groupMethodsByEntity(methods []Method) map[string][]Method {
	entityMethods := make(map[string][]Method)

	for _, method := range methods {
		// Extract entity name from method name
		// e.g., CreateSemester -> Semester, ListSemesters -> Semester
		var entityName string
		for _, prefix := range []string{"Create", "Get", "Update", "Delete", "List"} {
			if strings.HasPrefix(method.Name, prefix) {
				entityName = strings.TrimPrefix(method.Name, prefix)
				break
			}
		}

		// Normalize to singular form for grouping
		// But keep compound names intact (e.g., RoleSystem, CouncilsSchedule)
		normalized := normalizePlural(entityName)

		if normalized != "" {
			entityMethods[normalized] = append(entityMethods[normalized], method)
		}
	}

	return entityMethods
}

func normalizePlural(name string) string {
	// Simple plural -> singular conversion (works for both simple and compound words)
	if strings.HasSuffix(name, "ies") && len(name) > 3 {
		return name[:len(name)-3] + "y" // Faculties -> Faculty
	}
	if strings.HasSuffix(name, "ses") && len(name) > 3 {
		return name[:len(name)-2] // Defences -> Defence
	}
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		// Check if the 's' is part of a compound word ending (e.g., "Status" should not become "Statu")
		// But "RoleSystems" should become "RoleSystem"
		return name[:len(name)-1] // Semesters -> Semester, RoleSystems -> RoleSystem
	}

	return name
}

func generateMain(serviceDir string, data Data) {
	tmpl, err := template.ParseFiles("template/main.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(filepath.Join(serviceDir, "main.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s/main.go\n", serviceDir)
}

func generateHandlerRoot(handlerDir string, data HandlerData) {
	tmpl, err := template.ParseFiles("template/handler.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(filepath.Join(handlerDir, "handler.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s/handler.go\n", handlerDir)
}

func generateEntityHandler(handlerDir string, data EntityHandlerData) {
	tmpl, err := template.ParseFiles("template/entity_handler.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	filename := strings.ToLower(data.EntityName) + ".go"
	out, err := os.Create(filepath.Join(handlerDir, filename))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s/%s\n", handlerDir, filename)
}

func generateEnvFile(protoName string, data Data) {
	tmpl, err := template.ParseFiles("template/env.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	filename := filepath.Join("env", protoName+".env")
	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s\n", filename)
}

func generateDockerfile(protoName string, data Data) {
	tmpl, err := template.ParseFiles("template/dockerfile.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	filename := filepath.Join("docker", protoName+".Dockerfile")
	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		log.Fatal(err)
	}

	log.Printf("Generated %s\n", filename)
}
