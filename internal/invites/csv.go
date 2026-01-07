package invites

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type CSVRow struct {
	Email       string
	Name        *string
	MaxPlusOnes *int
}

func parseCSV(data []byte) ([]CSVRow, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, &models.ValidationError{
			Field:   "csv",
			Message: "CSV file is empty",
		}
	}

	header := records[0]
	emailIdx := findColumnIndex(header, "email")
	if emailIdx == -1 {
		return nil, &models.ValidationError{
			Field:   "csv",
			Message: "CSV must have 'email' column",
		}
	}

	nameIdx := findColumnIndex(header, "name")
	maxPlusOnesIdx := findColumnIndex(header, "max_plus_ones")

	dataRows := records[1:]
	if len(dataRows) > 500 {
		return nil, &models.ValidationError{
			Field:   "csv",
			Message: "CSV exceeds 500 row limit",
		}
	}

	var rows []CSVRow
	for i, record := range dataRows {
		if isEmptyRow(record) {
			continue
		}

		if len(record) < len(header) {
			return nil, &models.ValidationError{
				Field:   "csv",
				Message: fmt.Sprintf("row %d has fewer fields than header", i+2),
			}
		}

		row := CSVRow{
			Email: strings.TrimSpace(record[emailIdx]),
		}

		if nameIdx != -1 && nameIdx < len(record) {
			name := strings.TrimSpace(record[nameIdx])
			if name != "" {
				sanitized := sanitizeCSVField(name)
				row.Name = &sanitized
			}
		}

		if maxPlusOnesIdx != -1 && maxPlusOnesIdx < len(record) {
			maxPlusOnesStr := strings.TrimSpace(record[maxPlusOnesIdx])
			if maxPlusOnesStr != "" {
				maxPlusOnes, err := strconv.Atoi(maxPlusOnesStr)
				if err != nil {
					return nil, &models.ValidationError{
						Field:   "max_plus_ones",
						Message: fmt.Sprintf("row %d: invalid max_plus_ones value '%s'", i+2, maxPlusOnesStr),
					}
				}
				if maxPlusOnes < 0 {
					return nil, &models.ValidationError{
						Field:   "max_plus_ones",
						Message: fmt.Sprintf("row %d: max_plus_ones cannot be negative", i+2),
					}
				}
				row.MaxPlusOnes = &maxPlusOnes
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func findColumnIndex(header []string, columnName string) int {
	for i, col := range header {
		if strings.EqualFold(strings.TrimSpace(col), columnName) {
			return i
		}
	}
	return -1
}

func isEmptyRow(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

func sanitizeCSVField(field string) string {
	if len(field) > 0 {
		firstChar := field[0]
		if firstChar == '=' || firstChar == '+' || firstChar == '-' || firstChar == '@' {
			return "'" + field
		}
	}
	return field
}
