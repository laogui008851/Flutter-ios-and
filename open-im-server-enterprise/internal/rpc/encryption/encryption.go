package encryption

import (
	"context"

	"github.com/OpenIMSDK/protocol/encryption"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"google.golang.org/grpc"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"
)

type encryptionServer struct {
	conversationDatabase controller.EncryptionDatabase
}

func Start(_ discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error {
	rdb, err := cache.NewRedis()
	if err != nil {
		return err
	}
	encryption.RegisterEncryptionServer(server, &encryptionServer{
		conversationDatabase: controller.NewEncryptionDatabase(cache.NewEncryptionCacheRedis(rdb, cache.GetDefaultOpt())),
	})
	return nil
}

func (e *encryptionServer) GetEncryptionKey(ctx context.Context, req *encryption.GetEncryptionKeyReq) (resp *encryption.GetEncryptionKeyResp, err error) {
	if req.KeyVersion == 0 {
		keyList, err := e.conversationDatabase.GetConversationEncryptionAllKey(ctx, req.ConversationID)
		if err != nil {
			return nil, err
		}
		resp := &encryption.GetEncryptionKeyResp{VersionKeyList: keyList}
		return resp, nil
	} else {
		key, err := e.conversationDatabase.GetConversationVersionKey(ctx, req.ConversationID, req.KeyVersion)
		if err != nil {
			return nil, err
		}
		resp := &encryption.GetEncryptionKeyResp{VersionKeyList: []*encryption.VersionKey{{
			Version: req.KeyVersion,
			Key:     key,
		}}}
		return resp, nil
	}
}
func (e *encryptionServer) GenEncryptionKey(ctx context.Context, req *encryption.GenEncryptionKeyReq) (*encryption.GenEncryptionKeyResp, error) {
	err := e.conversationDatabase.GenConversationNextVersionKey(ctx, req.ConversationID, generateAESKey)
	if err != nil {
		return nil, err
	}
	return &encryption.GenEncryptionKeyResp{}, nil
}
