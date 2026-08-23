package auth_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gopuppy/internal/auth"
)

func TestIssuerConcurrentIssueKeepsClaimsIsolated(t *testing.T) {
	issuer := &auth.Issuer{
		Secret:     []byte("concurrent-login-test-secret"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
	}
	now := time.Now().Truncate(time.Second)
	const callers = 64

	start := make(chan struct{})
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for n := 0; n < callers; n++ {
		userID := uuid.New()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			tokens, err := issuer.Issue(userID, now)
			if err != nil {
				errs <- fmt.Errorf("issue token for %s: %w", userID, err)
				return
			}
			for token, wantKind := range map[string]string{
				tokens.AccessToken:  "access",
				tokens.RefreshToken: "refresh",
			} {
				claims, err := issuer.Parse(token)
				if err != nil {
					errs <- fmt.Errorf("parse %s token for %s: %w", wantKind, userID, err)
					return
				}
				if claims.UserID != userID || claims.Kind != wantKind {
					errs <- fmt.Errorf("%s token for %s contains uid=%s kind=%q", wantKind, userID, claims.UserID, claims.Kind)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
