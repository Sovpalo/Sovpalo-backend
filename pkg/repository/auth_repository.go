package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	userExistsPrefix       = "user:exists:"
	usernameExistsPrefix   = "user:username-exists:"
	userLoginPrefix        = "user:login:"
	authChallengeKeyPrefix = "auth:challenge:"
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

func (r *AuthRepository) UsernameExists(username string) (bool, error) {
	key := usernameExistsPrefix + strings.ToLower(username)
	ctx := context.Background()

	if r.cache != nil {
		val, err := r.cache.Get(ctx, key).Result()
		if err == nil {
			return val == "1", nil
		}
	}

	exists, err := r.postgres.UsernameExists(username)
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

	r.cacheUser(user, int64(id))
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

func (r *AuthRepository) GetUserByEmail(email string) (model.User, error) {
	return r.postgres.GetUserByEmail(email)
}

func (r *AuthRepository) GetUserByID(userID int64) (model.User, error) {
	return r.postgres.GetUserByID(userID)
}

func (r *AuthRepository) UpdateUserPassword(email string, passwordHash string) error {
	user, err := r.postgres.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if err := r.postgres.UpdateUserPassword(email, passwordHash); err != nil {
		return err
	}

	if r.cache != nil {
		ctx := context.Background()
		pattern := fmt.Sprintf("%s%s:*", userLoginPrefix, strings.ToLower(email))
		iter := r.cache.Scan(ctx, 0, pattern, 100).Iterator()
		for iter.Next(ctx) {
			_ = r.cache.Del(ctx, iter.Val()).Err()
		}
		if iter.Err() != nil {
			return iter.Err()
		}
	}

	r.cacheUser(model.User{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Password: passwordHash,
	}, user.ID)

	return nil
}

func (r *AuthRepository) SavePendingAuthChallenge(challenge model.PendingAuthChallenge, ttl time.Duration) error {
	if r.cache == nil {
		return errors.New("auth challenge storage unavailable")
	}

	payload, err := json.Marshal(challenge)
	if err != nil {
		return err
	}

	key := authChallengeKey(challenge.Type, challenge.Email)
	return r.cache.Set(context.Background(), key, payload, ttl).Err()
}

func (r *AuthRepository) GetPendingAuthChallenge(challengeType model.AuthChallengeType, email string) (model.PendingAuthChallenge, error) {
	if r.cache == nil {
		return model.PendingAuthChallenge{}, errors.New("auth challenge storage unavailable")
	}

	key := authChallengeKey(challengeType, email)
	val, err := r.cache.Get(context.Background(), key).Result()
	if err != nil {
		return model.PendingAuthChallenge{}, err
	}

	var challenge model.PendingAuthChallenge
	if err := json.Unmarshal([]byte(val), &challenge); err != nil {
		return model.PendingAuthChallenge{}, err
	}

	return challenge, nil
}

func (r *AuthRepository) DeletePendingAuthChallenge(challengeType model.AuthChallengeType, email string) error {
	if r.cache == nil {
		return errors.New("auth challenge storage unavailable")
	}

	key := authChallengeKey(challengeType, email)
	return r.cache.Del(context.Background(), key).Err()
}

func (r *AuthRepository) cacheUser(user model.User, id int64) {
	if r.cache == nil {
		return
	}

	ctx := context.Background()
	existsKey := userExistsPrefix + strings.ToLower(user.Email)
	usernameKey := usernameExistsPrefix + strings.ToLower(user.Username)
	loginKey := fmt.Sprintf("%s%s:%s", userLoginPrefix, strings.ToLower(user.Email), user.Password)
	_ = r.cache.Set(ctx, existsKey, "1", r.cacheTTL).Err()
	_ = r.cache.Set(ctx, usernameKey, "1", r.cacheTTL).Err()
	_ = r.cache.Set(ctx, loginKey, strconv.FormatInt(id, 10), r.cacheTTL).Err()
}

func authChallengeKey(challengeType model.AuthChallengeType, email string) string {
	return fmt.Sprintf("%s%s:%s", authChallengeKeyPrefix, challengeType, strings.ToLower(email))
}
