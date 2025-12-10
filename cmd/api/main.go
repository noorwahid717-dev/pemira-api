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
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pemira-api/internal/adminuser"
	"pemira-api/internal/auth"
	"pemira-api/internal/candidate"
	"pemira-api/internal/config"
	"pemira-api/internal/dpt"
	"pemira-api/internal/election"
	"pemira-api/internal/electionvoter"
	httpMiddleware "pemira-api/internal/http/middleware"
	"pemira-api/internal/http/response"
	"pemira-api/internal/master"
	"pemira-api/internal/monitoring"
	"pemira-api/internal/settings"
	"pemira-api/internal/tps"
	"pemira-api/internal/voter"
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
	monitoringRepo := monitoring.NewPgRepository(pool)
	tpsRepo := tps.NewPostgresRepositoryFromPool(pool)

	voterRepo := voting.NewVoterRepository()
	candidateRepo := voting.NewCandidateRepository()
	voteRepo := voting.NewVoteRepository()
	statsRepo := voting.NewVoteStatsRepository()
	auditSvc := voting.NewAuditService()

	// Voter profile repositories
	voterProfileRepo := voter.NewPgRepository(pool)
	voterAuthRepo := voter.NewAuthRepositoryAdapter(pool)

	// Settings repository
	settingsRepo := settings.NewRepository(pool)

	// Master data repository
	masterRepo := master.NewPgxRepository(pool)

	// Initialize services
	jwtConfig := auth.JWTConfig{
		Secret:          cfg.JWTSecret,
		AccessTokenTTL:  24 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	jwtManager := auth.NewJWTManager(jwtConfig)
	authService := auth.NewAuthService(authRepo, jwtManager, jwtConfig)

	// Set master repository for auth service (for registration validation)
	masterAdapter := auth.NewMasterRepositoryAdapter(masterRepo)
	authService.SetMasterRepository(masterAdapter)

	electionService := election.NewService(electionRepo, electionAdminRepo)
	electionAdminService := election.NewAdminService(electionAdminRepo)
	dptService := dpt.NewService(dptRepo)
	tpsAdminService := tps.NewAdminService(tpsAdminRepo)
	tpsService := tps.NewService(tpsRepo)
	tpsPanelService := tps.NewPanelService(tpsRepo)
	candidateService := candidate.NewService(candidatePgRepo, candidateStatsProvider)
	candidateHandler := candidate.NewHandler(candidateService)
	monitoringService := monitoring.NewService(monitoringRepo)

	votingService := voting.NewVotingService(
		pool,
		electionRepo,
		voterRepo,
		candidateRepo,
		voteRepo,
		statsRepo,
		auditSvc,
	)

	voterProfileService := voter.NewService(voterProfileRepo, voterAuthRepo)
	settingsService := settings.NewService(settingsRepo)
	electionVoterRepo := electionvoter.NewPgRepository(pool)
	electionVoterService := electionvoter.NewService(electionVoterRepo)
	adminUserRepo := adminuser.NewPgRepository(pool)
	adminUserService := adminuser.NewService(adminUserRepo)
	masterService := master.NewService(masterRepo)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService)
	electionHandler := election.NewHandler(electionService)
	electionAdminHandler := election.NewAdminHandler(electionAdminService)
	votingHandler := voting.NewVotingHandler(votingService)
	dptHandler := dpt.NewHandler(dptService)
	tpsAdminHandler := tps.NewAdminHandler(tpsAdminService)
	tpsHandler := tps.NewTPSHandler(tpsService)
	tpsPanelHandler := tps.NewPanelHandler(tpsPanelService)
	tpsPanelAuthHandler := tps.NewPanelAuthHandler(authService, tpsRepo)
	candidateAdminHandler := candidate.NewAdminHandler(candidateService)
	monitoringHandler := monitoring.NewHandler(monitoringService)
	voterProfileHandler := voter.NewProfileHandler(voterProfileService)
	settingsHandler := settings.NewHandler(settingsService)
	electionVoterHandler := electionvoter.NewHandler(electionVoterService)
	adminUserHandler := adminuser.NewHandler(adminUserService)
	masterHandler := master.NewHandler(masterService)

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

	// Leapcell health check endpoints
	r.Get("/kaithhealthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Get("/kaithheathcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
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

		// Metadata for dropdowns
		r.Get("/meta/faculties-programs", masterHandler.GetFacultyPrograms)
		r.Get("/master/faculties", masterHandler.GetFaculties)
		r.Get("/master/study-programs", masterHandler.GetStudyPrograms)
		r.Get("/master/lecturer-units", masterHandler.GetLecturerUnits)
		r.Get("/master/lecturer-positions", masterHandler.GetLecturerPositions)
		r.Get("/master/staff-units", masterHandler.GetStaffUnits)
		r.Get("/master/staff-positions", masterHandler.GetStaffPositions)

		// Auth routes (public)
		r.Post("/auth/register/student", authHandler.RegisterStudent)
		r.Post("/auth/register/lecturer-staff", authHandler.RegisterLecturerStaff)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.RefreshToken)
		r.Get("/auth/logout-page", authHandler.LogoutPage)
		r.Post("/tps-panel/auth/login", tpsPanelAuthHandler.PanelLogin)

		// Public election routes
		r.Get("/elections/current", electionHandler.GetCurrent)
		r.Get("/elections/current-for-registration", electionHandler.GetCurrentForRegistration)
		r.Get("/elections", electionHandler.ListPublic)
		r.Get("/elections/{electionID}/phases", electionHandler.GetPublicPhases)
		r.Get("/elections/{electionID}/timeline", electionHandler.GetPublicPhases)
		r.Get("/elections/{electionID}/candidates", candidateHandler.ListPublic)
		r.Get("/elections/{electionID}/candidates/{candidateID}", candidateHandler.DetailPublic)
		r.Get("/elections/{electionID}/candidates/{candidateID}/media/profile", candidateHandler.GetPublicProfileMedia)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(httpMiddleware.JWTAuth(jwtManager))

			// Auth protected
			r.Get("/auth/me", authHandler.Me)
			r.Post("/auth/logout", authHandler.Logout)

			// Voter profile routes (authenticated - voter only)
			voterProfileHandler.RegisterRoutes(r)

			// Election routes (authenticated)
			r.Get("/elections/{electionID}/me/status", electionHandler.GetMeStatus)
			r.Get("/elections/{electionID}/me/history", electionHandler.GetMeHistory)

			// Election-specific voter enrollment (self-service)
			r.Post("/voters/me/elections/{electionID}/register", electionVoterHandler.VoterSelfRegister)
			r.Get("/voters/me/elections/{electionID}/status", electionVoterHandler.VoterStatus)

			// Voter TPS QR (student/admin)
			r.Get("/voters/{voterID}/tps/qr", votingHandler.GetVoterTPSQR)
			r.Post("/voters/{voterID}/tps/qr", votingHandler.GenerateVoterTPSQR)

			// TPS student check-in
			r.Post("/tps/checkin/scan", tpsHandler.ScanQR)
			r.Get("/tps/checkin/status", tpsHandler.StudentCheckinStatus)

			// Voting routes (student only)
			r.Group(func(r chi.Router) {
				r.Use(httpMiddleware.AuthStudentOnly(jwtManager))
				r.Post("/voting/online/cast", votingHandler.CastOnlineVote)
				r.Post("/voting/tps/cast", votingHandler.CastTPSVote)
				r.Post("/voting/tps/ballots/parse-qr", votingHandler.ParseBallotQR)
				r.Post("/voting/tps/ballots/cast-from-qr", votingHandler.CastBallotFromQR)
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
					r.Patch("/{electionID}", electionAdminHandler.PatchGeneralInfo)
					r.Post("/{electionID}/open-voting", electionAdminHandler.OpenVoting)
					r.Post("/{electionID}/close-voting", electionAdminHandler.CloseVoting)
					r.Route("/{electionID}/actions", func(r chi.Router) {
						r.Post("/open-voting", electionAdminHandler.OpenVoting)
						r.Post("/close-voting", electionAdminHandler.CloseVoting)
						r.Post("/archive", electionAdminHandler.Archive)
					})
					r.Route("/{electionID}/phases", func(r chi.Router) {
						r.Get("/", electionAdminHandler.GetPhases)
						r.Put("/", electionAdminHandler.UpdatePhases)
					})
					r.Route("/{electionID}/settings", func(r chi.Router) {
						r.Get("/", electionAdminHandler.GetAllSettings)
						r.Get("/mode", electionAdminHandler.GetModeSettings)
						r.Put("/mode", electionAdminHandler.UpdateModeSettings)
					})
					r.Get("/{electionID}/summary", electionAdminHandler.GetSummary)
					r.Route("/{electionID}/branding", func(r chi.Router) {
						r.Get("/", electionAdminHandler.GetBranding)
						r.Get("/logo/{slot}", electionAdminHandler.GetBrandingLogo)
						r.Post("/logo/{slot}", electionAdminHandler.UploadBrandingLogo)
						r.Delete("/logo/{slot}", electionAdminHandler.DeleteBrandingLogo)
					})

					// TPS election-scoped management
					r.Route("/{electionID}/tps", func(r chi.Router) {
						r.Get("/", tpsHandler.AdminListTPSElection)
						r.Post("/", tpsHandler.AdminCreateTPSElection)
						r.Get("/{tpsID}", tpsHandler.AdminGetTPSElection)
						r.Put("/{tpsID}", tpsHandler.AdminUpdateTPSElection)
						r.Delete("/{tpsID}", tpsHandler.AdminDeleteTPSElection)
						r.Get("/{tpsID}/qr", tpsHandler.AdminGetQRMetadata)
						r.Post("/{tpsID}/qr/rotate", tpsHandler.AdminRotateQR)
						r.Get("/{tpsID}/qr/print", tpsHandler.AdminGetQRPrint)
						r.Get("/{tpsID}/operators", tpsHandler.AdminListOperators)
						r.Post("/{tpsID}/operators", tpsHandler.AdminCreateOperator)
						r.Delete("/{tpsID}/operators/{userID}", tpsHandler.AdminDeleteOperator)
						r.Get("/{tpsID}/allocation", tpsAdminHandler.Allocation)
						r.Get("/{tpsID}/activity", tpsAdminHandler.Activity)
					})

					// NOTE: Per-election TPS management moved to standalone route at line ~400
					// to avoid nested path issue: /admin/elections/{electionID}/tps/{tpsID}
					// (not /admin/elections/{electionID}/tps/{electionID}/tps/{tpsID})

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
					r.Route("/{electionID}/voters", func(r chi.Router) {
						r.Get("/", electionVoterHandler.AdminList)
						r.Post("/", electionVoterHandler.AdminUpsert)
						r.Get("/lookup", electionVoterHandler.AdminLookup)
						r.Patch("/{voterID}", electionVoterHandler.AdminPatch)
						r.Get("/export", dptHandler.Export)
						r.Get("/{voterID}", dptHandler.Get)
						r.Put("/{voterID}", dptHandler.Update)
						r.Delete("/{voterID}", dptHandler.Delete)
					})

					// TPS monitoring per election
					r.Get("/{electionID}/tps/monitor", tpsAdminHandler.Monitor)
				})

				// Candidate media management (global by candidate ID)
				r.Route("/admin/candidates", func(r chi.Router) {
					r.Post("/{candidateID}/media/profile", candidateAdminHandler.UploadProfileMedia)
					r.Get("/{candidateID}/media/profile", candidateAdminHandler.GetProfileMedia)
					r.Delete("/{candidateID}/media/profile", candidateAdminHandler.DeleteProfileMedia)
					r.Post("/{candidateID}/media", candidateAdminHandler.UploadMedia)
					r.Get("/{candidateID}/media/{mediaID}", candidateAdminHandler.GetMedia)
					r.Delete("/{candidateID}/media/{mediaID}", candidateAdminHandler.DeleteMedia)
				})

				// Global voters endpoint
				r.Route("/admin/voters", func(r chi.Router) {
					r.Get("/", dptHandler.ListAll)
				})

				// Admin user management
				r.Route("/admin/users", func(r chi.Router) {
					r.Get("/", adminUserHandler.List)
					r.Post("/", adminUserHandler.Create)
					r.Get("/{userID}", adminUserHandler.Detail)
					r.Patch("/{userID}", adminUserHandler.Update)
					r.Post("/{userID}/reset-password", adminUserHandler.ResetPassword)
					r.Post("/{userID}/activate", adminUserHandler.Activate)
					r.Post("/{userID}/deactivate", adminUserHandler.Deactivate)
					r.Delete("/{userID}", adminUserHandler.Delete)
				})

				// App Settings
				r.Route("/admin/settings", func(r chi.Router) {
					r.Get("/", settingsHandler.GetSettings)
					r.Get("/active-election", settingsHandler.GetActiveElection)
					r.Put("/active-election", settingsHandler.UpdateActiveElection)
				})

				// Monitoring (counts/participation)
				monitoringHandler.RegisterRoutes(r)

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

					// Allocation & activity
					r.Get("/{tpsID}/allocation", tpsAdminHandler.Allocation)
					r.Get("/{tpsID}/activity", tpsAdminHandler.Activity)
				})

			})

			// TPS panel endpoints under admin namespace (admin + TPS operator scoped)
			r.Route("/admin/elections/{electionID}/tps/{tpsID}", func(r chi.Router) {
				r.Use(httpMiddleware.AuthAdminOrTPSOperator(jwtManager))
				r.Get("/dashboard", tpsPanelHandler.Dashboard)
				r.Get("/stats", tpsPanelHandler.Stats)
				r.Get("/status", tpsPanelHandler.Status)
				r.Get("/checkins", tpsPanelHandler.ListCheckins)
				r.Get("/checkins/{checkinId}", tpsPanelHandler.GetCheckin)
				r.Post("/checkin/scan", tpsPanelHandler.ScanCheckin)
				r.Post("/checkin/manual", tpsPanelHandler.ManualCheckin)
				r.Get("/stats/timeline", tpsPanelHandler.Timeline)
				r.Get("/logs", tpsPanelHandler.Logs)

				// Admin-only TPS management endpoints
				r.With(httpMiddleware.AuthAdminOnly(jwtManager)).Get("/operators", tpsHandler.AdminListOperators)
				r.With(httpMiddleware.AuthAdminOnly(jwtManager)).Post("/operators", tpsHandler.AdminCreateOperator)
				r.With(httpMiddleware.AuthAdminOnly(jwtManager)).Delete("/operators/{userID}", tpsHandler.AdminDeleteOperator)
				r.With(httpMiddleware.AuthAdminOnly(jwtManager)).Get("/allocation", tpsAdminHandler.Allocation)
				r.With(httpMiddleware.AuthAdminOnly(jwtManager)).Get("/activity", tpsAdminHandler.Activity)
			})

			r.Group(func(r chi.Router) {
				r.Use(httpMiddleware.AuthTPSOperatorOnly(jwtManager))
				r.Post("/tps/{tpsID}/checkins/{checkinID}/scan-candidate", votingHandler.ScanTPSCandidate)
				r.Post("/tps/{tpsID}/checkins", tpsPanelHandler.CreateCheckinSimple)
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
