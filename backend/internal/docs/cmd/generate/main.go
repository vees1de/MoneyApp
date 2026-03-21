package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	const (
		sourcePath = "openapi.yaml"
		targetPath = "swagger.json"
	)

	input, err := os.ReadFile(sourcePath)
	if err != nil {
		fail("read %s: %v", sourcePath, err)
	}

	var document any
	if err := yaml.Unmarshal(input, &document); err != nil {
		fail("parse %s: %v", sourcePath, err)
	}

	output, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		fail("marshal %s: %v", targetPath, err)
	}
	output = append(output, '\n')

	if err := os.WriteFile(targetPath, output, 0o644); err != nil {
		fail("write %s: %v", targetPath, err)
	}
}

func fail(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
