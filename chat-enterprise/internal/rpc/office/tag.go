package office

import (
	"context"
	"fmt"
	table "github.com/OpenIMSDK/chat/pkg/common/db/table/office"
	"github.com/OpenIMSDK/chat/pkg/common/mctx"
	"github.com/OpenIMSDK/chat/pkg/proto/common"
	"github.com/OpenIMSDK/chat/pkg/proto/office"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"
	"time"
)

func (o *officeServer) GetUserTags(ctx context.Context, req *office.GetUserTagsReq) (*office.GetUserTagsResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	} else {
		if err := mctx.CheckAdminOr(ctx, req.UserID); err != nil {
			return nil, err
		}
	}
	tags, err := o.db.GetUserTags(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	var (
		userIDs []string
	)
	for _, tag := range tags {
		userIDs = append(userIDs, tag.UserIDs...)
	}
	var userMap map[string]*common.UserPublicInfo
	if len(userIDs) > 0 {
		userMap, err = o.user.MapUserPublicInfo(ctx, utils.Distinct(userIDs))
		if err != nil {
			return nil, err
		}
	}
	pbTags := make([]*office.Tag, 0, len(tags))
	for _, t := range tags {
		tag := &office.Tag{
			TagID:   t.TagID.Hex(),
			TagName: t.TagName,
		}
		for _, userID := range t.UserIDs {
			if user := userMap[userID]; user != nil {
				tag.Users = append(tag.Users, user)
			}
		}
		pbTags = append(pbTags, tag)
	}
	return &office.GetUserTagsResp{Tags: pbTags}, nil
}

func (o *officeServer) CreateTag(ctx context.Context, req *office.CreateTagReq) (*office.CreateTagResp, error) {
	if utils.Duplicate(req.UserIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate user id")
	}
	if utils.Duplicate(req.GroupIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate group id")
	}
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	} else {
		if err := mctx.CheckAdminOr(ctx, req.UserID); err != nil {
			return nil, err
		}
	}
	if len(req.UserIDs)+len(req.GroupIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("user id and group id is empty")
	}
	if len(req.UserIDs) > 0 {
		if infos, err := o.user.FindUserPublicInfo(ctx, req.UserIDs); err != nil {
			return nil, err
		} else if len(infos) != len(req.UserIDs) {
			return nil, errs.ErrUserIDNotFound.Wrap("user id not found")
		}
	}
	if err := o.db.CreateTag(ctx, &table.Tag{
		UserID:     req.UserID,
		TagName:    req.TagName,
		UserIDs:    req.UserIDs,
		GroupIDs:   req.GroupIDs,
		CreateTime: time.Now(),
	}); err != nil {
		return nil, err
	}
	return &office.CreateTagResp{}, nil
}

func (o *officeServer) DeleteTag(ctx context.Context, req *office.DeleteTagReq) (*office.DeleteTagResp, error) {
	tag, err := o.db.GetTagByID(ctx, req.TagID)
	if err != nil {
		return nil, err
	}
	if err := mctx.CheckAdminOr(ctx, tag.UserID); err != nil {
		return nil, err
	}
	if err := o.db.DeleteTag(ctx, req.TagID); err != nil {
		return nil, err
	}
	return &office.DeleteTagResp{}, nil
}

func (o *officeServer) SetTag(ctx context.Context, req *office.SetTagReq) (*office.SetTagResp, error) {
	if len(req.Name)+len(req.AddUserIDs)+len(req.DelUserIDs)+len(req.AddGroupIDs)+len(req.DelGroupIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("no change field")
	}
	if utils.Duplicate(req.AddUserIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate add user id")
	}
	if utils.Duplicate(req.DelUserIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate del user id")
	}
	if utils.Duplicate(req.AddGroupIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate add group id")
	}
	if utils.Duplicate(req.DelGroupIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate del group id")
	}
	tag, err := o.db.GetTagByID(ctx, req.TagID)
	if err != nil {
		return nil, err
	}
	if err := mctx.CheckAdminOr(ctx, tag.UserID); err != nil {
		return nil, err
	}
	if len(req.AddUserIDs) > 0 {
		if infos, err := o.user.FindUserPublicInfo(ctx, req.AddUserIDs); err != nil {
			return nil, err
		} else if len(infos) != len(req.AddUserIDs) {
			return nil, errs.ErrUserIDNotFound.Wrap("user id not found")
		}
	}
	if err := o.db.SetTag(ctx, req.TagID, req.Name, req.AddUserIDs, req.DelUserIDs, req.AddGroupIDs, req.DelGroupIDs); err != nil {
		return nil, err
	}
	return &office.SetTagResp{}, nil
}

