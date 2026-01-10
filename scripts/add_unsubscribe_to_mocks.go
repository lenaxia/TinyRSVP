package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	files := []string{
		"internal/handlers/invites_list_test.go",
		"internal/handlers/invites_delete_test.go",
		"internal/handlers/invites_get_test.go",
		"internal/handlers/invites_manual_test.go",
		"internal/handlers/invites_regenerate_test.go",
		"internal/handlers/invites_revoke_test.go",
		"internal/handlers/invites_send_test.go",
		"internal/handlers/invites_test.go",
		"internal/handlers/invites_update_test.go",
	}

	for _, filePath := range files {
		if err := addUnsubscribeMethod(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Updated %s\n", filePath)
		}
	}
}

func addUnsubscribeMethod(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	for i, line := range lines {
		if strings.Contains(line, "func (m *mock") && strings.Contains(line, "Service) MarkInviteResponded(ctx context.Context, inviteID int64) error {") {
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "return nil" {
				if i+2 < len(lines) && strings.TrimSpace(lines[i+2]) == "}" {
					newLines := make([]string, 0, len(lines)+4)
					newLines = append(newLines, lines[:i+3]...)
					newLines = append(newLines, "")
					
					receiverType := extractReceiverType(line)
					newLines = append(newLines, fmt.Sprintf("func (m *%s) UnsubscribeFromReminders(ctx context.Context, token string) error {", receiverType))
					newLines = append(newLines, "\treturn nil")
					newLines = append(newLines, "}")
					
					newLines = append(newLines, lines[i+3:]...)
					lines = newLines
					break
				}
			}
		}
	}

	output, err := os.Create(filePath)
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

func extractReceiverType(line string) string {
	start := strings.Index(line, "*") + 1
	end := strings.Index(line, ")")
	if start > 0 && end > start {
		return line[start:end]
	}
	return "mockService"
}
