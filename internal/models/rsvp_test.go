package models

import (
	"strings"
	"testing"
)

func TestRSVPResponse_Valid(t *testing.T) {
	tests := []struct {
		name     string
		response RSVPResponse
		valid    bool
	}{
		{"yes response", RSVPResponseYes, true},
		{"no response", RSVPResponseNo, true},
		{"maybe response", RSVPResponseMaybe, true},
		{"invalid response", RSVPResponse("invalid"), false},
		{"empty response", RSVPResponse(""), false},
		{"uppercase YES", RSVPResponse("YES"), false},
		{"mixed case Maybe", RSVPResponse("Maybe"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.response.Valid()
			if valid != tt.valid {
				t.Errorf("Valid() = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestRSVP_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rsvp    *RSVP
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid yes response with plus ones",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponseYes,
				PlusOnes: 2,
			},
			wantErr: false,
		},
		{
			name: "valid no response",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponseNo,
				PlusOnes: 0,
			},
			wantErr: false,
		},
		{
			name: "valid maybe response",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponseMaybe,
				PlusOnes: 1,
			},
			wantErr: false,
		},
		{
			name: "zero invite ID",
			rsvp: &RSVP{
				InviteID: 0,
				Response: RSVPResponseYes,
				PlusOnes: 0,
			},
			wantErr: true,
			errMsg:  "invite_id is required",
		},
		{
			name: "invalid response",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponse("invalid"),
				PlusOnes: 0,
			},
			wantErr: true,
			errMsg:  "response must be yes, no, or maybe",
		},
		{
			name: "empty response",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponse(""),
				PlusOnes: 0,
			},
			wantErr: true,
			errMsg:  "response must be yes, no, or maybe",
		},
		{
			name: "negative plus ones",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponseYes,
				PlusOnes: -1,
			},
			wantErr: true,
			errMsg:  "plus_ones cannot be negative",
		},
		{
			name: "zero plus ones is valid",
			rsvp: &RSVP{
				InviteID: 1,
				Response: RSVPResponseYes,
				PlusOnes: 0,
			},
			wantErr: false,
		},
		{
			name: "valid with adults count",
			rsvp: &RSVP{
				InviteID:    1,
				Response:    RSVPResponseYes,
				PlusOnes:    2,
				AdultsCount: func() *int { i := 2; return &i }(),
			},
			wantErr: false,
		},
		{
			name: "valid with kids count",
			rsvp: &RSVP{
				InviteID:  1,
				Response:  RSVPResponseYes,
				PlusOnes:  2,
				KidsCount: func() *int { i := 1; return &i }(),
			},
			wantErr: false,
		},
		{
			name: "valid with both adults and kids count",
			rsvp: &RSVP{
				InviteID:    1,
				Response:    RSVPResponseYes,
				PlusOnes:    3,
				AdultsCount: func() *int { i := 2; return &i }(),
				KidsCount:   func() *int { i := 1; return &i }(),
			},
			wantErr: false,
		},
		{
			name: "negative adults count",
			rsvp: &RSVP{
				InviteID:    1,
				Response:    RSVPResponseYes,
				PlusOnes:    0,
				AdultsCount: func() *int { i := -1; return &i }(),
			},
			wantErr: true,
			errMsg:  "adults_count cannot be negative",
		},
		{
			name: "negative kids count",
			rsvp: &RSVP{
				InviteID:  1,
				Response:  RSVPResponseYes,
				PlusOnes:  0,
				KidsCount: func() *int { i := -1; return &i }(),
			},
			wantErr: true,
			errMsg:  "kids_count cannot be negative",
		},
		{
			name: "zero adults count is valid",
			rsvp: &RSVP{
				InviteID:    1,
				Response:    RSVPResponseYes,
				PlusOnes:    0,
				AdultsCount: func() *int { i := 0; return &i }(),
			},
			wantErr: false,
		},
		{
			name: "zero kids count is valid",
			rsvp: &RSVP{
				InviteID:  1,
				Response:  RSVPResponseYes,
				PlusOnes:  0,
				KidsCount: func() *int { i := 0; return &i }(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rsvp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestRSVPAnswer_Validate(t *testing.T) {
	textAnswer := "Sample text answer"
	optionAnswer := "option1"
	boolAnswer := true

	tests := []struct {
		name    string
		answer  *RSVPAnswer
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text answer",
			answer: &RSVPAnswer{
				RSVPID:     1,
				QuestionID: 1,
				AnswerText: &textAnswer,
			},
			wantErr: false,
		},
		{
			name: "valid option answer",
			answer: &RSVPAnswer{
				RSVPID:       1,
				QuestionID:   1,
				AnswerOption: &optionAnswer,
			},
			wantErr: false,
		},
		{
			name: "valid boolean answer",
			answer: &RSVPAnswer{
				RSVPID:        1,
				QuestionID:    1,
				AnswerBoolean: &boolAnswer,
			},
			wantErr: false,
		},
		{
			name: "zero rsvp_id",
			answer: &RSVPAnswer{
				RSVPID:     0,
				QuestionID: 1,
				AnswerText: &textAnswer,
			},
			wantErr: true,
			errMsg:  "rsvp_id is required",
		},
		{
			name: "zero question_id",
			answer: &RSVPAnswer{
				RSVPID:     1,
				QuestionID: 0,
				AnswerText: &textAnswer,
			},
			wantErr: true,
			errMsg:  "question_id is required",
		},
		{
			name: "no answer fields populated",
			answer: &RSVPAnswer{
				RSVPID:     1,
				QuestionID: 1,
			},
			wantErr: true,
			errMsg:  "exactly one answer field must be populated",
		},
		{
			name: "multiple answer fields populated",
			answer: &RSVPAnswer{
				RSVPID:        1,
				QuestionID:    1,
				AnswerText:    &textAnswer,
				AnswerBoolean: &boolAnswer,
			},
			wantErr: true,
			errMsg:  "exactly one answer field must be populated",
		},
		{
			name: "text answer too long",
			answer: &RSVPAnswer{
				RSVPID:     1,
				QuestionID: 1,
				AnswerText: func() *string {
					s := strings.Repeat("a", 501)
					return &s
				}(),
			},
			wantErr: true,
			errMsg:  "answer_text cannot exceed 500 characters",
		},
		{
			name: "text answer exactly 500 chars",
			answer: &RSVPAnswer{
				RSVPID:     1,
				QuestionID: 1,
				AnswerText: func() *string {
					s := strings.Repeat("a", 500)
					return &s
				}(),
			},
			wantErr: false,
		},
		{
			name: "empty text answer",
			answer: &RSVPAnswer{
				RSVPID:     1,
				QuestionID: 1,
				AnswerText: func() *string {
					s := ""
					return &s
				}(),
			},
			wantErr: true,
			errMsg:  "answer_text cannot be empty",
		},
		{
			name: "empty option answer",
			answer: &RSVPAnswer{
				RSVPID:       1,
				QuestionID:   1,
				AnswerOption: func() *string {
					s := ""
					return &s
				}(),
			},
			wantErr: true,
			errMsg:  "answer_option cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.answer.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
