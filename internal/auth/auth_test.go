package auth

import (
	"testing"
	"fmt"
	"time"
	"encoding/hex"
	"net/http/httptest"
	"net/http"
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


func TestGetBearerToken(t *testing.T) {
	bearerToken := uuid.New()
	tokenHeader := fmt.Sprintf("Bearer %s",bearerToken)
	returnToken := bearerToken.String()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization",tokenHeader)

	reqFalse := httptest.NewRequest("GET","/fail",nil)

	tests := []struct{
		name string
		headers http.Header
		wantValid bool
		wantToken string
		wantErr bool
	}{
		{
			name: "No header",
			headers: reqFalse.Header,
			wantToken: "",
			wantErr: true,
		},
		{
			name: "Has header with prefix",
			headers: req.Header,
			wantToken: returnToken,
			wantErr: false,
		},
		{
			name: "Has header",
			headers: req.Header,
			wantToken: returnToken,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			gotBearerToken, err := GetBearerToken(tt.headers)

			t.Logf("Err: %s, WantErr: %v",err,tt.wantErr)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken: error = %v, wantErr = %v",err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if gotBearerToken != tt.wantToken {
				t.Errorf("GetBearerToken: token = %s, wantToken = %s, wantValid = %v",gotBearerToken,tt.wantToken,tt.wantValid)
			}
		})

	}
}

func TestMakeRefreshToken(t *testing.T) {
	refreshToken := MakeRefreshToken()

	t.Logf("Token Length: %d",len(refreshToken))
	if len(refreshToken) == 0 {
		t.Errorf("MakeRefreshToken: Generated empty refresh token")
		return
	}

	decoded,err := hex.DecodeString(refreshToken)
	if err != nil {
		t.Errorf("MakeRefreshToken: Decoding refresh token failed")
		return
	}
	if len(decoded) != 32 {
		t.Errorf("MakeRefreshToken: Generated refresh token of wrong length")
		return
	}
}
