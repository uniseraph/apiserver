package utils

import (
	"context"
	"errors"
	"github.com/zanecloud/apiserver/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
	"time"
)

func CreateSystemAuditLogWithCtx(ctx context.Context, r *http.Request, module types.SystemAuditModuleType, operation types.SystemAuditModuleOperationType, poolId string, applicationId string, detail interface{}) (err error) {
	opUser, _ := GetCurrentUser(ctx)
	mgoSession, err := GetMgoSessionClone(ctx)

	if err != nil {
		return err
	}
	defer mgoSession.Close()

	mgoDB := GetAPIServerConfig(ctx).MgoDB

	//没有操作者的话，不允许记录
	if opUser != nil {
		return CreateSystemAuditLog(mgoSession.DB(mgoDB), r, opUser.Id.Hex(), module, operation, poolId, applicationId, detail)
	} else {
		return errors.New("系统审计记录失败，找不到当前操作者记录。")
	}

	return nil
}

//生成系统审计日志
func CreateSystemAuditLog(db *mgo.Database, r *http.Request, userId string, module types.SystemAuditModuleType, operation types.SystemAuditModuleOperationType, poolId string, applicationId string, detail interface{}) (err error) {
	c := db.C("system_audit_log")

	ip := removePostFromAddr(getIpFromReuqest(r))

	//校验参数合法性
	if ip == "" {
		return errors.New("IP could not be empty")
	}

	if userId == "" {
		return errors.New("userId could not be empty")
	}

	if module == "" {
		return errors.New("module could not be empty")
	}

	if operation == "" {
		return errors.New("operation could not be empty")
	}

	log := types.SystemAuditLog{
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

func removePostFromAddr(addr string) string {
	if addr != "" {
		ips := strings.Split(addr, ":")
		ip := ips[0]
		return ip
	}
	return ""
}
