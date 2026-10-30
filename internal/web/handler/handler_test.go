package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tictactoe/internal/app"
	domainModel "tictactoe/internal/domain/model"
	webModel "tictactoe/internal/web/model"

	"github.com/google/uuid"
)

// --- Mocks for Web Layer ---

type stubAppService struct {
	createFn      func(ctx context.Context, hostID uuid.UUID, mode string) (domainModel.Game, error)
	joinFn        func(ctx context.Context, gameID, userID uuid.UUID) (domainModel.Game, error)
	moveFn        func(ctx context.Context, gameID, userID uuid.UUID, row, col int) (domainModel.Game, error)
	waitingFn     func(ctx context.Context) ([]domainModel.Game, error)
	finishedFn    func(ctx context.Context, userID uuid.UUID) ([]domainModel.Game, error)
	currentFn     func(ctx context.Context, userID uuid.UUID) (domainModel.Game, error)
	leaderboardFn func(ctx context.Context, limit int) ([]domainModel.LeaderboardEntry, error)
}

func (s *stubAppService) CreateGame(ctx context.Context, hostID uuid.UUID, mode string) (domainModel.Game, error) {
	return s.createFn(ctx, hostID, mode)
}
func (s *stubAppService) JoinGame(ctx context.Context, gameID, userID uuid.UUID) (domainModel.Game, error) {
	return s.joinFn(ctx, gameID, userID)
}
func (s *stubAppService) ApplyMove(ctx context.Context, gameID, userID uuid.UUID, row, col int) (domainModel.Game, error) {
	return s.moveFn(ctx, gameID, userID, row, col)
}
func (s *stubAppService) GetWaitingGames(ctx context.Context) ([]domainModel.Game, error) {
	return s.waitingFn(ctx)
}
func (s *stubAppService) GetFinishedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]domainModel.Game, error) {
	return s.finishedFn(ctx, userID)
}
func (s *stubAppService) GetCurrentGame(ctx context.Context, userID uuid.UUID) (domainModel.Game, error) {
	return s.currentFn(ctx, userID)
}
func (s *stubAppService) GetLeaderboard(ctx context.Context, limit int) ([]domainModel.LeaderboardEntry, error) {
	return s.leaderboardFn(ctx, limit)
}

type stubAuthService struct {
	registerFn      func(ctx context.Context, req domainModel.SignUpRequest) (domainModel.User, error)
	authenticateFn  func(ctx context.Context, req domainModel.JwtRequest) (domainModel.JwtResponse, error)
	refreshAccessFn func(ctx context.Context, token string) (domainModel.JwtResponse, error)
	refreshTokenFn  func(ctx context.Context, token string) (domainModel.JwtResponse, error)
}

func (s *stubAuthService) Register(ctx context.Context, req domainModel.SignUpRequest) (domainModel.User, error) {
	return s.registerFn(ctx, req)
}
func (s *stubAuthService) Authenticate(ctx context.Context, req domainModel.JwtRequest) (domainModel.JwtResponse, error) {
	return s.authenticateFn(ctx, req)
}
func (s *stubAuthService) RefreshAccessToken(ctx context.Context, token string) (domainModel.JwtResponse, error) {
	return s.refreshAccessFn(ctx, token)
}
func (s *stubAuthService) RefreshRefreshToken(ctx context.Context, token string) (domainModel.JwtResponse, error) {
	return s.refreshTokenFn(ctx, token)
}

type stubUserService struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (domainModel.User, error)
}

func (s *stubUserService) Create(context.Context, domainModel.SignUpRequest) (domainModel.User, error) {
	return domainModel.User{}, nil
}
func (s *stubUserService) VerifyCredentials(context.Context, string, string) (domainModel.User, error) {
	return domainModel.User{}, nil
}
func (s *stubUserService) GetByID(ctx context.Context, id uuid.UUID) (domainModel.User, error) {
	return s.getByIDFn(ctx, id)
}

func setupTestMux(appSvc app.AppService, authSvc app.AuthService, userSvc app.UserService) (*http.ServeMux, app.JWTProvider) {
	jwtProv := app.NewJWTProvider()
	authn := NewUserAuthenticator(jwtProv)
	h := NewHandler(appSvc, authSvc, userSvc, authn)
	return h.Routes(), jwtProv
}

