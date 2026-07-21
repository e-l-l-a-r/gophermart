package crypto

import (
	"testing"
)

func TestGetAuthKey(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		want     string
	}{
		{
			name:     "success",
			login:    "user1",
			password: "password123",
			want:     "d1e860468869fb5948069a59ba6cf803a000a2ce7e74e20532ff92bf32d9974d",
		},
		{
			name:     "different credentials",
			login:    "user2",
			password: "password123",
			want:     "8795dc649bf758a8526a7e66838fe5fd10fa7fef088b699e149611e08756deca",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAuthKey(tt.login, tt.password)
			if got == "" {
				t.Errorf("GetAuthKey() returned empty string")
			}
			if tt.want != "" && got != tt.want {
				t.Errorf("GetAuthKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
