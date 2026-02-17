package domain

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name    string
		pwd     string
		wantErr bool
	}{
		{"too short", "Ab1!", true},
		{"no upper", "password1!", true},
		{"no lower", "PASSWORD1!", true},
		{"no digit", "Password!", true},
		{"no special", "Password1", true},
		{"valid", "Password1!", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.pwd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
