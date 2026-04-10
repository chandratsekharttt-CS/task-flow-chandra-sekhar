package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"taskflow/internal/middleware"
	"taskflow/internal/models"
	"taskflow/internal/repository"
	"taskflow/internal/validator"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	userRepo   *repository.UserRepository
	authMiddle *middleware.AuthMiddleware
	bcryptCost int
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userRepo *repository.UserRepository, authMiddle *middleware.AuthMiddleware, bcryptCost int) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		authMiddle: authMiddle,
		bcryptCost: bcryptCost,
	}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Validate input
	if ve := validator.ValidateRegister(&req); ve != nil {
		respondValidationError(w, ve.Fields)
		return
	}

	// Check if email already exists
	exists, err := h.userRepo.ExistsByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Error("failed to check email existence", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if exists {
		respondValidationError(w, map[string]string{"email": "is already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.bcryptCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create user
	user, err := h.userRepo.Create(r.Context(), req.Name, req.Email, string(hashedPassword))
	if err != nil {
		slog.Error("failed to create user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Generate JWT
	token, err := h.authMiddle.GenerateToken(user.ID, user.Email)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("user registered", "user_id", user.ID, "email", user.Email)
	respondJSON(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Validate input
	if ve := validator.ValidateLogin(&req); ve != nil {
		respondValidationError(w, ve.Fields)
		return
	}

	// Find user
	user, err := h.userRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Error("failed to find user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if user == nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Generate JWT
	token, err := h.authMiddle.GenerateToken(user.ID, user.Email)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("user logged in", "user_id", user.ID, "email", user.Email)
	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Me handles GET /api/auth/me — returns the current user info.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		slog.Error("failed to find user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respondJSON(w, http.StatusOK, user)
}
