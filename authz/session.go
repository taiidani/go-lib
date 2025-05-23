package authz

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/taiidani/go-lib/cache"
)

type Session struct {
	Name   string
	Secure bool
	TTL    time.Duration
	Path   string

	backend cache.Cache
}

func NewSession(ctx context.Context, backend cache.Cache) Session {
	return Session{
		Name:    "session",
		Secure:  true,
		TTL:     time.Duration(time.Hour * 168),
		Path:    "/",
		backend: backend,
	}
}

func (s *Session) Create(ctx context.Context, sess any) (*http.Cookie, error) {
	sessionKey := uuid.New().String()
	err := s.backend.Set(ctx, "session:"+sessionKey, sess, s.TTL)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:     s.Name,
		Value:    sessionKey,
		Secure:   s.Secure,
		Path:     s.Path,
		HttpOnly: true,
		MaxAge:   int(s.TTL.Seconds()),
	}
	return &cookie, nil
}

func (s *Session) Delete() *http.Cookie {
	cookie := http.Cookie{
		Name:     s.Name,
		Value:    "",
		Secure:   s.Secure,
		Path:     s.Path,
		HttpOnly: true,
		MaxAge:   -1,
	}
	return &cookie
}

func (s *Session) Get(r *http.Request, sess any) error {
	cookie, err := r.Cookie(s.Name)
	if err != nil {
		// No cookie 🍪
		return nil
	}

	err = s.backend.Get(r.Context(), "session:"+cookie.Value, &sess)
	if err != nil {
		return fmt.Errorf("failed to load session from backend: %w", err)
	}

	return nil
}

func (s *Session) Update(r *http.Request, sess any) error {
	cookie, err := r.Cookie(s.Name)
	if err != nil {
		// No cookie 🍪
		return fmt.Errorf("no session found to update")
	}

	err = s.backend.Set(r.Context(), "session:"+cookie.Value, &sess, s.TTL)
	if err != nil {
		return fmt.Errorf("failed to update session in backend: %w", err)
	}

	return nil
}
