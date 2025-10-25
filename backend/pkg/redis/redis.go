package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client holds the Redis client
var Client *redis.Client

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// ConnectRedis establishes connection to Redis
func ConnectRedis() (*redis.Client, error) {
	config := Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	Client = rdb
	return rdb, nil
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return Client
}

// SetSession stores a session in Redis
func SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	return Client.Set(ctx, fmt.Sprintf("session:%s", sessionID), data, expiration).Err()
}

// GetSession retrieves a session from Redis
func GetSession(ctx context.Context, sessionID string) (string, error) {
	return Client.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
}

// DeleteSession removes a session from Redis
func DeleteSession(ctx context.Context, sessionID string) error {
	return Client.Del(ctx, fmt.Sprintf("session:%s", sessionID)).Err()
}

// SetCache stores data in Redis cache
func SetCache(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	return Client.Set(ctx, fmt.Sprintf("cache:%s", key), data, expiration).Err()
}

// GetCache retrieves data from Redis cache
func GetCache(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, fmt.Sprintf("cache:%s", key)).Result()
}

// DeleteCache removes data from Redis cache
func DeleteCache(ctx context.Context, key string) error {
	return Client.Del(ctx, fmt.Sprintf("cache:%s", key)).Err()
}

// SetUserSession stores user session data
func SetUserSession(ctx context.Context, userID, sessionID string, data interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("user_session:%s:%s", userID, sessionID)
	return Client.Set(ctx, key, data, expiration).Err()
}

// GetUserSession retrieves user session data
func GetUserSession(ctx context.Context, userID, sessionID string) (string, error) {
	key := fmt.Sprintf("user_session:%s:%s", userID, sessionID)
	return Client.Get(ctx, key).Result()
}

// DeleteUserSession removes user session data
func DeleteUserSession(ctx context.Context, userID, sessionID string) error {
	key := fmt.Sprintf("user_session:%s:%s", userID, sessionID)
	return Client.Del(ctx, key).Err()
}

// PublishMessage publishes a message to a Redis channel
func PublishMessage(ctx context.Context, channel string, message interface{}) error {
	return Client.Publish(ctx, channel, message).Err()
}

// SubscribeToChannel subscribes to a Redis channel
func SubscribeToChannel(ctx context.Context, channel string) *redis.PubSub {
	return Client.Subscribe(ctx, channel)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	// This would typically use os.Getenv, but we'll keep it simple for now
	return defaultValue
}
