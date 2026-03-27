package service

import (
	"context"
	"time"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret string
	blocklist tokenBlocklist
}

type tokenBlocklist interface {
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, blocklist tokenBlocklist) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, blocklist: blocklist}
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

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}
	return s.issueTokens(ctx, user.ID, time.Now().UTC())
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error) {
	tokens, err := s.users.ListValidRefreshTokens(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to load refresh token")
	}

	var matchedID int64
	var matchedUserID int64
	for _, item := range tokens {
		if bcrypt.CompareHashAndPassword([]byte(item.TokenHash), []byte(rawRefreshToken)) == nil {
			matchedID = item.ID
			matchedUserID = item.UserID
			break
		}
	}
	if matchedID == 0 {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	if _, err = s.users.RevokeRefreshTokenByID(ctx, matchedID); err != nil {
		return nil, apperror.Internal("failed to rotate refresh token")
	}

	return s.issueTokens(ctx, matchedUserID, time.Now().UTC())
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
