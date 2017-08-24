package types

import (
	"gopkg.in/mgo.v2/bson"
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

type SystemAuditModuleType string

const (
	SystemAuditModuleTypeUser                SystemAuditModuleType = "User"
	SystemAuditModuleTypeTeam                SystemAuditModuleType = "Team"
	SystemAuditModuleTypePool                SystemAuditModuleType = "Pool"
	SystemAuditModuleTypeEnv                 SystemAuditModuleType = "Env"
	SystemAuditModuleTypeApplicationTemplate SystemAuditModuleType = "ApplicationTemplate"
	SystemAuditModuleTypeApplication         SystemAuditModuleType = "Application"
)

type SystemAuditModuleOperationType string

const (
	SystemAuditModuleOperationTypeCreate                    SystemAuditModuleOperationType = "Create"                    //用户，模板，应用
	SystemAuditModuleOperationTypeUpdate                    SystemAuditModuleOperationType = "Update"                    //用户，模板，应用
	SystemAuditModuleOperationTypeDelete                    SystemAuditModuleOperationType = "Delete"                    //用户，模板，应用
	SystemAuditModuleOperationTypeLoginFailed               SystemAuditModuleOperationType = "LoginFailed"               //用户
	SystemAuditModuleOperationTypeLogined                   SystemAuditModuleOperationType = "Logined"                   //用户
	SystemAuditModuleOperationTypeAddUser                   SystemAuditModuleOperationType = "AddUser"                   //团队
	SystemAuditModuleOperationTypeRemoveUser                SystemAuditModuleOperationType = "RemoveUser"                //团队
	SystemAuditModuleOperationTypeAuthTeam                  SystemAuditModuleOperationType = "AuthTeam"                  //集群，应用
	SystemAuditModuleOperationTypeRevokeTeam                SystemAuditModuleOperationType = "RevokeTeam"                //集群，应用
	SystemAuditModuleOperationTypeAuthUser                  SystemAuditModuleOperationType = "AuthUser"                  //集群，应用
	SystemAuditModuleOperationTypeRevokeUser                SystemAuditModuleOperationType = "RevokeUser"                //集群，应用
	SystemAuditModuleOperationTypeUpdateEnvValue            SystemAuditModuleOperationType = "UpdateEnvValue"            //Env
	SystemAuditModuleOperationTypeUpdatePoolValue           SystemAuditModuleOperationType = "UpdatePoolValue"           //Env
	SystemAuditModuleOperationTypeUpdateServiceReplicaCount SystemAuditModuleOperationType = "UpdateServiceReplicaCount" //应用
	SystemAuditModuleOperationTypeRestartContainer          SystemAuditModuleOperationType = "RestartContainer"          //应用
	SystemAuditModuleOperationTypeUpgrade                   SystemAuditModuleOperationType = "Upgrade"                   //应用
	SystemAuditModuleOperationTypeRollback                  SystemAuditModuleOperationType = "Rollback"                  //应用
	SystemAuditModuleOperationTypeStop                      SystemAuditModuleOperationType = "Stop"                      //应用
	SystemAuditModuleOperationTypeStart                     SystemAuditModuleOperationType = "Start"                     //应用
)

type SystemAuditModuleEnvUpdatePoolValueItem struct {
	EnvValue map[string]string
	Pool     map[string]string
	OldValue *EnvTreeNodeParamValue
	NewValue *EnvTreeNodeParamValue
	ValueId  bson.ObjectId
}

//const (
//	SystemAuditModuleOperationTypeUserCreate SystemAuditModuleOperationType = 1 + iota
//	SystemAuditModuleOperationTypeUserUpdate
//	SystemAuditModuleOperationTypeUserDelete
//	SystemAuditModuleOperationTypeUserLoginFailed
//	SystemAuditModuleOperationTypeUserLogined
//
//	SystemAuditModuleOperationTypeTeamCreate
//	SystemAuditModuleOperationTypeTeamUpdate
//	SystemAuditModuleOperationTypeTeamDelete
//	SystemAuditModuleOperationTypeTeamAddUser
//	SystemAuditModuleOperationTypeTeamRemoveUser
//
//	SystemAuditModuleOperationTypePoolCreate
//	SystemAuditModuleOperationTypePoolUpdate
//	SystemAuditModuleOperationTypePoolDelete
//	SystemAuditModuleOperationTypePoolAuthTeam
//	SystemAuditModuleOperationTypePoolRevokeTeam
//	SystemAuditModuleOperationTypePoolAuthUser
//	SystemAuditModuleOperationTypePoolRevokeUser
//
//	SystemAuditModuleOperationTypeEnvUpdateEnvValue
//	SystemAuditModuleOperationTypeEnvUpdatePoolValue
//
//	SystemAuditModuleOperationTypeApplicatonTemplateCreate
//	SystemAuditModuleOperationTypeApplicatonTemplateUpdate
//	SystemAuditModuleOperationTypeApplicatonTemplateDelete
//
//	SystemAuditModuleOperationTypeApplicatonCreate
//	SystemAuditModuleOperationTypeApplicatonUpdate
//	SystemAuditModuleOperationTypeApplicatonUpdateServiceReplicaCount
//	SystemAuditModuleOperationTypeApplicatonRestartContainer
//	SystemAuditModuleOperationTypeApplicatonUpgrade
//	SystemAuditModuleOperationTypeApplicatonRollback
//	SystemAuditModuleOperationTypeApplicatonAuthTeam
//	SystemAuditModuleOperationTypeApplicatonRevokeTeam
//	SystemAuditModuleOperationTypeApplicatonAuthUser
//	SystemAuditModuleOperationTypeApplicatonRevokeUser
//)
