package controller

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/unrelation"
)

type MsgGroupReadDatabase interface {
	GetMsgNum(ctx context.Context, conversationID string, clientMsgID string) (readNum int64, unreadNum int64, err error)
	MarkGroupMessageRead(ctx context.Context, conversationID string, clientMsgID string, userID string, readTime *time.Time, getUserIDs cache.GetUserIDsFunc) (int32, error)
	PageReadUserList(ctx context.Context, conversationID string, clientMsgID string, read bool, pageNumber int32, showNumber int32) (int64, []*cache.UserReadTime, error)
}

func NewMsgGroupReadDatabase(rdb redis.UniversalClient, msg unrelation.GroupMsgReadInterface) MsgGroupReadDatabase {
	return &msgGroupReadDatabase{
		groupReadCache: cache.NewMsgGroupReadCache(rdb, msg),
	}
}

type msgGroupReadDatabase struct {
	groupReadCache cache.MsgGroupReadCache
}

func (db *msgGroupReadDatabase) GetMsgNum(ctx context.Context, conversationID string, clientMsgID string) (readNum int64, unreadNum int64, err error) {
	return db.groupReadCache.GetMsgNum(ctx, conversationID, clientMsgID)
}

func (db *msgGroupReadDatabase) MarkGroupMessageRead(
	ctx context.Context,
	conversationID string,
	clientMsgID string,
	userID string,
	readTime *time.Time,
	getUserIDs cache.GetUserIDsFunc,
) (int32, error) {
	return db.groupReadCache.MarkGroupMessageRead(ctx, conversationID, clientMsgID, userID, readTime, getUserIDs)
}

func (db *msgGroupReadDatabase) PageReadUserList(ctx context.Context, conversationID string, clientMsgID string, read bool, pageNumber int32, showNumber int32) (int64, []*cache.UserReadTime, error) {
	return db.groupReadCache.PageReadUserList(ctx, conversationID, clientMsgID, read, pageNumber, showNumber)
}
