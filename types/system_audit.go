package types

import (
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
	"time"
)

/*
通用字段：
	时间，IP，用户，模块，操作
	CreatedTime, IP, UserId, PoolId, ApplicationId
	说明：
		UserId在登录失败的时候为空；
		PoolId在某些不涉及到Pool的操作时候为空；
ApplicationId在某些不涉及到Application的操作时候为空；
*/

type SystemAuditLog struct {
	Id            bson.ObjectId "_id"
	IP            string
	UserId        bson.ObjectId                  //当前操作者
	Module        SystemAuditModuleType          //被操作模块名称，跟前端约定：User，Env系列，Team等等
	Operation     SystemAuditModuleOperationType //操作类型，跟前端约定：Create，Update，UpdateServiceReplicaCount，UpdatePoolValue等等
	Detail        interface{}                    //详细操作内容
	PoolId        bson.ObjectId                  `bson:",omitempty"` //被操作集群，可为空
	ApplicationId bson.ObjectId                  `bson:",omitempty"` //被操作应用，可为空
	RequestURI    string                         //当前操作发生的API地址

	CreatedTime int64
}

type SystemAuditModuleType int

const (
	SystemAuditModuleTypeUser SystemAuditModuleType = 1 + iota
	SystemAuditModuleTypeTeam
	SystemAuditModuleTypePool
	SystemAuditModuleTypeEnv
	SystemAuditModuleTypeApplicationTemplate
	SystemAuditModuleTypeApplication
)

type SystemAuditModuleOperationType int

const (
	SystemAuditModuleOperationTypeUserCreate SystemAuditModuleOperationType = 1 + iota
	SystemAuditModuleOperationTypeUserUpdate
	SystemAuditModuleOperationTypeUserDelete
	SystemAuditModuleOperationTypeUserLoginFailed
	SystemAuditModuleOperationTypeUserLogined

	SystemAuditModuleOperationTypeTeamCreate
	SystemAuditModuleOperationTypeTeamUpdate
	SystemAuditModuleOperationTypeTeamDelete
	SystemAuditModuleOperationTypeTeamAddUser
	SystemAuditModuleOperationTypeTeamRemoveUser

	SystemAuditModuleOperationTypePoolCreate
	SystemAuditModuleOperationTypePoolUpdate
	SystemAuditModuleOperationTypePoolDelete
	SystemAuditModuleOperationTypePoolAuthTeam
	SystemAuditModuleOperationTypePoolRevokeTeam
	SystemAuditModuleOperationTypePoolAuthUser
	SystemAuditModuleOperationTypePoolRevokeUser

	SystemAuditModuleOperationTypeEnvUpdateEnvValue
	SystemAuditModuleOperationTypeEnvUpdatePoolValue

	SystemAuditModuleOperationTypeApplicatonTemplateCreate
	SystemAuditModuleOperationTypeApplicatonTemplateUpdate
	SystemAuditModuleOperationTypeApplicatonTemplateDelete

	SystemAuditModuleOperationTypeApplicatonCreate
	SystemAuditModuleOperationTypeApplicatonUpdate
	SystemAuditModuleOperationTypeApplicatonUpdateServiceReplicaCount
	SystemAuditModuleOperationTypeApplicatonRestartContainer
	SystemAuditModuleOperationTypeApplicatonUpgrade
	SystemAuditModuleOperationTypeApplicatonRollback
	SystemAuditModuleOperationTypeApplicatonAuthTeam
	SystemAuditModuleOperationTypeApplicatonRevokeTeam
	SystemAuditModuleOperationTypeApplicatonAuthUser
	SystemAuditModuleOperationTypeApplicatonRevokeUser
)

//生成系统审计日志
func CreateSystemAuditLog(db *mgo.Database, r *http.Request, userId string, module SystemAuditModuleType, operation SystemAuditModuleOperationType, poolId string, applicationId string, detail interface{}) (err error) {
	c := db.C("system_audit_log")

	ip := getIpFromReuqest(r)

	//校验参数合法性
	if ip == "" {
		return errors.New("IP could not be empty")
	}

	if userId == "" {
		return errors.New("userId could not be empty")
	}

	if module == 0 {
		return errors.New("module could not be empty")
	}

	if operation == 0 {
		return errors.New("operation could not be empty")
	}

	log := SystemAuditLog{
		Id:          bson.NewObjectId(),
		IP:          ip,
		UserId:      bson.ObjectIdHex(userId),
		Module:      module,
		Operation:   operation,
		Detail:      detail,
		RequestURI:  r.RequestURI,
		CreatedTime: time.Now().Unix(),
	}

	if poolId != "" {
		log.PoolId = bson.ObjectIdHex(poolId)
	}

	if applicationId != "" {
		log.ApplicationId = bson.ObjectIdHex(applicationId)
	}

	if err := c.Insert(log); err != nil {
		return err
	}

	return nil
}

//从HTTP请求中获取客户端IP
func getIpFromReuqest(r *http.Request) string {
	var ip string

	xff := r.Header.Get("X-Forward-For")
	if xff != "" {
		forwards := strings.Split(xff, ",")
		ip = forwards[0]
	} else {
		ip = r.Header.Get("Remote_addr")
		if ip == "" {
			ip = r.RemoteAddr
		}
	}

	return ip
}
