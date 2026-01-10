package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	files := []string{
		"internal/handlers/templates_test.go",
		"internal/handlers/images_test.go",
		"internal/handlers/invites_regenerate_test.go",
		"internal/handlers/invites_revoke_test.go",
		"internal/handlers/invites_import_permission_test.go",
		"internal/handlers/invites_list_test.go",
	}

	for _, filename := range files {
		if err := addAcceptHeaders(filename); err != nil {
			fmt.Printf("Error processing %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("Processed %s\n", filename)
	}
}

func addAcceptHeaders(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		
		// If this line creates an httptest.NewRequest
		if strings.Contains(line, "httptest.NewRequest(") {
			// Check if next line already has Accept header
			if scanner.Scan() {
				nextLine := scanner.Text()
				lines = append(lines, nextLine)
				
				if !strings.Contains(nextLine, `Header.Set("Accept"`) {
					// Add Accept header with same indentation as next line
					indent := ""
					for _, ch := range nextLine {
						if ch == '\t' || ch == ' ' {
							indent += string(ch)
						} else {
							break
						}
					}
					lines = append(lines, indent+`req.Header.Set("Accept", "application/json")`)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write back
	output, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	
	return writer.Flush()
}
