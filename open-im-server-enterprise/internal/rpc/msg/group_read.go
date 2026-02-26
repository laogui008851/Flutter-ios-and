package msg

import (
	"context"
	"strings"
	"time"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"

	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
)

func (m *msgServer) markGroupMessageRead(ctx context.Context, conversationID, userID string, groupID string, mark map[string]int32, readTime int64) {
	users := []*sdkws.GroupMsgReadTipsUser{{UserID: userID, ReadTime: readTime}}
	tips := sdkws.GroupMsgReadTips{Reads: make([]*sdkws.GroupMsgReadTipsRead, 0, len(mark))}
	for clientID, ok := range mark {
		if ok == constant.GroupMessageReadMarkNew {
			readNum, unreadNum, err := m.MsgGroupReadDatabase.GetMsgNum(ctx, conversationID, clientID)
			if err != nil {
				log.ZError(ctx, "markGroupMessageRead GetMsgReadNum failed", err, "conversationID", conversationID, "clientID", clientID)
			}
			tips.Reads = append(tips.Reads, &sdkws.GroupMsgReadTipsRead{
				ConversationID: conversationID,
				ClientMsgID:    clientID,
				ReadNum:        readNum,
				UnreadNum:      unreadNum,
				Users:          users,
			})
		}
	}
	if len(tips.Reads) == 0 {
		return
	}
	_ = m.notificationSender.NotificationWithSesstionType(ctx, userID, groupID, constant.HasGroupReadReceipt, constant.SuperGroupChatType, &tips)
}

func (m *msgServer) MarkGroupMessageRead(ctx context.Context, req *msg.MarkGroupMessageReadReq) (*msg.MarkGroupMessageReadResp, error) {
	if len(req.ClientMsgs) == 0 {
		return nil, errs.ErrArgs.Wrap("clientMsgs is empty")
	}
	for id := range req.ClientMsgs {
		if id == "" {
			return nil, errs.ErrArgs.Wrap("clientMsgID is empty")
		}
	}
	if !strings.HasPrefix(req.ConversationID, "sg_") {
		return nil, errs.ErrArgs.Wrap("conversationID is invalid")
	}
	groupID := strings.TrimPrefix(req.ConversationID, "sg_")
	if groupID == "" {
		return nil, errs.ErrArgs.Wrap("conversationID is invalid")
	}
	if err := authverify.CheckAccessV3(ctx, req.UserID); err != nil {
		return nil, err
	}
	resp := &msg.MarkGroupMessageReadResp{Mark: make(map[string]int32)}
	now := time.Now()
	var userIDs []string
	for clientMsgID, sendUserID := range req.ClientMsgs {
		state, err := m.MsgGroupReadDatabase.MarkGroupMessageRead(ctx, req.ConversationID, clientMsgID, req.UserID, &now, func(ctx context.Context) ([]string, error) {
			if len(userIDs) == 0 {
				var err error
				userIDs, err = m.Group.GetGroupMemberIDs(ctx, groupID)
				if err != nil {
					return nil, err
				}
				if len(userIDs) == 0 {
					return nil, errs.ErrRecordNotFound.Wrap("group member is empty")
				}
			}
			if sendUserID != "" {
				for i, userID := range userIDs {
					if userID == sendUserID {
						return append(userIDs[:i], userIDs[i+1:]...), nil
					}
				}
			}
			return userIDs, nil
		})
		if err != nil {
			return nil, err
		}
		resp.Mark[clientMsgID] = state
	}
	m.markGroupMessageRead(ctx, req.ConversationID, req.UserID, groupID, resp.Mark, now.UnixMilli())
	return resp, nil
}

func (m *msgServer) GetGroupMessageReadNum(ctx context.Context, req *msg.GetGroupMessageReadNumReq) (*msg.GetGroupMessageReadNumResp, error) {
	if len(req.ClientMsgIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("clientMsgIDs is empty")
	}
	resp := &msg.GetGroupMessageReadNumResp{Num: make(map[string]*msg.MessageReadInfo)}
	for _, clientMsgID := range req.ClientMsgIDs {
		readNum, unreadNum, err := m.MsgGroupReadDatabase.GetMsgNum(ctx, req.ConversationID, clientMsgID)
		if err != nil {
			return nil, err
		}
		resp.Num[clientMsgID] = &msg.MessageReadInfo{
			ReadNum:   readNum,
			UnreadNum: unreadNum,
		}
	}
	return resp, nil
}

func (m *msgServer) GetGroupMessageHasRead(ctx context.Context, req *msg.GetGroupMessageHasReadReq) (*msg.GetGroupMessageHasReadResp, error) {
	if req.Pagination == nil || req.Pagination.PageNumber < 1 || req.Pagination.ShowNumber < 1 {
		return nil, errs.ErrArgs.Wrap("pagination is invalid")
	}
	var read bool
	switch req.Type {
	case constant.GroupMessageReadList:
		read = true
	case constant.GroupMessageUnreadList:
		read = false
	default:
		return nil, errs.ErrArgs.Wrap("type is invalid")
	}
	count, res, err := m.MsgGroupReadDatabase.PageReadUserList(ctx, req.ConversationID, req.ClientMsgID, read, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	resp := &msg.GetGroupMessageHasReadResp{
		Count: count,
		Reads: make([]*msg.GroupMessageUserReadTime, 0, len(res)),
	}
	for _, re := range res {
		resp.Reads = append(resp.Reads, &msg.GroupMessageUserReadTime{
			UserID:   re.UserID,
			ReadTime: re.Time,
		})
	}
	return resp, nil
}
