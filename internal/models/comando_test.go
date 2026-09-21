package models

import "testing"

func TestCommandIsValid(t *testing.T) {
	tests := []struct {
		name    string
		command *int
		want    bool
	}{
		{name: "missing", command: nil, want: false},
		{name: "off", command: intPointer(0), want: true},
		{name: "on", command: intPointer(1), want: true},
		{name: "negative", command: intPointer(-1), want: false},
		{name: "above allowed range", command: intPointer(2), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Command{Command: tt.command}).IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func intPointer(value int) *int {
	return &value
}
