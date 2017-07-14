package application

import (
	"context"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func AddDeploymentLog(ctx context.Context, app *types.Application, pool *types.PoolInfo, user *types.User, operation string, opts *types.DeploymentOpts) error {

	deployment := &types.Deployment{
		Id:                 bson.NewObjectId(),
		OperationType:      operation,
		ApplicationId:      app.Id.Hex(),
		ApplicationVersion: app.Version,
		PoolId:             pool.Id.Hex(),
		CreatedTime:        time.Now().Unix(),
		CreatorId:          user.Id.Hex(),
	}

	cb := func(cs map[string]*mgo.Collection) error {
		return nil
	}

	return utils.ExecMgoCollections(ctx, []string{"deployment"}, cb)

}
