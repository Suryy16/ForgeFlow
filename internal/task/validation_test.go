
package task

import (
	"errors"
	"testing"
)

func TestValidateCreateTaskInput(t *testing.T) {
	tests := []struct {
		name          string
		input         CreateTaskRequest
		expectedError error
		expectedTitle string
		expectedPriority string
	}{
		{
			name: "valid input",
			input: CreateTaskRequest{
				Title: "  Implement API  ",
				Priority: "high",
			},
			expectedTitle: "Implement API",
			expectedPriority: "HIGH",
		},
		{
			name: "empty title",
			input: CreateTaskRequest{
				Title: "   ",
			},
			expectedError: ErrInvalidTitle,
		},
		{
			name: "title too long",
			input: CreateTaskRequest{
				Title: string(make([]byte, 256)),
			},
			expectedError: ErrInvalidTitle,
		},
		{
			name: "invalid priority",
			input: CreateTaskRequest{
				Title: "Test task",
				Priority: "CRITICAL",
			},
			expectedError: ErrInvalidPriority,
		},
		{
			name: "default priority",
			input: CreateTaskRequest{
				Title: "Test task",
			},
			expectedPriority: "MEDIUM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateCreateTaskInput(&tt.input)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.expectedError,
					err,
				)
			}

			if err == nil {
				if tt.input.Title != tt.expectedTitle {
					t.Errorf(
						"expected title %q, got %q",
						tt.expectedTitle,
						tt.input.Title,
					)
				}

				if tt.input.Priority != tt.expectedPriority {
					t.Errorf(
						"expected priority %q, got %q",
						tt.expectedPriority,
						tt.input.Priority,
					)
				}
			}
		})
	}
}