func TestHandler_AuthAndMiddleware(t *testing.T) {
	testUser := domainModel.User{ID: uuid.New(), Login: "nikolai"}

	authSvc := &stubAuthService{
		registerFn: func(_ context.Context, req domainModel.SignUpRequest) (domainModel.User, error) {
			if req.Login == "taken" {
				return domainModel.User{}, app.ErrLoginAlreadyTaken
			}
			return testUser, nil
		},
		authenticateFn: func(_ context.Context, req domainModel.JwtRequest) (domainModel.JwtResponse, error) {
			if req.Password != "secret" {
				return domainModel.JwtResponse{}, app.ErrInvalidCredentials
			}
			return domainModel.JwtResponse{Type: "Bearer", AccessToken: "acc", RefreshToken: "ref"}, nil
		},
	}
	userSvc := &stubUserService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (domainModel.User, error) {
			if id == testUser.ID {
				return testUser, nil
			}
			return domainModel.User{}, app.ErrUserNotFound
		},
	}

	mux, jwtProv := setupTestMux(&stubAppService{}, authSvc, userSvc)
	validToken, _ := jwtProv.GenerateAccessToken(testUser)
	refreshToken, _ := jwtProv.GenerateRefreshToken(testUser)

	t.Run("POST /user returns 201 on valid signup and 409 on duplicate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(`{"login":"nikolai","password":"secret"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		reqDup := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(`{"login":"taken","password":"secret"}`))
		recDup := httptest.NewRecorder()
		mux.ServeHTTP(recDup, reqDup)
		if recDup.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", recDup.Code)
		}
	})

	t.Run("POST /user rejects unknown JSON fields with 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(`{"login":"a","password":"b","extra":1}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 on unknown field, got %d", rec.Code)
		}
	})

	t.Run("POST /user/login returns 401 with WWW-Authenticate on wrong password", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString(`{"login":"nikolai","password":"wrong"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		if rec.Header().Get("WWW-Authenticate") == "" {
			t.Fatal("expected WWW-Authenticate header on 401")
		}
	})

	t.Run("Middleware rejects missing token, refresh token in place of access, and accepts valid access token", func(t *testing.T) {
		// Без заголовка
		reqNoAuth := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		recNoAuth := httptest.NewRecorder()
		mux.ServeHTTP(recNoAuth, reqNoAuth)
		if recNoAuth.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without header, got %d", recNoAuth.Code)
		}

		// С refresh-токеном вместо access-токена
		reqRef := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		reqRef.Header.Set("Authorization", "Bearer "+refreshToken)
		recRef := httptest.NewRecorder()
		mux.ServeHTTP(recRef, reqRef)
		if recRef.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 when using refresh token as access token, got %d", recRef.Code)
		}

		// С валидным access-токеном
		reqValid := httptest.NewRequest(http.MethodGet, "/user/me", nil)
		reqValid.Header.Set("Authorization", "Bearer "+validToken)
		recValid := httptest.NewRecorder()
		mux.ServeHTTP(recValid, reqValid)
		if recValid.Code != http.StatusOK {
			t.Fatalf("expected 200 with valid token, got %d: %s", recValid.Code, recValid.Body.String())
		}
	})
}

func TestHandler_GameAndLeaderboardEndpoints(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	var capturedLimit int

	appSvc := &stubAppService{
		createFn: func(_ context.Context, hostID uuid.UUID, mode string) (domainModel.Game, error) {
			if mode != domainModel.ModeVSComputer && mode != domainModel.ModeVSPlayer {
				return domainModel.Game{}, app.ErrInvalidMode
			}
			return domainModel.Game{
				ID:        gameID,
				Board:     domainModel.Board{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
				GameState: domainModel.StatePlaying,
				Mode:      mode,
				HostID:    hostID,
				CreatedAt: now,
			}, nil
		},
		moveFn: func(_ context.Context, gID, uID uuid.UUID, row, col int) (domainModel.Game, error) {
			if row == 1 && col == 1 {
				return domainModel.Game{}, app.ErrCellOccupied
			}
			return domainModel.Game{
				ID:        gID,
				Board:     domainModel.Board{{1, 0, 0}, {0, 2, 0}, {0, 0, 0}},
				GameState: domainModel.StatePlaying,
				Mode:      domainModel.ModeVSComputer,
				HostID:    uID,
				CreatedAt: now,
			}, nil
		},
		leaderboardFn: func(_ context.Context, limit int) ([]domainModel.LeaderboardEntry, error) {
			capturedLimit = limit
			return []domainModel.LeaderboardEntry{
				{UserID: userID, Login: "nikolai", Wins: 5, Losses: 1, Draws: 0, WinRatio: 0.8333},
			}, nil
		},
	}

	mux, jwtProv := setupTestMux(appSvc, &stubAuthService{}, &stubUserService{})
	token, _ := jwtProv.GenerateAccessToken(domainModel.User{ID: userID, Login: "nikolai"})

	t.Run("POST /game returns 201 and non-zero created_at", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/game", bytes.NewBufferString(`{"mode":"VScomputer"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp webModel.GameResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if !resp.CreatedAt.Equal(now) {
			t.Fatalf("expected CreatedAt %v, got %v", now, resp.CreatedAt)
		}
	})

	t.Run("POST /game/{id}/move rejects missing row or column pointer fields with 400", func(t *testing.T) {
		// Только row без column
		req := httptest.NewRequest(http.MethodPost, "/game/"+gameID.String()+"/move", bytes.NewBufferString(`{"row":0}`))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 when column is missing, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /game/{id}/move returns 409 on occupied cell and 200 on valid move", func(t *testing.T) {
		reqConflict := httptest.NewRequest(http.MethodPost, "/game/"+gameID.String()+"/move", bytes.NewBufferString(`{"row":1,"column":1}`))
		reqConflict.Header.Set("Authorization", "Bearer "+token)
		recConflict := httptest.NewRecorder()
		mux.ServeHTTP(recConflict, reqConflict)
		if recConflict.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", recConflict.Code)
		}

		reqOK := httptest.NewRequest(http.MethodPost, "/game/"+gameID.String()+"/move", bytes.NewBufferString(`{"row":0,"column":0}`))
		reqOK.Header.Set("Authorization", "Bearer "+token)
		recOK := httptest.NewRecorder()
		mux.ServeHTTP(recOK, reqOK)
		if recOK.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", recOK.Code, recOK.Body.String())
		}
	})

	t.Run("GET /game/leaderboard supports default, query param, path param and rejects invalid limit", func(t *testing.T) {
		// По умолчанию limit=10
		reqDef := httptest.NewRequest(http.MethodGet, "/game/leaderboard", nil)
		reqDef.Header.Set("Authorization", "Bearer "+token)
		recDef := httptest.NewRecorder()
		mux.ServeHTTP(recDef, reqDef)
		if recDef.Code != http.StatusOK || capturedLimit != 10 {
			t.Fatalf("expected 200 and limit=10, got code=%d limit=%d", recDef.Code, capturedLimit)
		}

		// Через query-параметр ?limit=5
		reqQuery := httptest.NewRequest(http.MethodGet, "/game/leaderboard?limit=5", nil)
		reqQuery.Header.Set("Authorization", "Bearer "+token)
		recQuery := httptest.NewRecorder()
		mux.ServeHTTP(recQuery, reqQuery)
		if recQuery.Code != http.StatusOK || capturedLimit != 5 {
			t.Fatalf("expected 200 and limit=5, got code=%d limit=%d", recQuery.Code, capturedLimit)
		}

		// Через path-параметр /game/leaderboard/3
		reqPath := httptest.NewRequest(http.MethodGet, "/game/leaderboard/3", nil)
		reqPath.Header.Set("Authorization", "Bearer "+token)
		recPath := httptest.NewRecorder()
		mux.ServeHTTP(recPath, reqPath)
		if recPath.Code != http.StatusOK || capturedLimit != 3 {
			t.Fatalf("expected 200 and limit=3, got code=%d limit=%d", recPath.Code, capturedLimit)
		}

		// Невалидный limit=0 -> 400 Bad Request
		reqBad := httptest.NewRequest(http.MethodGet, "/game/leaderboard?limit=0", nil)
		reqBad.Header.Set("Authorization", "Bearer "+token)
		recBad := httptest.NewRecorder()
		mux.ServeHTTP(recBad, reqBad)
		if recBad.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 on limit=0, got %d", recBad.Code)
		}
	})
}
