package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/OpenIMSDK/protocol/encryption"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

const (
	EncryptionKeyPrefix = "CON_ENCRYPTION_KEY:"
	MaxVersionKeyField  = "MAX_VERSION"
)

type EncryptionCache interface {
	metaCache
	GetConversationEncryptionAllKey(ctx context.Context, conversationID string) ([]*encryption.VersionKey, error)
	GetConversationVersionKey(ctx context.Context, conversationID string, version int32) (string, error)
	GenConversationNextVersionKey(ctx context.Context, conversationID string, fn func() (string, error)) error
	GetConversationMaxVersionKey(ctx context.Context, conversationID string) (*encryption.VersionKey, error)
}
type EncryptionCacheRedis struct {
	metaCache
	expireTime time.Duration
	rdb        redis.UniversalClient
	rcClient   *rockscache.Client
}

func NewEncryptionCacheRedis(
	rdb redis.UniversalClient,
	options rockscache.Options,
) EncryptionCache {
	rcClient := rockscache.NewClient(rdb, options)
	return &EncryptionCacheRedis{
		metaCache: NewMetaCacheRedis(rcClient),
		rcClient:  rcClient,
		rdb:       rdb,
	}
}
func (e EncryptionCacheRedis) getConversationKey(conversationID string) string {
	return EncryptionKeyPrefix + conversationID
}
func (e EncryptionCacheRedis) GetConversationEncryptionAllKey(ctx context.Context, conversationID string) ([]*encryption.VersionKey, error) {
	result, err := e.rdb.HGetAll(ctx, e.getConversationKey(conversationID)).Result()
	if err != nil {
		return nil, err
	}
	var versionKeys []*encryption.VersionKey
	for version, key := range result {
		if version != MaxVersionKeyField {
			v, err := strconv.Atoi(version)
			if err != nil {
				continue
			}
			versionKeys = append(versionKeys, &encryption.VersionKey{
				Version: int32(v),
				Key:     key,
			})
		}
	}
	return versionKeys, nil
}

func (e EncryptionCacheRedis) GetConversationVersionKey(ctx context.Context, conversationID string, version int32) (string, error) {
	result, err := e.rdb.HGet(ctx, e.getConversationKey(conversationID), strconv.Itoa(int(version))).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (e EncryptionCacheRedis) GenConversationNextVersionKey(ctx context.Context, conversationID string,
	fn func() (string, error)) error {
	maxVersion, err := e.rdb.HGet(ctx, e.getConversationKey(conversationID), MaxVersionKeyField).Result()
	if err != nil {
		if err == redis.Nil {
			data := make(map[string]interface{})
			data[MaxVersionKeyField] = "1"
			key, err := fn()
			if err != nil {
				return err
			}
			data["1"] = key
			err = e.rdb.HMSet(ctx, e.getConversationKey(conversationID), data).Err()
			return err
		}
		return err
	}
	maxVersionInt, err := strconv.Atoi(maxVersion)
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	data[MaxVersionKeyField] = strconv.Itoa(maxVersionInt + 1)
	key, err := fn()
	if err != nil {
		return err
	}
	data[strconv.Itoa(maxVersionInt+1)] = key
	_, err = e.rdb.HMSet(ctx, e.getConversationKey(conversationID), data).Result()
	return err
}

func (e EncryptionCacheRedis) GetConversationMaxVersionKey(ctx context.Context, conversationID string) (*encryption.VersionKey, error) {
	maxVersion, err := e.rdb.HGet(ctx, e.getConversationKey(conversationID), MaxVersionKeyField).Result()
	if err != nil {
		return nil, err
	}
	key, err := e.rdb.HGet(ctx, e.getConversationKey(conversationID), maxVersion).Result()
	if err != nil {
		return nil, err
	}
	version, err := strconv.Atoi(maxVersion)
	if err != nil {
		return nil, err
	}
	return &encryption.VersionKey{
		Version: int32(version),
		Key:     key,
	}, nil
}
