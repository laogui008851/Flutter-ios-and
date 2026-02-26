package office

import (
	"context"
	"encoding/hex"
	"fmt"
	constant2 "github.com/OpenIMSDK/chat/pkg/common/constant"
	table "github.com/OpenIMSDK/chat/pkg/common/db/table/office"
	"github.com/OpenIMSDK/chat/pkg/common/mctx"
	"github.com/OpenIMSDK/chat/pkg/proto/common"
	"github.com/OpenIMSDK/chat/pkg/proto/office"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/google/uuid"
	"time"
)

func (o *officeServer) CreateOneWorkMoment(ctx context.Context, req *office.CreateOneWorkMomentReq) (*office.CreateOneWorkMomentResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	} else {
		if err := mctx.CheckAdminOr(ctx, req.UserID); err != nil {
			return nil, err
		}
	}
	switch req.Permission {
	case constant2.WorkMomentPublic:
	case constant2.WorkMomentPrivate:
	case constant2.WorkMomentPermissionCanSee:
	case constant2.WorkMomentPermissionCantSee:
	default:
		return nil, errs.ErrArgs.Wrap("permission not support")
	}
	var userIDs []string
	userIDs = append(userIDs, req.UserID)
	userIDs = append(userIDs, req.LikeUserIDs...)
	userIDs = append(userIDs, req.AtUserIDs...)
	userIDs = append(userIDs, req.PermissionUserIDs...)
	userIDs = utils.Distinct(userIDs)
	users, err := o.user.FindUserPublicInfo(ctx, userIDs)
	if err != nil {
		return nil, err
	} else if len(userIDs) != len(users) {
		return nil, errs.ErrUserIDNotFound.Wrap("part user id not found")
	}
	workMoment := table.WorkMoment{
		UserID: req.UserID,
		Content: &table.WorkMomentContent{
			Metas: make([]*table.Meta, 0, len(req.Content.Metas)),
			Text:  req.Content.Text,
			Type:  req.Content.Type,
		},
		AtUserIDs:  utils.Distinct(req.AtUserIDs),
		Permission: req.Permission,
		CreateTime: time.Now(),
	}
	for _, meta := range req.Content.Metas {
		if meta == nil {
			continue
		}
		workMoment.Content.Metas = append(workMoment.Content.Metas, &table.Meta{
			Original: meta.Original,
			Thumb:    meta.Thumb,
			Width:    meta.Width,
			Height:   meta.Height,
		})
	}
	switch req.Permission {
	case constant2.WorkMomentPublic:
		workMoment.PermissionUserIDs = []string{}
	case constant2.WorkMomentPrivate:
		workMoment.PermissionUserIDs = []string{}
	case constant2.WorkMomentPermissionCanSee:
		fallthrough
	case constant2.WorkMomentPermissionCantSee:
		workMoment.PermissionUserIDs = append(req.PermissionUserIDs, req.AtUserIDs...)
		workMoment.PermissionUserIDs = utils.Distinct(workMoment.PermissionUserIDs)
		if len(workMoment.PermissionUserIDs) == 0 {
			return nil, errs.ErrArgs.Wrap("permission user ids is empty")
		}
	}
	if err := o.db.CreateOneWorkMoment(ctx, &workMoment); err != nil {
		return nil, err
	}
	//o.WorkMomentSendNotificationNew(ctx, append(workMoment.AtUserIDs, workMoment.UserID), constant2.WorkMomentAtNotification, &workMoment)
	return &office.CreateOneWorkMomentResp{
		WorkMoment: o.DbToPbWorkMoment(&workMoment, utils.SliceToMap(users, func(user *common.UserPublicInfo) string {
			return user.UserID
		})),
	}, nil
}

func (o *officeServer) DeleteOneWorkMoment(ctx context.Context, req *office.DeleteOneWorkMomentReq) (*office.DeleteOneWorkMomentResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	workMoment, err := o.db.GetWorkMomentByID(ctx, req.WorkMomentID)
	if err != nil {
		return nil, err
	}
	if err := mctx.CheckAdminOr(ctx, workMoment.UserID); err != nil {
		return nil, errs.ErrNoPermission.Wrap("no permission to delete this work moment")
	}
	if err = o.db.DeleteOneWorkMoment(ctx, req.WorkMomentID); err != nil {
		return nil, err
	}
	return &office.DeleteOneWorkMomentResp{UserIDs: utils.Distinct(append(workMoment.AtUserIDs, workMoment.UserID, req.UserID))}, nil
}

func (o *officeServer) findWorkMomentAndCheckPermission(ctx context.Context, workMomentID string, userID string) (*table.WorkMoment, error) {
	workMoment, err := o.db.GetWorkMomentByID(ctx, workMomentID)
	if err != nil {
		return nil, err
	}
	if workMoment.UserID == userID {
		return workMoment, nil
	}
	switch workMoment.Permission {
	case constant2.WorkMomentPublic:
	case constant2.WorkMomentPrivate:
		return nil, errs.ErrNoPermission.Wrap("work moment is private")
	case constant2.WorkMomentPermissionCanSee:
		if !utils.Contain(userID, workMoment.PermissionUserIDs...) {
			return nil, errs.ErrNoPermission.Wrap("only allow whitelist to see")
		}
	case constant2.WorkMomentPermissionCantSee:
		if utils.Contain(userID, workMoment.PermissionUserIDs...) {
			return nil, errs.ErrNoPermission.Wrap("user id in blacklist")
		}
	default:
		return nil, errs.ErrInternalServer.Wrap(fmt.Sprintf("unknown work moment permission %d", workMoment.Permission))
	}
	return workMoment, nil
}

