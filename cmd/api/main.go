package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pemira-api/internal/auth"
	"pemira-api/internal/candidate"
	"pemira-api/internal/config"
	"pemira-api/internal/dpt"
	"pemira-api/internal/election"
	httpMiddleware "pemira-api/internal/http/middleware"
	"pemira-api/internal/http/response"
	"pemira-api/internal/tps"
	"pemira-api/internal/voting"
	"pemira-api/internal/ws"
	"pemira-api/pkg/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to database")

	// Initialize repositories
	authRepo := auth.NewPgRepository(pool)
	electionRepo := election.NewRepository(pool)
	electionAdminRepo := election.NewPgAdminRepository(pool)
	dptRepo := dpt.NewRepository(pool)
	tpsAdminRepo := tps.NewPgAdminRepository(pool)
	candidatePgRepo := candidate.NewPgCandidateRepository(pool)
	candidateStatsProvider := candidate.NewPgStatsProvider(pool)

	voterRepo := voting.NewVoterRepository()
	candidateRepo := voting.NewCandidateRepository()
	voteRepo := voting.NewVoteRepository()
	statsRepo := voting.NewVoteStatsRepository()
	auditSvc := voting.NewAuditService()

	// Initialize services
	jwtConfig := auth.JWTConfig{
		Secret:          cfg.JWTSecret,
		AccessTokenTTL:  24 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	jwtManager := auth.NewJWTManager(jwtConfig)
	authService := auth.NewAuthService(authRepo, jwtManager, jwtConfig)
	electionService := election.NewService(electionRepo)
	electionAdminService := election.NewAdminService(electionAdminRepo)
	dptService := dpt.NewService(dptRepo)
	tpsAdminService := tps.NewAdminService(tpsAdminRepo)
	candidateService := candidate.NewService(candidatePgRepo, candidateStatsProvider)
	candidateHandler := candidate.NewHandler(candidateService)

	votingService := voting.NewVotingService(
		pool,
		electionRepo,
		voterRepo,
		candidateRepo,
		voteRepo,
		statsRepo,
		auditSvc,
	)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService)
	electionHandler := election.NewHandler(electionService)
	electionAdminHandler := election.NewAdminHandler(electionAdminService)
	votingHandler := voting.NewVotingHandler(votingService)
	dptHandler := dpt.NewHandler(dptService)
	tpsAdminHandler := tps.NewAdminHandler(tpsAdminService)
	candidateAdminHandler := candidate.NewAdminHandler(candidateService)

	logger.Info("services initialized successfully")

	allowedOrigins := parseOrigins(cfg.CORSAllowedOrigins)
	hub := ws.NewHub()
	go hub.Run(ctx)

	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	r.Handle("/metrics", promhttp.Handler())

	wsHandler := ws.NewHandler(hub)
	wsHandler.RegisterRoutes(r)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response.Success(w, http.StatusOK, map[string]string{
				"message": "PEMIRA API v1",
			})
		})

		// Auth routes (public)
		r.Post("/auth/register/student", authHandler.RegisterStudent)
		r.Post("/auth/register/lecturer-staff", authHandler.RegisterLecturerStaff)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.RefreshToken)

		// Public election routes
		r.Get("/elections/current", electionHandler.GetCurrent)
		r.Get("/elections/{electionID}/candidates", candidateHandler.ListPublic)
		r.Get("/elections/{electionID}/candidates/{candidateID}", candidateHandler.DetailPublic)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(httpMiddleware.JWTAuth(jwtManager))

			// Auth protected
			r.Get("/auth/me", authHandler.Me)
			r.Post("/auth/logout", authHandler.Logout)

			// Election routes (authenticated)
			r.Get("/elections/{electionID}/me/status", electionHandler.GetMeStatus)

			// Voting routes (student only)
			r.Group(func(r chi.Router) {
				r.Use(httpMiddleware.AuthStudentOnly(jwtManager))
				r.Post("/voting/online/cast", votingHandler.CastOnlineVote)
				r.Post("/voting/tps/cast", votingHandler.CastTPSVote)
				r.Get("/voting/tps/status", votingHandler.GetTPSVotingStatus)
				r.Get("/voting/receipt", votingHandler.GetVotingReceipt)
			})

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(httpMiddleware.AuthAdminOnly(jwtManager))

				// Election management
				r.Route("/admin/elections", func(r chi.Router) {
					r.Get("/", electionAdminHandler.List)
					r.Post("/", electionAdminHandler.Create)
					r.Get("/{electionID}", electionAdminHandler.Get)
					r.Put("/{electionID}", electionAdminHandler.Update)
					r.Post("/{electionID}/open-voting", electionAdminHandler.OpenVoting)
					r.Post("/{electionID}/close-voting", electionAdminHandler.CloseVoting)

					// Candidate management
					r.Route("/{electionID}/candidates", func(r chi.Router) {
						r.Get("/", candidateAdminHandler.List)
						r.Post("/", candidateAdminHandler.Create)
						r.Get("/{candidateID}", candidateAdminHandler.Detail)
						r.Put("/{candidateID}", candidateAdminHandler.Update)
						r.Delete("/{candidateID}", candidateAdminHandler.Delete)
						r.Post("/{candidateID}/publish", candidateAdminHandler.Publish)
						r.Post("/{candidateID}/unpublish", candidateAdminHandler.Unpublish)
					})

					// DPT management
					r.Post("/{electionID}/voters/import", dptHandler.Import)
					r.Get("/{electionID}/voters", dptHandler.List)
					r.Get("/{electionID}/voters/export", dptHandler.Export)

					// TPS monitoring per election
					r.Get("/{electionID}/tps/monitor", tpsAdminHandler.Monitor)
				})

				// TPS management
				r.Route("/admin/tps", func(r chi.Router) {
					r.Get("/", tpsAdminHandler.List)
					r.Post("/", tpsAdminHandler.Create)
					r.Get("/{tpsID}", tpsAdminHandler.Get)
					r.Put("/{tpsID}", tpsAdminHandler.Update)
					r.Delete("/{tpsID}", tpsAdminHandler.Delete)

					// QR management
					r.Get("/{tpsID}/qr", tpsAdminHandler.GetQRMetadata)
					r.Post("/{tpsID}/qr/rotate", tpsAdminHandler.RotateQR)
					r.Get("/{tpsID}/qr/print", tpsAdminHandler.GetQRForPrint)

					// Operator management
					r.Get("/{tpsID}/operators", tpsAdminHandler.ListOperators)
					r.Post("/{tpsID}/operators", tpsAdminHandler.CreateOperator)
					r.Delete("/{tpsID}/operators/{userID}", tpsAdminHandler.RemoveOperator)
				})
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting server", "port", cfg.HTTPPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
