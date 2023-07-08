package ports

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	chilogger "github.com/chi-middleware/logrus-logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ibgl/microservice-users/internal/app"
	apperrors "github.com/ibgl/microservice-users/internal/app/errors"
	auth "github.com/ibgl/microservice-users/internal/app/user"
)

type HttpServer struct {
	app       app.Application
	validator *validator.Validate
}

const DEFAULT_PORT = "8088"

func NewHttpServer(app app.Application) *HttpServer {
	validate := validator.New()

	return &HttpServer{
		app:       app,
		validator: validate,
	}
}

func (h *HttpServer) Start() {
	port := DEFAULT_PORT
	if os.Getenv("APP_PORT") != "" {
		port = os.Getenv("APP_PORT")
	}

	r := chi.NewRouter()

	h.registerMiddlewares(r)
	h.registerRoutes(r)

	h.app.GetLogger().Printf("Server started on %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func (h *HttpServer) registerMiddlewares(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chilogger.Logger("router", h.app.GetLogger()))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
}

func (h *HttpServer) registerRoutes(r *chi.Mux) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, r, map[string]string{"status": "ok"})
		})

		r.Post("/signIn", h.signIn)
		r.Post("/signUp", h.signUp)
		r.Post("/refresh", h.refresh)

		r.Post("/google-signIn", h.googleSignIn)

		r.Get("/me", h.me)
		r.Put("/settings", h.updateSettings)
	})
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5"`
}

type RefreshRequest struct {
	Refresh string `json:"refresh" validate:"required"`
}

type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5"`
	Name     string `json:"name" validate:"required,gte=1"`
}

type TokenPairResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type UserResponse struct {
	UUID     uuid.UUID       `json:"uuid"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
	Settings SettingsPayload `json:"settings"`
}

type SettingsPayload struct {
	Currency          string `json:"currency" validate:"required"`
	FirstDayOfWeek    string `json:"first_day_of_week" validate:"required"`
	ProfilePictureUrl string `json:"profile_picture_url" validate:"required"`
}

type GoogleSignInRequest struct {
	Credential string `json:"credential"`
}

func (e *TokenPairResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

func (e *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, 200)
	return nil
}

func (h *HttpServer) signIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	var request LoginRequest
	if err := json.Unmarshal(body, &request); err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	serviceRequest := &auth.SignInRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	tokens, err := h.app.GetAuthService().SignIn(r.Context(), serviceRequest)
	if err != nil {
		h.RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Value,
		Refresh: tokens.Refresh.Value,
	})
}

func (h *HttpServer) RespondValidationError(errs []validator.FieldError, w http.ResponseWriter, r *http.Request) {
	for _, err := range errs {
		fmt.Println(err.Namespace())
		fmt.Println(err.Field())
		fmt.Println(err.StructNamespace())
		fmt.Println(err.StructField())
		fmt.Println(err.Tag())
		fmt.Println(err.ActualTag())
		fmt.Println(err.Kind())
		fmt.Println(err.Type())
		fmt.Println(err.Value())
		fmt.Println(err.Param())
	}

	err := errs[0]

	slug := "invalid-input"
	if err.Field() == "Password" && err.Tag() == "gte" {
		slug = "field-password-invalid-length"
	} else if err.Field() == "Email" && err.Tag() == "email" {
		slug = "field-email-invalid"
	} else if err.Field() == "Name" && err.Tag() == "gte" {
		slug = "field-name-invalid-length"
	} else if err.Field() == "Name" && err.Tag() == "required" {
		slug = "field-name-required"
	} else if err.Field() == "Password" && err.Tag() == "required" {
		slug = "field-password-required"
	} else if err.Field() == "Email" && err.Tag() == "required" {
		slug = "field-email-required"
	} else if err.Field() == "Refresh" && err.Tag() == "required" {
		slug = "invalid-token"
	} else if err.Field() == "Currency" && err.Tag() == "required" {
		slug = "field-currency-required"
	}

	h.BadRequest(slug, err, w, r)
}

func (h *HttpServer) signUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	var request SignUpRequest
	if err := json.Unmarshal(body, &request); err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	err = h.validator.Struct(request)
	if err != nil {
		h.RespondValidationError(err.(validator.ValidationErrors), w, r)
		return
	}

	serviceRequest := &auth.SignUpRequest{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
	}

	tokens, err := h.app.GetAuthService().SignUp(r.Context(), serviceRequest)
	if err != nil {
		h.RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Value,
		Refresh: tokens.Refresh.Value,
	})
}

func (h *HttpServer) updateSettings(w http.ResponseWriter, r *http.Request) {
	access, err := h.getAccessFromHeader(w, r)
	if err != nil {
		h.Unauthorised("invalid-token", err, w, r)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	var payload SettingsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	err = h.validator.Struct(payload)
	if err != nil {
		h.RespondValidationError(err.(validator.ValidationErrors), w, r)
		return
	}

	serviceRequest := auth.UpdateSettingsRequest{
		Currency:          payload.Currency,
		FirstDayOfWeek:    payload.FirstDayOfWeek,
		ProfilePictureUrl: payload.ProfilePictureUrl,
		UserUUID:          access.UserId,
	}

	user, err := h.app.GetAuthService().UpdateSettings(r.Context(), &serviceRequest)
	if err != nil {
		h.RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &UserResponse{
		UUID:  user.UUID,
		Email: user.Email,
		Name:  user.Name,
		Settings: SettingsPayload{
			Currency:          user.Settings.Currency.String(),
			FirstDayOfWeek:    user.Settings.FirstDayOfWeek.String(),
			ProfilePictureUrl: user.Settings.ProfilePictureUrl,
		},
	})
}

func (h *HttpServer) me(w http.ResponseWriter, r *http.Request) {
	access, err := h.getAccessFromHeader(w, r)
	if err != nil {
		h.Unauthorised("invalid-token", err, w, r)
		return
	}

	user, err := h.app.GetAuthService().GetUser(r.Context(), access.UserId)
	if err != nil {
		h.RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &UserResponse{
		UUID:  user.UUID,
		Email: user.Email,
		Name:  user.Name,
		Settings: SettingsPayload{
			Currency:          user.Settings.Currency.String(),
			FirstDayOfWeek:    user.Settings.FirstDayOfWeek.String(),
			ProfilePictureUrl: user.Settings.ProfilePictureUrl,
		},
	})
}

func (h *HttpServer) refresh(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	var request RefreshRequest
	if err := json.Unmarshal(body, &request); err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	err = h.validator.Struct(request)
	if err != nil {
		h.RespondValidationError(err.(validator.ValidationErrors), w, r)
		return
	}

	tokens, err := h.app.GetAuthService().Refresh(r.Context(), request.Refresh)
	if err != nil {
		h.RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Value,
		Refresh: tokens.Refresh.Value,
	})
}

func (h *HttpServer) googleSignIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // response body is []byte
	if err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	var request GoogleSignInRequest
	if err := json.Unmarshal(body, &request); err != nil {
		h.BadRequest("invalid-input", err, w, r)
		return
	}

	tokens, err := h.app.GetAuthService().GoogleSignIn(r.Context(), request.Credential)
	if err != nil {
		h.RespondWithAppError(err, w, r)
		return
	}

	render.Render(w, r, &TokenPairResponse{
		Access:  tokens.Access.Value,
		Refresh: tokens.Refresh.Value,
	})
}

func (h *HttpServer) getAccessFromHeader(w http.ResponseWriter, r *http.Request) (auth.Token, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) < 2 {
		return auth.Token{}, apperrors.NewAuthorizationError("Token not presented", "invalid-token")
	}

	tokenHeader := splitToken[1]

	access, err := h.app.GetAuthService().ValidateToken(r.Context(), tokenHeader)
	if err != nil {
		return auth.Token{}, err
	}

	return access, nil
}

func (h *HttpServer) InternalError(slug string, err error, w http.ResponseWriter, r *http.Request) {
	h.httpRespondWithError(err, slug, w, r, "Internal server error", http.StatusInternalServerError)
}

func (h *HttpServer) Unauthorised(slug string, err error, w http.ResponseWriter, r *http.Request) {
	h.app.GetLogger().Printf("Unathorized error %s", slug)
	h.httpRespondWithError(err, slug, w, r, "Unauthorised", http.StatusUnauthorized)
}

func (h *HttpServer) BadRequest(slug string, err error, w http.ResponseWriter, r *http.Request) {
	h.httpRespondWithError(err, slug, w, r, "Bad request", http.StatusBadRequest)
}

func (h *HttpServer) NotFound(slug string, err error, w http.ResponseWriter, r *http.Request) {
	h.httpRespondWithError(err, slug, w, r, "Not found", http.StatusNotFound)
}

func (h *HttpServer) RespondWithAppError(err error, w http.ResponseWriter, r *http.Request) {
	appError, ok := err.(apperrors.AppError)
	if !ok {
		h.InternalError("internal-server-error", err, w, r)
		return
	}

	switch appError.ErrorType() {
	case apperrors.ErrorTypeAuthorization:
		h.Unauthorised(appError.Slug(), appError, w, r)
	case apperrors.ErrorTypeIncorrectInput:
		h.BadRequest(appError.Slug(), appError, w, r)
	case apperrors.ErrorNotFound:
		h.NotFound(appError.Slug(), appError, w, r)
	default:
		h.InternalError(appError.Slug(), appError, w, r)
	}
}

func (h *HttpServer) httpRespondWithError(err error, slug string, w http.ResponseWriter, r *http.Request, logMSg string, status int) {
	h.app.GetLogger().Debug(map[string]string{
		"error-type": "HTTP Request Error",
		"slug":       slug,
		"error":      err.Error(),
	})

	resp := ErrorResponse{slug, status}

	if err := render.Render(w, r, resp); err != nil {
		panic(err)
	}
}

type ErrorResponse struct {
	Slug       string `json:"slug"`
	httpStatus int
}

func (e ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.httpStatus)
	return nil
}