func (o *officeServer) LikeOneWorkMoment(ctx context.Context, req *office.LikeOneWorkMomentReq) (*office.LikeOneWorkMomentResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	workMoment, err := o.findWorkMomentAndCheckPermission(ctx, req.WorkMomentID, req.UserID)
	if err != nil {
		return nil, err
	}
	var like bool
	for _, likeUser := range workMoment.LikeUsers {
		if likeUser.UserID == req.UserID {
			like = true
			break
		}
	}
	if like == req.Like {
		return &office.LikeOneWorkMomentResp{}, nil
	}
	userIDSet := make(map[string]struct{})
	userIDSet[req.UserID] = struct{}{}
	userIDSet[workMoment.UserID] = struct{}{}
	for _, likeUser := range workMoment.LikeUsers {
		userIDSet[likeUser.UserID] = struct{}{}
	}
	for _, userID := range workMoment.PermissionUserIDs {
		userIDSet[userID] = struct{}{}
	}
	for _, comment := range workMoment.Comments {
		userIDSet[comment.UserID] = struct{}{}
		userIDSet[comment.ReplyUserID] = struct{}{}
	}
	delete(userIDSet, "")
	userMap, err := o.user.MapUserPublicInfo(ctx, utils.Keys(userIDSet))
	if err != nil {
		return nil, err
	}
	workMoment, err = o.db.SetWorkMomentLike(ctx, req.WorkMomentID, req.UserID, req.Like, time.Now())
	if err != nil {
		return nil, err
	}
	return &office.LikeOneWorkMomentResp{WorkMoment: o.DbToPbWorkMoment(workMoment, userMap)}, nil
}

func (o *officeServer) CommentOneWorkMoment(ctx context.Context, req *office.CommentOneWorkMomentReq) (*office.CommentOneWorkMomentResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	workMoment, err := o.findWorkMomentAndCheckPermission(ctx, req.WorkMomentID, req.UserID)
	if err != nil {
		return nil, err
	}
	if req.ReplyUserID != "" {
		var ok bool
		for _, comment := range workMoment.Comments {
			if comment.UserID == req.ReplyUserID {
				ok = true
				break
			}
		}
		if !ok {
			return nil, errs.ErrArgs.Wrap("reply user id not found")
		}
	}
	commentID := uuid.New()
	comment := &table.Comment{
		CommentID:   hex.EncodeToString(commentID[:]),
		UserID:      req.UserID,
		ReplyUserID: req.ReplyUserID,
		Content:     req.Content,
		CreateTime:  time.Now(),
	}
	workMoment, err = o.db.CommentOneWorkMoment(ctx, req.WorkMomentID, comment)
	if err != nil {
		return nil, err
	}
	userIDSet := make(map[string]struct{})
	userIDSet[req.UserID] = struct{}{}
	userIDSet[workMoment.UserID] = struct{}{}
	for _, likeUser := range workMoment.LikeUsers {
		userIDSet[likeUser.UserID] = struct{}{}
	}
	for _, userID := range workMoment.PermissionUserIDs {
		userIDSet[userID] = struct{}{}
	}
	for _, comment := range workMoment.Comments {
		userIDSet[comment.UserID] = struct{}{}
		userIDSet[comment.ReplyUserID] = struct{}{}
	}
	delete(userIDSet, "")
	userMap, err := o.user.MapUserPublicInfo(ctx, utils.Keys(userIDSet))
	if err != nil {
		return nil, err
	}
	return &office.CommentOneWorkMomentResp{
		CommentID:  comment.CommentID,
		WorkMoment: o.DbToPbWorkMoment(workMoment, userMap),
	}, nil
}

