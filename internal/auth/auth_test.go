package auth

import (
	"testing"
)

func TestBuildAndParseToken(t *testing.T) {
	var userID int64 = 42

	token, err := BuildJWTString(userID)
	if err != nil {
		t.Fatalf("BuildJWTString returned error: %v", err)
	}
	if token == "" {
		t.Fatal("BuildJWTString returned empty token")
	}
	id, err := GetUserID(token)
	if err != nil {
		t.Fatalf("GetUserID returned error: %v", err)
	}
	if id != userID {
		t.Errorf("GetUserID = %d, want %d", id, userID)
	}
}

func TestGetUserID_InvalidToken(t *testing.T) {
	id, err := GetUserID("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if id != 0 {
		t.Errorf("expected zero id for invalid token, got %d", id)
	}
}

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := CheckPassword(hash, "secret"); err != nil {
		t.Errorf("CheckPassword with correct password: %v", err)
	}
	if err := CheckPassword(hash, "wrong"); err == nil {
		t.Error("CheckPassword with wrong password: expected error, got nil")
	}
}
