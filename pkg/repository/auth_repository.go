package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	userExistsPrefix = "user:exists:"
	userLoginPrefix  = "user:login:"
)

type AuthRepository struct {
	postgres *AuthPostgres
	cache    *redis.Client
	cacheTTL time.Duration
}

func NewAuthRepository(pool *pgxpool.Pool, cache *redis.Client) *AuthRepository {
	return &AuthRepository{
		postgres: NewAuthPostgres(pool),
		cache:    cache,
		cacheTTL: 10 * time.Minute,
	}
}

func (r *AuthRepository) UserExists(email string) (bool, error) {
	key := userExistsPrefix + strings.ToLower(email)
	ctx := context.Background()

	if r.cache != nil {
		val, err := r.cache.Get(ctx, key).Result()
		if err == nil {
			return val == "1", nil
		}
	}

	exists, err := r.postgres.UserExists(email)
	if err != nil {
		return false, err
	}

	if r.cache != nil {
		value := "0"
		if exists {
			value = "1"
		}
		_ = r.cache.Set(ctx, key, value, r.cacheTTL).Err()
	}

	return exists, nil
}

func (r *AuthRepository) CreateUser(user model.User) (int, error) {
	id, err := r.postgres.CreateUser(user)
	if err != nil {
		return 0, err
	}

	if r.cache != nil {
		ctx := context.Background()
		existsKey := userExistsPrefix + strings.ToLower(user.Email)
		loginKey := fmt.Sprintf("%s%s:%s", userLoginPrefix, strings.ToLower(user.Email), user.Password)
		_ = r.cache.Set(ctx, existsKey, "1", r.cacheTTL).Err()
		_ = r.cache.Set(ctx, loginKey, strconv.Itoa(id), r.cacheTTL).Err()
	}

	return id, nil
}

func (r *AuthRepository) GetUser(email, password string) (model.User, error) {
	key := fmt.Sprintf("%s%s:%s", userLoginPrefix, strings.ToLower(email), password)
	ctx := context.Background()

	if r.cache != nil {
		val, err := r.cache.Get(ctx, key).Result()
		if err == nil {
			id, convErr := strconv.Atoi(val)
			if convErr != nil {
				return model.User{}, convErr
			}
			return model.User{ID: int64(id)}, nil
		}
	}

	user, err := r.postgres.GetUser(email, password)
	if err != nil {
		return model.User{}, err
	}

	if r.cache != nil && user.ID != 0 {
		_ = r.cache.Set(ctx, key, strconv.FormatInt(user.ID, 10), r.cacheTTL).Err()
	}

	return user, nil
}