func (o *officeServer) SendMsg2Tag(ctx context.Context, req *office.SendMsg2TagReq) (*office.SendMsg2TagResp, error) {
	if utils.Duplicate(req.TagIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate tag id")
	}
	if utils.Duplicate(req.UserIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate user id")
	}
	if utils.Duplicate(req.GroupIDs) {
		return nil, errs.ErrArgs.Wrap("duplicate group id")
	}
	if req.SendID == "" {
		req.SendID = mctx.GetOpUserID(ctx)
	} else {
		if err := mctx.CheckAdminOr(ctx, req.SendID); err != nil {
			return nil, err
		}
	}
	var sendUser *common.UserPublicInfo
	reqUserIDs := utils.Distinct(append(req.UserIDs, req.SendID))
	infos, err := o.user.FindUserPublicInfo(ctx, reqUserIDs)
	if err != nil {
		return nil, err
	}
	if len(infos) != len(reqUserIDs) {
		return nil, errs.ErrUserIDNotFound.Wrap("user id not found")
	}
	for i, info := range infos {
		if info.UserID == req.SendID {
			sendUser = infos[i]
			break
		}
	}
	if sendUser == nil {
		return nil, errs.ErrUserIDNotFound.Wrap(fmt.Sprintf("send user id %s not found", sendUser.UserID))
	}
	var tags []*table.Tag
	if len(req.TagIDs) > 0 {
		tags, err = o.db.FindTag(ctx, req.TagIDs)
		if err != nil {
			return nil, err
		}
	}
	userIDs := req.UserIDs
	groupIDs := req.GroupIDs
	for _, tag := range tags {
		userIDs = append(userIDs, tag.UserIDs...)
		groupIDs = append(groupIDs, tag.GroupIDs...)
	}
	if len(userIDs)+len(groupIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("no recv user")
	}
	tagSendLogs := table.TagSendLog{
		SendUserID: req.SendID,
		PlatformID: req.SenderPlatformID,
		UserIDs:    req.UserIDs,
		TagIDs:     req.TagIDs,
		GroupIDs:   req.GroupIDs,
		Content:    req.Content,
		SendTime:   time.Now(),
	}
	if err := o.db.CreateTagSendLog(ctx, &tagSendLogs); err != nil {
		return nil, err
	}
	return &office.SendMsg2TagResp{
		SendUser:     sendUser,
		RecvGroupIDs: utils.Distinct(groupIDs),
		RecvUserIDs:  utils.Distinct(userIDs),
		Content:      req.Content,
	}, nil
}

func (o *officeServer) GetTagSendLogs(ctx context.Context, req *office.GetTagSendLogsReq) (*office.GetTagSendLogsResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	} else {
		if err := mctx.CheckAdminOr(ctx, req.UserID); err != nil {
			return nil, err
		}
	}
	tagSendLogs, err := o.db.GetTagSendLogs(ctx, req.UserID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		return nil, err
	}
	var (
		tagIDs  []string
		userIDs []string
	)
	for _, tsg := range tagSendLogs {
		tagIDs = append(tagIDs, tsg.TagIDs...)
		userIDs = append(userIDs, tsg.UserIDs...)
	}
	var userMap map[string]*common.UserPublicInfo
	if len(userIDs) > 0 {
		userMap, err = o.user.MapUserPublicInfo(ctx, utils.Distinct(userIDs))
		if err != nil {
			return nil, err
		}
	}
	tagMap := make(map[string]*table.Tag)
	if len(tagIDs) > 0 {
		tags, err := o.db.FindTag(ctx, tagIDs)
		if err != nil {
			return nil, err
		}
		for i, tag := range tags {
			tagMap[tag.TagID.Hex()] = tags[i]
		}
	}
	pbTagSendLogs := make([]*office.TagSendLog, 0, len(tagSendLogs))
	for _, tsl := range tagSendLogs {
		pbTsl := &office.TagSendLog{
			Id:       tsl.ID.Hex(),
			Users:    make([]*common.UserPublicInfo, 0, len(tsl.UserIDs)),
			GroupIDs: tsl.GroupIDs,
			Content:  tsl.Content,
			SendTime: tsl.SendTime.UnixMilli(),
		}
		for _, tagID := range tsl.TagIDs {
			if tag := tagMap[tagID]; tag != nil {
				pbTsl.Tags = append(pbTsl.Tags, &office.Tag{
					TagID:      tag.TagID.Hex(),
					TagName:    tag.TagName,
					CreateTime: tag.CreateTime.UnixMilli(),
				})
			}
		}
		for _, userID := range tsl.UserIDs {
			if user := userMap[userID]; user != nil {
				pbTsl.Users = append(pbTsl.Users, user)
			}
		}
		pbTagSendLogs = append(pbTagSendLogs, pbTsl)
	}
	return &office.GetTagSendLogsResp{
		TagSendLogs: pbTagSendLogs,
	}, nil
}

func (o *officeServer) DelTagSendLog(ctx context.Context, req *office.DelTagSendLogReq) (*office.DelTagSendLogResp, error) {
	if len(req.Ids) == 0 {
		return nil, errs.ErrArgs.Wrap("no ids to delete")
	}
	if utils.Duplicate(req.Ids) {
		return nil, errs.ErrArgs.Wrap("duplicate ids")
	}
	res, err := o.db.GetTagLogSendUser(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	if len(req.Ids) != len(res) {
		return nil, errs.ErrRecordNotFound.Wrap("some ids not found")
	}
	for _, sendUserID := range res {
		if err := mctx.CheckAdminOrIn(ctx, sendUserID); err != nil {
			return nil, err
		}
	}
	if err := o.db.DeleteTagLog(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &office.DelTagSendLogResp{}, nil
}

func (o *officeServer) GetUserTagByID(ctx context.Context, req *office.GetUserTagByIDReq) (*office.GetUserTagByIDResp, error) {
	tag, err := o.db.GetTagByID(ctx, req.TagID)
	if err != nil {
		return nil, err
	}
	if err := mctx.CheckAdminOr(ctx, tag.UserID); err != nil {
		return nil, err
	}
	pbTag := &office.Tag{
		TagID:      tag.TagID.Hex(),
		TagName:    tag.TagName,
		CreateTime: tag.CreateTime.UnixMilli(),
	}
	if len(tag.UserIDs) > 0 {
		pbTag.Users, err = o.user.FindUserPublicInfo(ctx, tag.UserIDs)
		if err != nil {
			return nil, err
		}
	}
	return &office.GetUserTagByIDResp{Tag: pbTag}, nil
}
