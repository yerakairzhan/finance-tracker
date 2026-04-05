package service

import (
	"context"
	"time"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"
	"finance-tracker/pkg/cache"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users         *repository.UserRepository
	jwtSecret     string
	blocklist     tokenBlocklist
	refreshStore  refreshSessionStore
	refreshPepper string
}

type tokenBlocklist interface {
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
}

type refreshSessionStore interface {
	CreateRefreshSession(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error
	GetRefreshSession(ctx context.Context, tokenHash string) (*cache.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, tokenHash string) error
	RotateRefreshSession(ctx context.Context, oldTokenHash, newTokenHash string, userID int64, ttl time.Duration) error
}

func NewAuthService(
	users *repository.UserRepository,
	jwtSecret string,
	blocklist tokenBlocklist,
	refreshStore refreshSessionStore,
	refreshPepper string,
) *AuthService {
	return &AuthService{
		users:         users,
		jwtSecret:     jwtSecret,
		blocklist:     blocklist,
		refreshStore:  refreshStore,
		refreshPepper: refreshPepper,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal("failed to hash password")
	}

	user, err := s.users.Create(ctx, req.Email, string(passwordHash), req.Name, req.Currency)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errorAs(err, &pgErr); ok && pgErr.Code == "23505" {
			return nil, apperror.Conflict("email already exists")
		}
		return nil, apperror.Internal("failed to create user")
	}

	tokens, appErr := s.issueTokens(ctx, user.ID, time.Now().UTC())
	if appErr != nil {
		return nil, appErr
	}
	return tokens, nil
}

// Login: create refresh session in Redis (hashed token), return refresh token only for cookie setting.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	now := time.Now().UTC()
	accessToken, err := auth.GenerateAccessToken(s.jwtSecret, user.ID, now)
	if err != nil {
		return nil, apperror.Internal("failed to issue access token")
	}
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal("failed to issue refresh token")
	}
	refreshHash, err := auth.HashRefreshToken(s.refreshPepper, refreshToken)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}
	if err := s.refreshStore.CreateRefreshSession(ctx, refreshHash, user.ID, auth.RefreshTokenTTL); err != nil {
		return nil, apperror.Internal("failed to store refresh token")
	}

	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // internal only; handler must put into HttpOnly cookie only.
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.AuthTokens, *apperror.Error) {
	oldHash, err := auth.HashRefreshToken(s.refreshPepper, refreshToken)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}
	session, err := s.refreshStore.GetRefreshSession(ctx, oldHash)
	if err != nil || session == nil || session.UserID <= 0 {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal("failed to issue refresh token")
	}
	newHash, err := auth.HashRefreshToken(s.refreshPepper, newRefreshToken)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}
	if err := s.refreshStore.RotateRefreshSession(ctx, oldHash, newHash, session.UserID, auth.RefreshTokenTTL); err != nil {
		return nil, apperror.Internal("failed to rotate refresh token")
	}

	accessToken, err := auth.GenerateAccessToken(s.jwtSecret, session.UserID, time.Now().UTC())
	if err != nil {
		return nil, apperror.Internal("failed to issue access token")
	}
	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken, // internal only; handler must put into cookie.
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int64, rawRefreshToken, rawAccessToken string) *apperror.Error {
	tokens, err := s.users.ListValidRefreshTokensByUser(ctx, userID)
	if err != nil {
		return apperror.Internal("failed to load refresh token")
	}

	var matchedID int64
	for _, item := range tokens {
		if bcrypt.CompareHashAndPassword([]byte(item.TokenHash), []byte(rawRefreshToken)) == nil {
			matchedID = item.ID
			break
		}
	}
	if matchedID == 0 {
		return apperror.NotFound("refresh token not found")
	}

	affected, err := s.users.RevokeRefreshTokenByIDForUser(ctx, matchedID, userID)
	if err != nil {
		return apperror.Internal("failed to revoke refresh token")
	}
	if affected == 0 {
		return apperror.NotFound("refresh token not found")
	}

	claims, err := auth.ParseAccessToken(s.jwtSecret, rawAccessToken)
	if err != nil {
		return apperror.Unauthorized("invalid or expired token")
	}
	if claims.ID == "" || claims.ExpiresAt == nil {
		return apperror.Unauthorized("invalid token")
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl > 0 {
		if err = s.blocklist.Revoke(ctx, claims.ID, ttl); err != nil {
			return apperror.Internal("failed to revoke access token")
		}
	}
	return nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID int64, now time.Time) (*models.AuthTokens, *apperror.Error) {
	access, err := auth.GenerateAccessToken(s.jwtSecret, userID, now)
	if err != nil {
		return nil, apperror.Internal("failed to issue access token")
	}

	refreshRaw, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal("failed to issue refresh token")
	}

	refreshHash, err := bcrypt.GenerateFromPassword([]byte(refreshRaw), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal("failed to hash refresh token")
	}

	expiresAt := pgtype.Timestamptz{Time: now.Add(auth.RefreshTokenTTL), Valid: true}
	if err = s.users.InsertRefreshToken(ctx, userID, string(refreshHash), expiresAt); err != nil {
		return nil, apperror.Internal("failed to store refresh token")
	}

	return &models.AuthTokens{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	}, nil
}
