package api

import (
	"github.com/OpenIMSDK/chat/pkg/proto/rtc"
	"github.com/OpenIMSDK/tools/a2r"
	"github.com/gin-gonic/gin"
)

func (o *AdminApi) DeleteSignalRecords(c *gin.Context) {
	a2r.Call(rtc.RtcServiceClient.DeleteSignalRecords, o.rtcClient, c)
}

func (o *AdminApi) GetSignalInvitationRecords(c *gin.Context) {
	a2r.Call(rtc.RtcServiceClient.GetSignalInvitationRecords, o.rtcClient, c)
}

func (o *AdminApi) GetMeetingRecords(c *gin.Context) {
	a2r.Call(rtc.RtcServiceClient.GetMeetingRecords, o.rtcClient, c)
}

func (o *AdminApi) DeleteMeetingRecords(c *gin.Context) {
	a2r.Call(rtc.RtcServiceClient.DeleteMeetingRecords, o.rtcClient, c)
}
