package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestPasswordHash(t *testing.T) {
	test_password := "pa$$word"
	test_other_password := "letmein1"

	hashed1, _ := HashPassword(test_password)
	hashed2, _ := HashPassword(test_other_password)


	if len(hashed1) == 0 {
		t.Errorf("Error generating password hash")
	}

	tests := []struct {
		name 		string
		password string
		hash  string
		wantErr bool
		matchPassword bool
	}{
		{
			name: "Correct password",
			password: test_password,
			hash: hashed1,
			wantErr: false,
			matchPassword: true,
		},
		{
			name: "Incorrect password",
			password: test_password,
			hash: hashed2,
			wantErr: false,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswwordHash: error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash: expects %v, got %v", tt.matchPassword, match)
			}
		})
	}

}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name string
		tokenString string
		tokenSecret string
		wantUserID uuid.UUID
		wantErr bool
	}{
		{
			name: "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID: userID,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT: error = %v, wantErr %v",err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT: gotUserID = %v, want %v",gotUserID,tt.wantUserID)
			}
		})
	}


}
