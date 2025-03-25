package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorInvalidCreds  = errors.New("invalid email or password")
	ErrorInvalidUserID = errors.New("invalid user id")
	ErrorUserExists    = errors.New("user already exists")
	ErrorInvalidAppID  = errors.New("user already exists")
)

type Auth struct {
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	log         *slog.Logger
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, id int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

// New returns a new instance of Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration) *Auth {
	return &Auth{userSaver, userProvider, appProvider, log, tokenTTL}
}

// Login check if user with given creds exists in the system. Returns auth token.
//
// If user does't exist, returns error
// If password is incorrect, returns error
func (a *Auth) Login(ctx context.Context, email, password string, appID int) (string, error) {
	const op = "Auth.Login"

	log := slog.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)

			return "", fmt.Errorf("%s: %w", op, ErrorInvalidCreds)
		}

		a.log.Error("failed to get user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid credentials", err)

		return "", fmt.Errorf("%s: %w", op, ErrorInvalidCreds)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			return "", fmt.Errorf("%s: %w", op, ErrorInvalidAppID)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user was logged successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser create new user with given creds. Returns created user ID.
func (a *Auth) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("register new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", err)
			return 0, fmt.Errorf("%s: %w", op, ErrorUserExists)
		}
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registered")

	return id, nil
}

// IsAdmin checks the user with given ID is admin. Returns boolean value.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := slog.With(
		slog.String("op", op),
		slog.Int64("userid", userID),
	)

	log.Info("check user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", err)

			return false, fmt.Errorf("%s: %w", op, ErrorInvalidUserID)
		}

		a.log.Error("failed to check user id admin", err)
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
