package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rishabh-sonic/orbit/internal/admin"
	"github.com/rishabh-sonic/orbit/internal/auth"
	"github.com/rishabh-sonic/orbit/internal/comment"
	"github.com/rishabh-sonic/orbit/internal/db"
	"github.com/rishabh-sonic/orbit/internal/email"
	"github.com/rishabh-sonic/orbit/internal/message"
	appMiddleware "github.com/rishabh-sonic/orbit/internal/middleware"
	"github.com/rishabh-sonic/orbit/internal/misc"
	"github.com/rishabh-sonic/orbit/internal/notification"
	"github.com/rishabh-sonic/orbit/internal/post"
	"github.com/rishabh-sonic/orbit/internal/push"
	"github.com/rishabh-sonic/orbit/internal/search"
	"github.com/rishabh-sonic/orbit/internal/upload"
	"github.com/rishabh-sonic/orbit/internal/user"
	"github.com/rishabh-sonic/orbit/pkg/config"
	pkgOpenSearch "github.com/rishabh-sonic/orbit/pkg/opensearch"
	pkgRabbitMQ "github.com/rishabh-sonic/orbit/pkg/rabbitmq"
	pkgRedis "github.com/rishabh-sonic/orbit/pkg/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	// --- Database ---
	sqlDB, err := sql.Open("pgx", cfg.DBURL)
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		slog.Error("ping db", "err", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	queries := db.New(sqlDB)

	// --- Redis ---
	rdb, err := pkgRedis.New(cfg)
	if err != nil {
		slog.Error("redis connect", "err", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	// --- RabbitMQ ---
	mqClient, err := pkgRabbitMQ.New(cfg)
	if err != nil {
		slog.Error("rabbitmq connect", "err", err)
		os.Exit(1)
	}
	defer mqClient.Close()
	slog.Info("rabbitmq connected")

	// --- OpenSearch ---
	osClient, err := pkgOpenSearch.New(cfg)
	if err != nil {
		slog.Error("opensearch connect", "err", err)
		os.Exit(1)
	}

	// --- Services ---
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.JWTResetSecret, cfg.JWTExpiration)
	emailSvc := email.New(cfg)
	pushSvc := push.NewService(queries, cfg)
	notifSvc := notification.NewService(queries, mqClient, emailSvc, pushSvc)
	uploadSvc, err := upload.NewService(cfg)
	if err != nil {
		slog.Error("upload service", "err", err)
		os.Exit(1)
	}
	searchSvc := search.NewService(osClient, queries, cfg)
	authSvc := auth.NewService(queries, jwtSvc, emailSvc, searchSvc)
	oauthSvc := auth.NewOAuthService(cfg, authSvc)
	userSvc := user.NewService(queries, rdb)
	postSvc := post.NewService(queries, searchSvc)
	commentSvc := comment.NewService(queries)
	messageSvc := message.NewService(queries)

	// Ensure OpenSearch indices
	if cfg.SearchEnabled {
		idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := searchSvc.EnsureIndices(idxCtx); err != nil {
			slog.Error("opensearch ensure indices", "err", err)
		}
		idxCancel()

		if cfg.SearchReindexOnStartup {
			go func() {
				slog.Info("reindexing all posts and users into OpenSearch…")
				if err := searchSvc.ReindexAll(context.Background()); err != nil {
					slog.Error("reindex", "err", err)
				} else {
					slog.Info("reindex complete")
				}
			}()
		}
	}

	// --- Handlers ---
	authH := auth.NewHandler(authSvc, oauthSvc, jwtSvc)
	userH := user.NewHandler(userSvc)
	postH := post.NewHandler(postSvc)
	subH := post.NewSubscriptionHandler(queries)
	commentH := comment.NewHandler(commentSvc)
	notifH := notification.NewHandler(notifSvc)
	messageH := message.NewHandler(messageSvc, queries)
	searchH := search.NewHandler(searchSvc)
	uploadH := upload.NewHandler(uploadSvc)
	pushH := push.NewHandler(pushSvc)
	adminH := admin.NewHandler(queries, postSvc, commentSvc)
	miscH := misc.NewHandler(queries, rdb, cfg)

	// --- Router ---
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)
	r.Use(appMiddleware.CORS([]string{cfg.WebsiteURL}))
	r.Use(appMiddleware.Authenticate(jwtSvc))

	// OAuth initiation – these live outside /api so nginx can proxy /oauth/ separately
	r.Get("/oauth/google", authH.OAuthGoogleRedirect)
	r.Get("/oauth/github", authH.OAuthGitHubRedirect)

	r.Route("/api", func(r chi.Router) {
		// Health / misc
		r.Get("/hello", miscH.Health)
		r.Get("/config", miscH.PublicConfig)
		r.Post("/online/heartbeat", miscH.Heartbeat)
		r.Get("/online/count", miscH.OnlineCount)

		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Get("/check", authH.Check)
			r.Post("/forgot/send", authH.ForgotSend)
			r.Post("/forgot/verify", authH.ForgotVerify)
			r.Post("/forgot/reset", authH.ForgotReset)
			r.Post("/google", authH.Google)
			r.Post("/github", authH.GitHub)
		})

		// Users
		r.Route("/users", func(r chi.Router) {
			r.Get("/{identifier}", userH.GetUser)
			r.Get("/{identifier}/following", userH.GetFollowing)
			r.Get("/{identifier}/followers", userH.GetFollowers)

			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.RequireAuth)
				r.Get("/me", userH.GetMe)
				r.Put("/me", userH.UpdateMe)
				r.Post("/me/avatar", userH.UpdateAvatar)
			})
		})

		// Posts
		r.Route("/posts", func(r chi.Router) {
			r.Get("/", postH.List)
			r.Get("/recent", postH.Recent)
			r.Get("/featured", postH.Featured)
			r.Get("/{id}", postH.GetByID)

			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.RequireAuth)
				r.Post("/", postH.Create)
				r.Put("/{id}", postH.Update)
				r.Delete("/{id}", postH.Delete)
				r.Post("/{id}/close", postH.Close)
				r.Post("/{id}/reopen", postH.Reopen)
			})

			// Comments on posts
			r.Get("/{postId}/comments", commentH.ListForPost)
			r.With(appMiddleware.RequireAuth).Post("/{postId}/comments", commentH.Create)
		})

		// Comments
		r.Route("/comments", func(r chi.Router) {
			r.With(appMiddleware.RequireAuth).Post("/{id}/replies", commentH.Reply)
			r.Get("/{id}/replies", commentH.ListReplies)
			r.With(appMiddleware.RequireAuth).Delete("/{id}", commentH.Delete)
		})

		// Subscriptions
		r.Route("/subscriptions", func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth)
			r.Post("/posts/{postId}", subH.Subscribe)
			r.Delete("/posts/{postId}", subH.Unsubscribe)
			r.Post("/users/{username}", userH.Follow)
			r.Delete("/users/{username}", userH.Unfollow)
		})

		// Notifications
		r.Route("/notifications", func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth)
			r.Get("/", notifH.List)
			r.Get("/unread-count", notifH.UnreadCount)
			r.Post("/read", notifH.MarkRead)
			r.Get("/prefs", notifH.GetPrefs)
			r.Post("/prefs", notifH.UpdatePrefs)
			r.Get("/email-prefs", notifH.GetEmailPrefs)
			r.Post("/email-prefs", notifH.UpdateEmailPrefs)
		})

		// Messages
		r.Route("/messages", func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth)
			r.Get("/conversations", messageH.ListConversations)
			r.Get("/conversations/{id}", messageH.GetConversation)
			r.Get("/conversations/{id}/messages", messageH.ListMessages)
			r.Post("/conversations/{id}/messages", messageH.SendMessage)
			r.Post("/conversations/{id}/read", messageH.MarkRead)
			r.Get("/unread-count", messageH.UnreadCount)
			r.Post("/", messageH.StartConversation)
		})

		// Search
		r.Route("/search", func(r chi.Router) {
			r.Get("/posts", searchH.SearchPosts)
			r.Get("/posts/title", searchH.SearchPostTitles)
			r.Get("/posts/content", searchH.SearchPostContent)
			r.Get("/users", searchH.SearchUsers)
			r.Get("/global", searchH.SearchGlobal)
		})

		// Upload
		r.Route("/upload", func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth)
			r.Post("/", uploadH.Upload)
			r.Post("/url", uploadH.UploadFromURL)
			r.Get("/presign", uploadH.Presign)
		})

		// Push
		r.Route("/push", func(r chi.Router) {
			r.Get("/public-key", pushH.PublicKey)
			r.With(appMiddleware.RequireAuth).Post("/subscribe", pushH.Subscribe)
		})

		// Admin
		r.Route("/admin", func(r chi.Router) {
			r.Use(appMiddleware.RequireAdmin)
			r.Get("/config", adminH.GetConfig)
			r.Post("/config", adminH.UpdateConfig)
			r.Get("/users", adminH.ListUsers)
			r.Post("/users/{id}/ban", adminH.BanUser)
			r.Post("/users/{id}/unban", adminH.UnbanUser)
			r.Delete("/posts/{id}", adminH.DeletePost)
			r.Post("/posts/{id}/pin", adminH.PinPost)
			r.Post("/posts/{id}/unpin", adminH.UnpinPost)
			r.Delete("/comments/{id}", adminH.DeleteComment)
			r.Post("/comments/{id}/pin", adminH.PinComment)
			r.Post("/comments/{id}/unpin", adminH.UnpinComment)
			r.Get("/stats/dau", adminH.StatDAU)
			r.Get("/stats/dau-range", adminH.StatDAURange)
			r.Get("/stats/new-users-range", adminH.StatNewUsers)
			r.Get("/stats/posts-range", adminH.StatPosts)
		})
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("API server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
