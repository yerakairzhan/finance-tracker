package repository

import (
	"context"

	"finance-tracker/pkg/generated/sqlc"
	"finance-tracker/pkg/models"
)

type UserRepository struct {
	q *sqlc.Queries
}

func NewUserRepository(q *sqlc.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (ur *UserRepository) CreateUser(ctx context.Context, email, name string) (*models.User, error) {
	row, err := ur.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email: email,
		Name:  name,
	})

	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        int(row.ID),
		Email:     row.Email,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (ur *UserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	row, err := ur.q.GetUserByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        int(row.ID),
		Email:     row.Email,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (ur *UserRepository) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := ur.q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	var users []models.User

	for _, row := range rows {
		users = append(users, models.User{
			ID:        int(row.ID),
			Email:     row.Email,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return users, nil
}

func (ur *UserRepository) UpdateUser(ctx context.Context, id int, email, name string) (*models.User, error) {

	row, err := ur.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:    int32(id),
		Email: email,
		Name:  name,
	})

	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        int(row.ID),
		Email:     row.Email,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (ur *UserRepository) DeleteUser(ctx context.Context, id int) error {
	return ur.q.DeleteUser(ctx, int32(id))
}