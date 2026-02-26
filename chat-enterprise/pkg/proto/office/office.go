package office

import "github.com/OpenIMSDK/tools/errs"

func (x *CreateOneWorkMomentReq) Check() error {
	if x.Content == nil {
		return errs.ErrArgs.Wrap("content is nil")
	}
	return nil
}

func (x *FindRelevantWorkMomentsReq) Check() error {
	if x.Pagination == nil {
		return errs.ErrArgs.Wrap("pagination is nil")
	}
	if x.Pagination.PageNumber < 1 {
		return errs.ErrArgs.Wrap("page number is invalid")
	}
	return nil
}
