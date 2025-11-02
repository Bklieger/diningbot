package utils

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "specific date",
			time: time.Date(2024, 11, 4, 0, 0, 0, 0, time.UTC),
			want: "11/4/2024",
		},
		{
			name: "single digit month and day",
			time: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			want: "1/5/2024",
		},
		{
			name: "today",
			time: time.Now(),
			want: FormatDate(time.Now()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.time)
			if got != tt.want {
				t.Errorf("FormatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
