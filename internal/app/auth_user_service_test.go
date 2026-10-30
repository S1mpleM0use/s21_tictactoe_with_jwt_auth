package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	domainModel "tictactoe/internal/domain/model"
	"tictactoe/internal/infrastructure/datasource/user"

	"github.com/google/uuid"
)

type mockUserRepo struct {
	usersByID    map[uuid.UUID]domainModel.User
	usersByLogin map[string]domainModel.User
	saveErr      error
	getErr       error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:    make(map[uuid.UUID]domainModel.User),
		usersByLogin: make(map[string]domainModel.User),
	}
}

func (m *mockUserRepo) SaveUser(_ context.Context, u domainModel.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if _, exists := m.usersByLogin[u.Login]; exists {
		return user.ErrLoginAlreadyTaken
	}
	m.usersByID[u.ID] = u
	m.usersByLogin[u.Login] = u
	return nil
}

func (m *mockUserRepo) GetUserByLogin(_ context.Context, login string) (domainModel.User, error) {
	if m.getErr != nil {
		return domainModel.User{}, m.getErr
	}
	u, ok := m.usersByLogin[login]
	if !ok {
		return domainModel.User{}, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByID(_ context.Context, id uuid.UUID) (domainModel.User, error) {
	if m.getErr != nil {
		return domainModel.User{}, m.getErr
	}
	u, ok := m.usersByID[id]
	if !ok {
		return domainModel.User{}, user.ErrUserNotFound
	}
	return u, nil
}

func TestAuthAndUserService_RegisterErrors(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	userSvc := NewUserService(repo)
	jwtProv := NewJWTProvider()
	authSvc := NewAuthService(userSvc, jwtProv)

	tests := []struct {
		name    string
		req     domainModel.SignUpRequest
		wantErr error
	}{
		{
			name:    "empty login",
			req:     domainModel.SignUpRequest{Login: "   ", Password: "password123"},
			wantErr: ErrLoginRequired,
		},
		{
			name:    "login too long (>20 runes)",
			req:     domainModel.SignUpRequest{Login: strings.Repeat("a", 21), Password: "password123"},
			wantErr: ErrLoginTooLong,
		},
		{
			name:    "password too short (<4 chars)",
			req:     domainModel.SignUpRequest{Login: "nikolai", Password: "123"},
			wantErr: ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authSvc.Register(ctx, tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}

	t.Run("successful registration and duplicate login check", func(t *testing.T) {
		req := domainModel.SignUpRequest{Login: "player1", Password: "secret"}
		created, err := authSvc.Register(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ID == uuid.Nil || created.Login != "player1" {
			t.Fatalf("unexpected created user: %+v", created)
		}

		_, err = authSvc.Register(ctx, req)
		if !errors.Is(err, ErrLoginAlreadyTaken) {
			t.Fatalf("expected ErrLoginAlreadyTaken, got %v", err)
		}
	})
}

func TestAuthService_AuthenticateAndTokens(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	userSvc := NewUserService(repo)
	jwtProv := NewJWTProvider()
	authSvc := NewAuthService(userSvc, jwtProv)

	_, err := authSvc.Register(ctx, domainModel.SignUpRequest{
		Login:    "valid_user",
		Password: "correct_password",
	})
	if err != nil {
		t.Fatalf("setup register failed: %v", err)
	}

	t.Run("non-existent user returns ErrInvalidCredentials", func(t *testing.T) {
		_, err := authSvc.Authenticate(ctx, domainModel.JwtRequest{Login: "ghost", Password: "password123"})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong password returns ErrInvalidCredentials", func(t *testing.T) {
		_, err := authSvc.Authenticate(ctx, domainModel.JwtRequest{Login: "valid_user", Password: "wrong"})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	var jwtResp domainModel.JwtResponse
	t.Run("successful authentication returns jwt response", func(t *testing.T) {
		var err error
		jwtResp, err = authSvc.Authenticate(ctx, domainModel.JwtRequest{Login: "valid_user", Password: "correct_password"})
		if err != nil {
			t.Fatalf("unexpected auth error: %v", err)
		}
		if jwtResp.AccessToken == "" || jwtResp.RefreshToken == "" || jwtResp.Type != "Bearer" {
			t.Fatalf("invalid jwt response: %+v", jwtResp)
		}
	})

	t.Run("refresh access token success", func(t *testing.T) {
		refResp, err := authSvc.RefreshAccessToken(ctx, jwtResp.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected refresh access error: %v", err)
		}
		if refResp.AccessToken == "" {
			t.Fatal("expected non-empty new access token")
		}
	})

	t.Run("refresh refresh token success", func(t *testing.T) {
		refResp, err := authSvc.RefreshRefreshToken(ctx, jwtResp.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected refresh tokens error: %v", err)
		}
		if refResp.AccessToken == "" || refResp.RefreshToken == "" {
			t.Fatal("expected new token pair")
		}
	})

	t.Run("invalid refresh token fails", func(t *testing.T) {
		_, err := authSvc.RefreshAccessToken(ctx, "invalid.token.string")
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestUserService_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	userSvc := NewUserService(repo)

	created, err := userSvc.Create(ctx, domainModel.SignUpRequest{
		Login:    "lookup_user",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	t.Run("existing user", func(t *testing.T) {
		found, err := userSvc.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found.ID != created.ID || found.Login != "lookup_user" {
			t.Fatalf("unexpected user: %+v", found)
		}
	})

	t.Run("non-existent user returns ErrUserNotFound", func(t *testing.T) {
		_, err := userSvc.GetByID(ctx, uuid.New())
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})
}
