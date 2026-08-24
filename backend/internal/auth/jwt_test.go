package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueContainsUserNotRole(t *testing.T) {
	iss := &Issuer{Secret: []byte("dev-secret"), AccessTTL: 2 * time.Hour, RefreshTTL: 7 * 24 * time.Hour}
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tok, err := iss.Issue(id, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	c, err := iss.Parse(tok.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if c.UserID != id || c.Kind != "access" {
		t.Fatalf("claims %+v", c)
	}
	if _, ok := any(c).(interface{ Role() string }); ok {
		t.Fatal("token must not carry role")
	}
}