func (o *officeServer) DeleteComment(ctx context.Context, req *office.DeleteCommentReq) (*office.DeleteCommentResp, error) {
	userID := mctx.GetOpUserID(ctx)
	workMoment, err := o.findWorkMomentAndCheckPermission(ctx, req.WorkMomentID, userID)
	if err != nil {
		return nil, err
	}
	index := -1
	for i, comment := range workMoment.Comments {
		if comment.CommentID != req.CommentID {
			continue
		}
		index = i
		break
	}
	if index < 0 {
		return nil, errs.ErrRecordNotFound.Wrap("comment id not found")
	}
	comment := workMoment.Comments[index]
	if !(userID == comment.UserID || userID == workMoment.UserID || userID == comment.ReplyUserID) {
		return nil, errs.ErrNoPermission.Wrap("no permission delete comment")
	}
	workMoment.Comments = append(workMoment.Comments[:index], workMoment.Comments[index+1:]...)
	userIDSet := make(map[string]struct{})
	userIDSet[userID] = struct{}{}
	userIDSet[workMoment.UserID] = struct{}{}
	for _, likeUser := range workMoment.LikeUsers {
		userIDSet[likeUser.UserID] = struct{}{}
	}
	for _, userID := range workMoment.PermissionUserIDs {
		userIDSet[userID] = struct{}{}
	}
	for _, comment := range workMoment.Comments {
		userIDSet[comment.UserID] = struct{}{}
		userIDSet[comment.ReplyUserID] = struct{}{}
	}
	delete(userIDSet, "")
	userMap, err := o.user.MapUserPublicInfo(ctx, utils.Keys(userIDSet))
	if err != nil {
		return nil, err
	}
	if err := o.db.DeleteComment(ctx, req.WorkMomentID, req.CommentID); err != nil {
		return nil, err
	}
	return &office.DeleteCommentResp{
		UserIDs:    utils.Distinct(append(workMoment.AtUserIDs, workMoment.UserID, userID)),
		WorkMoment: o.DbToPbWorkMoment(workMoment, userMap),
	}, nil
}

func (o *officeServer) GetWorkMomentByID(ctx context.Context, req *office.GetWorkMomentByIDReq) (*office.GetWorkMomentByIDResp, error) {
	workMoment, err := o.findWorkMomentAndCheckPermission(ctx, req.WorkMomentID, mctx.GetOpUserID(ctx))
	if err != nil {
		return nil, err
	}
	pbWorkMoment, err := o.FillWorkMoment(ctx, workMoment)
	if err != nil {
		return nil, err
	}
	return &office.GetWorkMomentByIDResp{WorkMoment: pbWorkMoment}, nil
}

func (o *officeServer) GetUserSendWorkMoments(ctx context.Context, req *office.GetUserSendWorkMomentsReq) (*office.GetUserSendWorkMomentsResp, error) {
	opUserID, userType, err := mctx.Check(ctx)
	if err != nil {
		return nil, err
	}
	if userType == constant2.AdminUser {
		opUserID = req.UserID
	}
	workMoments, err := o.db.GetUserSendWorkMoments(ctx, req.UserID, opUserID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		return nil, err
	}
	pbWorkMoments, err := o.FillWorkMoments(ctx, workMoments)
	if err != nil {
		return nil, err
	}
	return &office.GetUserSendWorkMomentsResp{WorkMoments: pbWorkMoments}, nil
}

func (o *officeServer) GetUserRecvWorkMoments(ctx context.Context, req *office.GetUserRecvWorkMomentsReq) (*office.GetUserRecvWorkMomentsResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	workMoments, err := o.db.GetUserRecvWorkMoments(ctx, req.UserID, req.FriendIDs, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		return nil, err
	}
	pbWorkMoments, err := o.FillWorkMoments(ctx, workMoments)
	if err != nil {
		return nil, err
	}
	return &office.GetUserRecvWorkMomentsResp{WorkMoments: pbWorkMoments}, nil
}

func (o *officeServer) FindRelevantWorkMoments(ctx context.Context, req *office.FindRelevantWorkMomentsReq) (*office.FindRelevantWorkMomentsResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	read, err := o.db.GetUserReadTime(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	workMoments, err := o.db.FindRelevantWorkMoments(ctx, req.UserID, read.ListTime, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		return nil, err
	}
	pbWorkMoments, err := o.FillWorkMomentLogs(ctx, workMoments)
	if err != nil {
		return nil, err
	}
	return &office.FindRelevantWorkMomentsResp{WorkMoments: pbWorkMoments}, nil
}

func (o *officeServer) GetUnreadWorkMomentsCount(ctx context.Context, req *office.GetUnreadWorkMomentsCountReq) (*office.GetUnreadWorkMomentsCountResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	read, err := o.db.GetUserReadTime(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	log.ZInfo(ctx, "GetUnreadWorkMomentsCount GetUserReadTime", "user_id", req.UserID, "read", read)
	total, err := o.db.GetUnreadWorkMomentsCount(ctx, req.UserID, read.CountTime)
	if err != nil {
		return nil, err
	}
	log.ZInfo(ctx, "GetUnreadWorkMomentsCount GetUnreadWorkMomentsCount", "user_id", req.UserID, "read", read.CountTime, "total", total)
	return &office.GetUnreadWorkMomentsCountResp{Total: total}, nil
}

func (o *officeServer) ReadWorkMoments(ctx context.Context, req *office.ReadWorkMomentsReq) (*office.ReadWorkMomentsResp, error) {
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(ctx)
	}
	switch req.Type {
	case constant2.OfficeReadTypeCount:
	case constant2.OfficeReadTypeList:
	case constant2.OfficeReadTypeAll:
	default:
		return nil, errs.ErrArgs.Wrap("invalid read type")
	}
	if err := o.db.SetUserReadTime(ctx, req.UserID, time.Now(), req.Type); err != nil {
		return nil, err
	}
	return &office.ReadWorkMomentsResp{}, nil
}
