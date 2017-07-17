package handlers_test

import (
	"github.com/zanecloud/apiserver/handlers"
	"testing"
)

func TestContainerAudit(t *testing.T) {
	//pool := &handlers.PoolsRegisterRequest{
	//	Name:      "pool1",
	//	Driver:    "swarm",
	//	EnvTreeId: bson.NewObjectId().Hex(),
	//	DriverOpts: types.DriverOpts{
	//		Version:    "v1.0",
	//		EndPoint:   "tcp://47.92.49.245:2375",
	//		APIVersion: "v1.23",
	//	},
	//}
	//
	//poolRsp := &handlers.PoolsRegisterResponse{}
	//
	//postTestRequest("pools/register", pool, poolRsp)
	//
	//app := &handlers.ApplicationCreateRequest{
	//	ApplicationTemplateId: bson.NewObjectId().Hex(),
	//	PoolId:                poolRsp.Id,
	//	Title:                 "ContainerTest",
	//	Description:           "ContainerTestAppDesc",
	//}
	//
	//appRsp := &handlers.ApplicationCreateResponse{}
	//
	//postTestRequest("applications/create", app, appRsp)
	//
	//caRsp := &handlers.CreateSSHSessionResponse{}
	//
	////用户未对pool授权
	////访问失败
	//t.Run("CA=1", func(t *testing.T) {
	//	if err := getTestRequest("audit/ssh", caRsp); err != nil {
	//		if err != mgo.ErrNotFound {
	//			t.Error("Current user could not permit for this pool!")
	//		} else {
	//			t.Log()
	//		}
	//	} else {
	//		t.Error(caRsp)
	//	}
	//})
	//
	////授权
	//addUserPoolId(poolRsp.Id, user.Id)

	//用户对pool授权
	//访问成功
	//t.Run("CA=2", func(t *testing.T) {
	//	if err := getTestRequest("audit/ssh", caRsp); err != nil {
	//		t.Error(err)
	//	} else if len(caRsp.Token) <= 0 {
	//		t.Error(caRsp)
	//	} else {
	//		t.Log(caRsp)
	//	}
	//})
	//
	////获得用户授权的token
	//token := caRsp.Token
	//
	////验证token格式的合法性
	////ssh -p %s %s@%s
	//t.Run("CA=3", func(t *testing.T) {
	//
	//})
}

func TestJSONUnmarshal(t *testing.T) {
	t.Run("JSON=1", func(t *testing.T) {
		req := handlers.GetAuditListRequest{
			StartTime: 1400113570,
			EndTime:   1500113570,
		}

		rsp := handlers.GetAuditListResponse{}

		if err := postTestRequest("audit/list", req, &rsp); err != nil {
			t.Error(err.Error())
		} else {
			t.Log(rsp)
		}
	})
}

/*
	测试数据
*/

//func updatePoolInfo(pid string) {
//	//TODO
//	//
//	s := mgoSession.Clone()
//	container := swarm.Container{
//		Id:   bson.NewObjectId().Hex(),
//		Name: "Container01",
//	}
//	s.DB("zanecloud").C("pool").Insert()
//
//	cmd := "mongo"
//	args := fmt.Sprintf("zanecloud --eval \"db.pool.update({_id:ObjectId('%s')}, {TunneldAddr: '192.168.10.100', TunneldPort: '28080'})\"", pid)
//	_, err := exec.Command("sh", "-c", fmt.Sprintf("%s %s", cmd, args)).Output()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//}
