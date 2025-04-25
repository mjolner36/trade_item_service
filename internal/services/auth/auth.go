package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/mjolner36/trade_item_service/internal/domain/models"
	"github.com/mjolner36/trade_item_service/internal/lib/jwt"
	"github.com/mjolner36/trade_item_service/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("Invalid username or password")
	ErrInvalidAppID       = errors.New("Invalid app ID")
	ErrUserExist          = errors.New("User exists")
)

// New return a new instance of the Auth service.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
		log:          log,
	}
}

func (auth *Auth) Login(ctx context.Context, email, password string, appID int) (string, error) {
	const op = "auth.Login"
	log := auth.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	user, err := auth.userProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			auth.log.Warn("user not found") // slog.Error(err)

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		auth.log.Error("failed to get user", slog.Error) //// slog.Error(err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		auth.log.Info("invalid credentials", slog.Error)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := auth.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")
	token, err := jwt.NewToken(user, app, auth.tokenTTL)
	if err != nil {
		auth.log.Error("failed to create token", slog.Error)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (auth *Auth) RegisterNewUser(ctx context.Context, email, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := auth.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash") //sl.Err(err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := auth.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			slog.Warn("user already exists", err)
			//log.Warn(" slog.Error) тут заменил на slog надо смотреть разницу, пока не понимаю
			return 0, fmt.Errorf("%s: %w", op, ErrUserExist)
		}
		log.Error("failed to save user") //sl.Err(err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

func (auth *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	log := auth.log.With(
		slog.String("op", op),
		slog.Int64("userID", userID),
	)
	log.Info("checking if user is admin")
	isAdmin, err := auth.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("failed to check if user is admin")
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))
	return isAdmin, nil
}
