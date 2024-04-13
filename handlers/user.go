package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rovilay/auth-service/models"
	"github.com/rovilay/auth-service/repository"
	"github.com/rovilay/auth-service/utils"
	"github.com/rs/zerolog"
)

type UserHandler struct {
	repo repository.UserRepository
	log  *zerolog.Logger
}

func NewUserHandler(repo repository.UserRepository, l *zerolog.Logger) *UserHandler {
	logger := l.With().Str("handlers", "UserHandler").Logger()

	return &UserHandler{
		repo: repo,
		log:  &logger,
	}
}

var validate = validator.New()

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("handler", "Signup").Logger()

	user := &models.User{}

	err := user.FromJSON(r.Body)
	if err != nil {
		h.sendError(w, err, "failed to read payload", http.StatusBadRequest, &log)
		return
	}

	if user.Username == "" {
		user.Username = h.generateUniqueUsername(r.Context(), user.Firstname, user.Lastname, &log)
	}

	// validate user input
	err = user.Validate()
	if err != nil {
		h.sendError(w, err, fmt.Sprintf("Error validating payload: %s", err), http.StatusBadRequest, &log)
		return
	}

	user.ID = uuid.New()
	err = h.repo.CreateUser(r.Context(), user)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		h.sendError(w, err, "error generating token", 0, &log)
		return
	}

	var res struct {
		Token string `json:"token"`
	}

	res.Token = token
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(res); err != nil {
		h.sendError(w, err, "failed to marshal response", 0, &log)
		return
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("handler", "Login").Logger()

	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, err, "failed to read payload", http.StatusBadRequest, &log)
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(w, err, "", http.StatusBadRequest, &log)
		return
	}

	user, err := h.repo.GetUserByIDorEmail(r.Context(), input.Email)
	if err != nil {
		h.sendError(w, err, "", http.StatusUnauthorized, &log)
		return
	}
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		h.sendError(w, errors.New("invalid email or password"), "", http.StatusUnauthorized, &log)
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		h.sendError(w, err, "error generating token", 0, &log)
		return
	}

	var res struct {
		Token string `json:"token"`
	}

	res.Token = token

	if err = json.NewEncoder(w).Encode(res); err != nil {
		h.sendError(w, err, "failed to marshal response", 0, &log)
		return
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("handler", "GetUser").Logger()
	userID := r.Context().Value(userIDKey).(string)

	user, err := h.repo.GetUserByIDorEmail(r.Context(), userID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	res := &models.UserResponse{
		ID:        user.ID,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if err = json.NewEncoder(w).Encode(res); err != nil {
		h.sendError(w, err, "failed to marshal response", 0, &log)
		return
	}
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("handler", "UpdateUser").Logger()
	userID := r.Context().Value(userIDKey).(string)

	var input models.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, err, "failed to read payload", http.StatusBadRequest, &log)
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(w, err, "", http.StatusBadRequest, &log)
		return
	}

	user, err := h.repo.GetUserByIDorEmail(r.Context(), userID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}
	if input.Firstname != "" {
		user.Firstname = input.Firstname
	}
	if input.Lastname != "" {
		user.Lastname = input.Lastname
	}
	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}

	err = h.repo.UpdateUser(r.Context(), user)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	res := &models.UserResponse{
		ID:        user.ID,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if err = json.NewEncoder(w).Encode(res); err != nil {
		h.sendError(w, err, "failed to marshal response", 0, &log)
		return
	}
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	log := h.log.With().Str("handler", "UpdateUser").Logger()
	userID := r.Context().Value(userIDKey).(string)

	var input models.UpdatePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.sendError(w, err, "failed to read payload", http.StatusBadRequest, &log)
		return
	}

	if err := validate.Struct(input); err != nil {
		h.sendError(w, err, "", http.StatusBadRequest, &log)
		return
	}

	user, err := h.repo.GetUserByIDorEmail(r.Context(), userID)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		h.sendError(w, errors.New("invalid password"), "", http.StatusBadRequest, &log)
		return
	}

	_, err = h.repo.UpdatePassword(r.Context(), userID, input.NewPassword)
	if err != nil {
		h.sendError(w, err, "", 0, &log)
		return
	}

	var res struct {
		Success string `json:"success"`
	}

	res.Success = "operation successful!"

	if err = json.NewEncoder(w).Encode(res); err != nil {
		h.sendError(w, err, "failed to marshal response", 0, &log)
		return
	}
}

func (h *UserHandler) sendError(w http.ResponseWriter, err error, errMsg string, statusCode int, log *zerolog.Logger) {
	log.Err(err)
	if errMsg == "" {
		errMsg = err.Error()
	}

	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	errRes := fmt.Sprintf(`{"error": "%v"}`, errMsg)

	if errors.Is(err, utils.ErrDuplicateEntry) || errors.Is(err, utils.ErrForeignKeyViolation) {
		http.Error(w, errRes, http.StatusBadRequest)
		return
	} else if errors.Is(err, utils.ErrNotFound) {
		http.Error(w, errRes, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, errRes, statusCode)
		return
	}
}

func (h *UserHandler) generateUniqueUsername(ctx context.Context, firstName, lastName string, log *zerolog.Logger) string {
	for {
		username := utils.GenerateUsername(firstName)

		exists, err := h.repo.CheckUserNameExist(ctx, username)
		if err != nil {
			log.Err(err).Msg("error validating username")
			return username + "." + lastName
		}

		if !exists {
			return username
		}
	}
}
