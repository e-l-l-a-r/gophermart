package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrder_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid luhn 1",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "valid luhn 2",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "invalid luhn",
			number: "12345678904",
			want:   false,
		},
		{
			name:   "invalid characters",
			number: "1234567890a",
			want:   false,
		},
		{
			name:   "valid with spaces",
			number: "799 273 987 13",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				Number: tt.number,
			}
			assert.Equal(t, tt.want, o.IsValid())
		})
	}
}
