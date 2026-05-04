package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"marketing-revenue-analytics/auth"
	"marketing-revenue-analytics/cache"
	"marketing-revenue-analytics/internal/clients"
	"marketing-revenue-analytics/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type App struct {
	server  *http.Server
	pool    *pgxpool.Pool
	cache   *cache.Cache
	logger  *zap.Logger
	clients *clients.Clients
}

type Config struct {
	Port         string
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

// New wires every dependency and returns a ready-to-start App.
// Order: logger → db → cache → jwt → handlers → router → server
func New(
	cfg Config,
	pool *pgxpool.Pool,
	queries *models.Queries,
	jwtManager *auth.JWTManager,
	cacheClient *cache.Cache,
	logger *zap.Logger,
	clients *clients.Clients,

) (*App, error) {

	if cfg.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.UseH2C = true

	if err := router.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	// Global middleware
	router.Use(
		requestIdInterceptor(),
		ginLogger(),
		ginRecovery(),
	)

	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"status": "failure", "message": "route not found"})
	})
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": "failure", "message": "method not allowed"})
	})

	// Wire all routes
	bindRoutes(router, queries, jwtManager, cacheClient, clients, logger)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:           cfg.Port,
		Handler:        h2c.NewHandler(router, h2s),
		ReadTimeout:    time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 10,
	}

	return &App{
		server: srv,
		pool:   pool,
		cache:  cacheClient,
		logger: logger,
	}, nil
}

// Start blocks until the server stops.
func (a *App) Start() error {
	a.logger.Info("server starting", zap.String("addr", a.server.Addr))
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown drains connections then closes dependencies in order.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down server gracefully...")

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	// Close dependencies after HTTP drains
	a.cache.Close()
	a.pool.Close()
	a.logger.Info("server exited cleanly")
	return nil
}
