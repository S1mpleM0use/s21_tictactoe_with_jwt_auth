package di

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"tictactoe/internal/app"
	"tictactoe/internal/domain/service"
	"tictactoe/internal/infrastructure/datasource/storage"
	"tictactoe/internal/infrastructure/datasource/storage/pgsql"
	"tictactoe/internal/infrastructure/datasource/user"
	"tictactoe/internal/web/handler"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		pgsql.NewConfig,
		pgsql.NewGameStorage,
		storage.NewRepository,
		user.NewUserRepository,
		service.NewGameService,
		app.NewUserService,
		app.NewAuthService,
		app.NewJWTProvider,
		app.NewAppService,
		handler.NewUserAuthenticator,
		handler.NewHandler,
		newHTTPServer,
		provideTransactor,
	),
	fx.Invoke(migrateSchema),
	fx.Invoke(func(*http.Server) {}),
)

func migrateSchema(lc fx.Lifecycle, pool *pgsql.GameStorage) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pgsql.Migrate(ctx, pool)
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
}

func provideTransactor(storage *pgsql.GameStorage) app.Transactor {
	return storage
}

func newHTTPServer(lc fx.Lifecycle, h *handler.Handler) *http.Server {
	mux := h.Routes()
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Printf("HTTP server listening on %s", srv.Addr)
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Printf("HTTP server failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down HTTP server gracefully...")
			return srv.Shutdown(ctx)
		},
	})
	return srv
}
