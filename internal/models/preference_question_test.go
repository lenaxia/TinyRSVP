package models

import (
	"testing"
)

func TestQuestionType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		qtype QuestionType
		valid bool
	}{
		{"text type", QuestionTypeText, true},
		{"single choice type", QuestionTypeSingleChoice, true},
		{"multiple choice type", QuestionTypeMultipleChoice, true},
		{"invalid type", QuestionType("invalid"), false},
		{"empty type", QuestionType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.qtype == QuestionTypeText ||
				tt.qtype == QuestionTypeSingleChoice ||
				tt.qtype == QuestionTypeMultipleChoice

			if valid != tt.valid {
				t.Errorf("QuestionType.Valid() = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestPreferenceQuestion_ParseOptions(t *testing.T) {
	tests := []struct {
		name    string
		options *string
		want    []string
		wantErr bool
	}{
		{
			name:    "valid options",
			options: stringPtr(`["Option 1", "Option 2", "Option 3"]`),
			want:    []string{"Option 1", "Option 2", "Option 3"},
			wantErr: false,
		},
		{
			name:    "nil options",
			options: nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "empty string",
			options: stringPtr(""),
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			options: stringPtr("not json"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty array",
			options: stringPtr("[]"),
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "single option",
			options: stringPtr(`["Only one"]`),
			want:    []string{"Only one"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &PreferenceQuestion{Options: tt.options}
			got, err := q.ParseOptions()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParseOptions() len = %d, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseOptions()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestPreferenceQuestion_SetOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []string
		wantErr bool
	}{
		{
			name:    "valid options",
			options: []string{"Option 1", "Option 2"},
			wantErr: false,
		},
		{
			name:    "nil options",
			options: nil,
			wantErr: false,
		},
		{
			name:    "empty options",
			options: []string{},
			wantErr: false,
		},
		{
			name:    "single option",
			options: []string{"Only one"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &PreferenceQuestion{}
			err := q.SetOptions(tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.options != nil {
				got, err := q.ParseOptions()
				if err != nil {
					t.Errorf("ParseOptions() after SetOptions() error = %v", err)
					return
				}

				if len(got) != len(tt.options) {
					t.Errorf("After SetOptions() len = %d, want %d", len(got), len(tt.options))
					return
				}

				for i := range got {
					if got[i] != tt.options[i] {
						t.Errorf("After SetOptions()[%d] = %v, want %v", i, got[i], tt.options[i])
					}
				}
			}
		})
	}
}

func TestPreferenceQuestion_HelpText(t *testing.T) {
	tests := []struct {
		name     string
		helpText *string
		want     string
		isNil    bool
	}{
		{
			name:     "with help text",
			helpText: stringPtr("This is helpful information"),
			want:     "This is helpful information",
			isNil:    false,
		},
		{
			name:     "nil help text",
			helpText: nil,
			want:     "",
			isNil:    true,
		},
		{
			name:     "empty help text",
			helpText: stringPtr(""),
			want:     "",
			isNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &PreferenceQuestion{HelpText: tt.helpText}

			if tt.isNil && q.HelpText != nil {
				t.Errorf("HelpText should be nil, got %v", q.HelpText)
				return
			}

			if !tt.isNil && q.HelpText == nil {
				t.Errorf("HelpText should not be nil")
				return
			}

			if !tt.isNil && *q.HelpText != tt.want {
				t.Errorf("HelpText = %v, want %v", *q.HelpText, tt.want)
			}
		})
	}
}
