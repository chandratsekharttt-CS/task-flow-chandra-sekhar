package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"golang.org/x/crypto/bcrypt"

	"taskflow/internal/config"
	"taskflow/internal/database"
	"taskflow/internal/handlers"
	appMiddleware "taskflow/internal/middleware"
	"taskflow/internal/repository"
)

func main() {
	// Set up structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	slog.Info("starting TaskFlow API server")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)

	// Seed database if enabled
	if cfg.SeedDB {
		if err := seedDatabase(ctx, userRepo, projectRepo, taskRepo, cfg.BcryptCost); err != nil {
			slog.Error("failed to seed database", "error", err)
			// Don't exit — seeding failure shouldn't prevent startup
		}
	}

	// Initialize middleware & handlers
	authMiddle := appMiddleware.NewAuthMiddleware(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(userRepo, authMiddle, cfg.BcryptCost)
	projectHandler := handlers.NewProjectHandler(projectRepo, taskRepo)
	taskHandler := handlers.NewTaskHandler(taskRepo, projectRepo)

	// Set up Chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appMiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(appMiddleware.CORSOptions()))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public auth routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddle.Authenticate)

		// Auth
		r.Get("/api/auth/me", authHandler.Me)

		// Projects
		r.Get("/api/projects", projectHandler.List)
		r.Post("/api/projects", projectHandler.Create)
		r.Get("/api/projects/{id}", projectHandler.Get)
		r.Patch("/api/projects/{id}", projectHandler.Update)
		r.Delete("/api/projects/{id}", projectHandler.Delete)
		r.Get("/api/projects/{id}/stats", projectHandler.Stats)

		// Tasks
		r.Get("/api/projects/{id}/tasks", taskHandler.List)
		r.Post("/api/projects/{id}/tasks", taskHandler.Create)
		r.Patch("/api/tasks/{id}", taskHandler.Update)
		r.Delete("/api/tasks/{id}", taskHandler.Delete)
	})

	// HTTP server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Block until shutdown signal
	sig := <-shutdownCh
	slog.Info("shutdown signal received", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}

// runMigrations applies all pending database migrations.
func runMigrations(databaseURL string) error {
	// Migration files path — configurable for Docker vs local dev
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	m, err := migrate.New(migrationsPath, "pgx5://"+stripScheme(databaseURL))
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	slog.Info("database migrations applied successfully")
	return nil
}

// stripScheme removes the "postgres://" or "postgresql://" prefix for pgx5 driver.
func stripScheme(url string) string {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if len(url) > len(prefix) && url[:len(prefix)] == prefix {
			return url[len(prefix):]
		}
	}
	return url
}

// seedDatabase creates test data if the database is empty.
func seedDatabase(ctx context.Context, userRepo *repository.UserRepository, projectRepo *repository.ProjectRepository, taskRepo *repository.TaskRepository, bcryptCost int) error {
	// Check if data already exists
	count, err := userRepo.CountAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	if count > 0 {
		slog.Info("database already has data, skipping seed")
		return nil
	}

	slog.Info("seeding database with test data...")

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash seed password: %w", err)
	}

	user, err := userRepo.Create(ctx, "Test User", "testuser@taskflow.com", string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to create seed user: %w", err)
	}
	slog.Info("seed user created", "email", user.Email, "id", user.ID)

	// Create sample project
	desc := "A sample project to demonstrate TaskFlow features."
	project, err := projectRepo.Create(ctx, "Sample Project", &desc, user.ID)
	if err != nil {
		return fmt.Errorf("failed to create seed project: %w", err)
	}
	slog.Info("seed project created", "name", project.Name, "id", project.ID)

	// Create 3 tasks with different statuses
	taskDesc1 := "Create wireframes and mockups for the main landing page."
	dueDate := "2025-04-30"
	_, err = taskRepo.Create(ctx, "Design the landing page", &taskDesc1, "todo", "high", project.ID, &user.ID, user.ID, &dueDate)
	if err != nil {
		return fmt.Errorf("failed to create seed task 1: %w", err)
	}

	taskDesc2 := "Configure GitHub Actions for automated testing and deployment."
	_, err = taskRepo.Create(ctx, "Set up CI/CD pipeline", &taskDesc2, "in_progress", "medium", project.ID, &user.ID, user.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create seed task 2: %w", err)
	}

	taskDesc3 := "Document setup instructions and API reference."
	_, err = taskRepo.Create(ctx, "Write project README", &taskDesc3, "done", "low", project.ID, nil, user.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create seed task 3: %w", err)
	}

	slog.Info("database seeded successfully", "user", user.Email, "project", project.Name, "tasks", 3)
	return nil
}
