package invites

import (
	"strings"
	"testing"
)

func TestParseCSV_ValidMinimalCSV(t *testing.T) {
	csvData := `email
john@example.com
jane@example.com
bob@example.com`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	if rows[0].Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", rows[0].Email)
	}

	if rows[0].Name != nil {
		t.Errorf("Expected nil name, got '%v'", rows[0].Name)
	}

	if rows[0].MaxPlusOnes != nil {
		t.Errorf("Expected nil max_plus_ones, got '%v'", rows[0].MaxPlusOnes)
	}
}

func TestParseCSV_ValidFullCSV(t *testing.T) {
	csvData := `name,email,max_plus_ones
John Doe,john@example.com,2
Jane Smith,jane@example.com,1
Bob Johnson,bob@example.com,0`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	if rows[0].Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", rows[0].Email)
	}

	if rows[0].Name == nil || *rows[0].Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%v'", rows[0].Name)
	}

	if rows[0].MaxPlusOnes == nil || *rows[0].MaxPlusOnes != 2 {
		t.Errorf("Expected max_plus_ones 2, got '%v'", rows[0].MaxPlusOnes)
	}
}

func TestParseCSV_MissingEmailColumn(t *testing.T) {
	csvData := `name,max_plus_ones
John Doe,2
Jane Smith,1`

	_, err := parseCSV([]byte(csvData))
	if err == nil {
		t.Fatal("Expected error for missing email column, got nil")
	}

	if !strings.Contains(err.Error(), "email") {
		t.Errorf("Expected error message to mention 'email', got '%s'", err.Error())
	}
}

func TestParseCSV_EmptyCSV(t *testing.T) {
	csvData := ``

	_, err := parseCSV([]byte(csvData))
	if err == nil {
		t.Fatal("Expected error for empty CSV, got nil")
	}
}

func TestParseCSV_OnlyHeader(t *testing.T) {
	csvData := `email`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

func TestParseCSV_ExceedsRowLimit(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("email\n")
	for i := 0; i < 501; i++ {
		sb.WriteString("user")
		sb.WriteString(strings.Repeat("0", 3-len(strings.Split(string(rune(i)), ""))))
		sb.WriteString("@example.com\n")
	}

	_, err := parseCSV([]byte(sb.String()))
	if err == nil {
		t.Fatal("Expected error for exceeding row limit, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error message to mention row limit, got '%s'", err.Error())
	}
}

func TestParseCSV_QuotedFields(t *testing.T) {
	csvData := `name,email,max_plus_ones
"Doe, John",john@example.com,2
"Smith, Jane",jane@example.com,1`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	if rows[0].Name == nil || *rows[0].Name != "Doe, John" {
		t.Errorf("Expected name 'Doe, John', got '%v'", rows[0].Name)
	}
}

func TestParseCSV_WhitespaceHandling(t *testing.T) {
	csvData := `email,name,max_plus_ones
  john@example.com  ,  John Doe  ,  2  
jane@example.com,Jane Smith,1`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rows[0].Email != "john@example.com" {
		t.Errorf("Expected trimmed email 'john@example.com', got '%s'", rows[0].Email)
	}

	if rows[0].Name == nil || *rows[0].Name != "John Doe" {
		t.Errorf("Expected trimmed name 'John Doe', got '%v'", rows[0].Name)
	}
}

func TestParseCSV_EmptyRows(t *testing.T) {
	csvData := `email
john@example.com

jane@example.com
`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows (empty rows skipped), got %d", len(rows))
	}
}

func TestParseCSV_MalformedCSV(t *testing.T) {
	csvData := `email,name
john@example.com,John Doe
jane@example.com`

	_, err := parseCSV([]byte(csvData))
	if err == nil {
		t.Fatal("Expected error for malformed CSV (missing field), got nil")
	}

	if !strings.Contains(err.Error(), "fewer fields") {
		t.Errorf("Expected error about fewer fields, got '%s'", err.Error())
	}
}

func TestParseCSV_InvalidMaxPlusOnes(t *testing.T) {
	csvData := `email,max_plus_ones
john@example.com,invalid
jane@example.com,2`

	_, err := parseCSV([]byte(csvData))
	if err == nil {
		t.Fatal("Expected error for invalid max_plus_ones, got nil")
	}
}

func TestParseCSV_NegativeMaxPlusOnes(t *testing.T) {
	csvData := `email,max_plus_ones
john@example.com,-1`

	_, err := parseCSV([]byte(csvData))
	if err == nil {
		t.Fatal("Expected error for negative max_plus_ones, got nil")
	}
}

func TestParseCSV_ExtraColumns(t *testing.T) {
	csvData := `email,name,max_plus_ones,extra_column
john@example.com,John Doe,2,extra_value
jane@example.com,Jane Smith,1,another_value`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error for extra columns, got %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}
}

func TestParseCSV_CSVInjectionPrevention(t *testing.T) {
	tests := []struct {
		name     string
		csvData  string
		wantName string
	}{
		{
			name: "equals sign prefix",
			csvData: `email,name
user@example.com,=1+1`,
			wantName: "'=1+1",
		},
		{
			name: "plus sign prefix",
			csvData: `email,name
user@example.com,+1+1`,
			wantName: "'+1+1",
		},
		{
			name: "minus sign prefix",
			csvData: `email,name
user@example.com,-1-1`,
			wantName: "'-1-1",
		},
		{
			name: "at sign prefix",
			csvData: `email,name
user@example.com,@SUM(A1:A10)`,
			wantName: "'@SUM(A1:A10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := parseCSV([]byte(tt.csvData))
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(rows) != 1 {
				t.Fatalf("Expected 1 row, got %d", len(rows))
			}

			if rows[0].Name == nil || *rows[0].Name != tt.wantName {
				t.Errorf("Expected sanitized name '%s', got '%v'", tt.wantName, rows[0].Name)
			}
		})
	}
}

func TestParseCSV_CaseInsensitiveHeaders(t *testing.T) {
	csvData := `EMAIL,NAME,MAX_PLUS_ONES
john@example.com,John Doe,2`

	rows, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}

	if rows[0].Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", rows[0].Email)
	}
}
