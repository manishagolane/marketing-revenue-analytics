package server

import (
	"net/http"

	"marketing-revenue-analytics/auth"
	"marketing-revenue-analytics/cache"
	"marketing-revenue-analytics/handlers"
	"marketing-revenue-analytics/internal/clients"
	"marketing-revenue-analytics/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JWT = stateless identity
// Redis = stateful session control
// Token expiry = 24 hours
// Logout = delete from Redis

func bindRoutes(
	router *gin.Engine,
	queries *models.Queries,
	jwtManager *auth.JWTManager,
	cacheClient *cache.Cache,
	clients *clients.Clients,
	logger *zap.Logger,
) {
	// Health check — no auth, no rate limit
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Handlers
	authHandler := handlers.NewAuthHandler(queries, jwtManager, cacheClient)
	campaignHandler := handlers.NewCampaignHandler(queries, logger)
	eventHandler := handlers.NewEventHandler(queries, clients.SqsClient, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(queries, cacheClient)
	engagementHandler := handlers.NewEngagementHandler(queries)

	// Middleware
	authenticate := authMiddleware(cacheClient, jwtManager)
	anonRL := rateLimitMiddleware(cacheClient, 30) // 30 req/min per IP
	authRL := rateLimitMiddleware(cacheClient, 60) // 60 req/min per user

	api := router.Group("/api/v1")

	// ── Public: auth endpoints
	public := api.Group("/auth")
	public.Use(anonRL)
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// ── Public: campaign preview — no JWT
	api.GET("/campaigns/public", anonRL, campaignHandler.ListPublic)
	api.GET("/campaigns/:id/preview", anonRL, campaignHandler.GetPublicPreview)
	api.POST("/events/track", rateLimitMiddleware(cacheClient, 60), eventHandler.Track)
	api.GET("/analytics/campaigns/:id/summary", anonRL, analyticsHandler.PublicSummary)

	// ── Protected: JWT required
	protected := api.Group("/")
	protected.Use(authenticate, authRL)
	{
		// Profile — any authenticated role
		profile := protected.Group("/auth")
		{
			profile.GET("/profile",
				rbacMiddleware(queries, cacheClient, "profile", "read"),
				authHandler.GetProfile,
			)
			profile.PATCH("/profile",
				rbacMiddleware(queries, cacheClient, "profile", "update"),
				authHandler.UpdateProfile,
			)
			profile.POST("/logout", authHandler.Logout)
		}

		// Campaigns — Part 2 (wire handler here when ready)
		campaigns := protected.Group("/campaigns")
		{
			campaigns.POST("",
				rbacMiddleware(queries, cacheClient, "campaign", "create"),
				campaignHandler.Create,
			)
			campaigns.GET("",
				rbacMiddleware(queries, cacheClient, "campaign", "read"),
				campaignHandler.List,
			)
			campaigns.GET("/search",
				rbacMiddleware(queries, cacheClient, "campaign", "read"),
				campaignHandler.Search,
			)
			campaigns.GET("/:id",
				rbacMiddleware(queries, cacheClient, "campaign", "read"),
				campaignHandler.Get,
			)
			campaigns.PATCH("/:id",
				rbacMiddleware(queries, cacheClient, "campaign", "update"),
				campaignHandler.Update,
			)
			campaigns.PATCH("/:id/status",
				rbacMiddleware(queries, cacheClient, "campaign", "update"),
				campaignHandler.UpdateStatus,
			)
			campaigns.DELETE("/:id",
				rbacMiddleware(queries, cacheClient, "campaign", "delete"),
				campaignHandler.Delete,
			)
		}

		analytics := protected.Group("/analytics")
		analytics.Use(rbacMiddleware(queries, cacheClient, "stats", "read"))
		{
			analytics.GET("/daily", analyticsHandler.Daily)
			analytics.GET("/weekly", analyticsHandler.Weekly)
			analytics.GET("/monthly", analyticsHandler.Monthly)
		}

		engagement := protected.Group("/engagement")
		engagement.Use(rbacMiddleware(queries, cacheClient, "stats", "read"))
		{
			engagement.GET("/:campaign_id/funnel", engagementHandler.Funnel)
			engagement.GET("/:campaign_id/time-spent", engagementHandler.TimeSpent)
			engagement.GET("/:campaign_id/click-path", engagementHandler.ClickPath)
		}
	}

}
