package auth

import "testing"

func TestPasswordHash(t *testing.T) {
	test_password := "pa$$word"

	hashed, err := HashPassword(test_password)

	if err != nil {
		t.Errorf("Error generating password hash")
	}

	if len(hashed) == 0 {
		t.Errorf("Error generating password hash")
	}
}
