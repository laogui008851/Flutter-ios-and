package office

import (
	"github.com/OpenIMSDK/chat/pkg/common/config"
	"github.com/OpenIMSDK/chat/pkg/common/db/database"
	"github.com/OpenIMSDK/chat/pkg/common/dbconn"
	"github.com/OpenIMSDK/chat/pkg/proto/office"
	"github.com/OpenIMSDK/chat/pkg/rpclient/chat"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"google.golang.org/grpc"
)

func Start(zk discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error {
	mgo, err := dbconn.NewMongo()
	if err != nil {
		return err
	}
	db, err := database.NewOfficeDatabase(mgo.Database(*config.Config.Mongo.Database))
	if err != nil {
		return err
	}
	office.RegisterOfficeServer(server, &officeServer{
		db:   db,
		user: chat.NewChatClient(zk),
	})
	return nil
}

type officeServer struct {
    office.UnimplementedOfficeServer
	db   database.OfficeDatabaseInterface
	user *chat.ChatClient
}
