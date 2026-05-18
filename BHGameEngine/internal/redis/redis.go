package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Addr     string // Redis地址
	Password string // Redis密码
	DB       int    // 数据库编号
	PoolSize int    // 连接池大小
}

type RedisClient struct {
	client *redis.Client   // Redis客户端实例
	config RedisConfig     // Redis配置
	ctx    context.Context // 上下文
}

func NewRedisClient(config RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	return &RedisClient{
		client: client,
		config: config,
		ctx:    context.Background(),
	}
}

func (r *RedisClient) Ping() error {
	_, err := r.client.Ping(r.ctx).Result()
	return err
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisClient) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisClient) Exists(key string) (bool, error) {
	val, err := r.client.Exists(r.ctx, key).Result()
	return val > 0, err
}

func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

func (r *RedisClient) HSet(key string, values ...interface{}) error {
	return r.client.HSet(r.ctx, key, values...).Err()
}

func (r *RedisClient) HGet(key string, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, key).Result()
}

func (r *RedisClient) HDel(key string, fields ...string) error {
	return r.client.HDel(r.ctx, key, fields...).Err()
}

func (r *RedisClient) LPush(key string, values ...interface{}) error {
	return r.client.LPush(r.ctx, key, values...).Err()
}

func (r *RedisClient) RPop(key string) (string, error) {
	return r.client.RPop(r.ctx, key).Result()
}

func (r *RedisClient) LLen(key string) (int64, error) {
	return r.client.LLen(r.ctx, key).Result()
}

func (r *RedisClient) SAdd(key string, members ...interface{}) error {
	return r.client.SAdd(r.ctx, key, members...).Err()
}

func (r *RedisClient) SMembers(key string) ([]string, error) {
	return r.client.SMembers(r.ctx, key).Result()
}

func (r *RedisClient) SRem(key string, members ...interface{}) error {
	return r.client.SRem(r.ctx, key, members...).Err()
}

func (r *RedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return r.client.SIsMember(r.ctx, key, member).Result()
}

func (r *RedisClient) Incr(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

func (r *RedisClient) Decr(key string) (int64, error) {
	return r.client.Decr(r.ctx, key).Result()
}

func (r *RedisClient) IncrBy(key string, value int64) (int64, error) {
	return r.client.IncrBy(r.ctx, key, value).Result()
}

func (r *RedisClient) ZAdd(key string, members ...*redis.Z) error {
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

func (r *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, stop).Result()
}

func (r *RedisClient) ZRevRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(r.ctx, key, start, stop).Result()
}

func (r *RedisClient) ZScore(key string, member string) (float64, error) {
	return r.client.ZScore(r.ctx, key, member).Result()
}

func (r *RedisClient) ZRank(key string, member string) (int64, error) {
	return r.client.ZRank(r.ctx, key, member).Result()
}

func (r *RedisClient) Lock(key string, value string, expiration time.Duration) (bool, error) {
	result, err := r.client.SetNX(r.ctx, key, value, expiration).Result()
	return result, err
}

func (r *RedisClient) Unlock(key string, value string) (bool, error) {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := r.client.Eval(r.ctx, script, []string{key}, value).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

func (r *RedisClient) GetOnlinePlayers() ([]string, error) {
	return r.client.SMembers(r.ctx, "online_players").Result()
}

func (r *RedisClient) AddOnlinePlayer(playerID int64) error {
	return r.client.SAdd(r.ctx, "online_players", playerID).Err()
}

func (r *RedisClient) RemoveOnlinePlayer(playerID int64) error {
	return r.client.SRem(r.ctx, "online_players", playerID).Err()
}

func (r *RedisClient) IsOnline(playerID int64) (bool, error) {
	return r.client.SIsMember(r.ctx, "online_players", playerID).Result()
}

func (r *RedisClient) GetSession(sessionID string) (string, error) {
	return r.client.Get(r.ctx, fmt.Sprintf("session:%s", sessionID)).Result()
}

func (r *RedisClient) SetSession(sessionID string, playerID int64, expiration time.Duration) error {
	return r.client.Set(r.ctx, fmt.Sprintf("session:%s", sessionID), playerID, expiration).Err()
}

func (r *RedisClient) DeleteSession(sessionID string) error {
	return r.client.Del(r.ctx, fmt.Sprintf("session:%s", sessionID)).Err()
}

func (r *RedisClient) RemoveSession(sessionID string) error {
	return r.DeleteSession(sessionID)
}

func (r *RedisClient) GetPlayerPosition(playerID int64) (string, error) {
	return r.client.HGet(r.ctx, fmt.Sprintf("player:%d", playerID), "position").Result()
}

func (r *RedisClient) SetPlayerPosition(playerID int64, position string) error {
	return r.client.HSet(r.ctx, fmt.Sprintf("player:%d", playerID), "position", position).Err()
}

func (r *RedisClient) GetPlayerData(playerID int64) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, fmt.Sprintf("player:%d", playerID)).Result()
}

func (r *RedisClient) SetPlayerData(playerID int64, data map[string]interface{}) error {
	return r.client.HSet(r.ctx, fmt.Sprintf("player:%d", playerID), data).Err()
}

func (r *RedisClient) GetServerLoad(serverID string) (float64, error) {
	val, err := r.client.Get(r.ctx, fmt.Sprintf("server_load:%s", serverID)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (r *RedisClient) SetServerLoad(serverID string, load float64) error {
	return r.client.Set(r.ctx, fmt.Sprintf("server_load:%s", serverID), load, 0).Err()
}

func (r *RedisClient) Publish(channel string, message string) error {
	return r.client.Publish(r.ctx, channel, message).Err()
}

func (r *RedisClient) Subscribe(channel string) *redis.PubSub {
	return r.client.Subscribe(r.ctx, channel)
}
