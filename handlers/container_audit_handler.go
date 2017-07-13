package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zanecloud/apiserver/proxy/swarm"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
	"time"
)

/*
0、/audit/ssh 拿到临时的SSH登录字符串
1、/audit/login Turnnel验证登录身份
2、/audit/log Turnnel记录操作行为
3、/audit/list 审计数据历史记录
*/

/*
/audit/ssh
*/

type CreateSSHSessionResponse struct {
	Token string
}

func createSSHSession(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := r.Form.Get("ContainerId")

	if len(id) <= 0 {
		HttpError(w, "need ContainerId in url params", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"container_audit", "container", "containeraudittrace"}, func(cs map[string]*mgo.Collection) {
		container := swarm.Container{}
		selector := bson.M{
			"_id": bson.ObjectIdHex(id),
		}

		if err := cs["container"].Find(selector).One(&container); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, fmt.Sprintf("no such id for container: %s", id), http.StatusNotFound)
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := utils.GetCurrentUser(ctx)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//检查当前用户是否有权限操作该容器
		if user.RoleSet&types.ROLESET_SYSADMIN == types.ROLESET_SYSADMIN {
			//如果用户是系统管理员
			//则不需要校验用户对该机器的权限
		} else {
			//如果不是系统管理员
			//则找到该用户能查看的所有pool id准备查询
			poolIds := make([]bson.ObjectId, 0, 10)
			//如果该用户加入过某些团队
			//则该团队能查看的pool
			//该用户也可以查看
			if len(user.TeamIds) > 0 {
				teams := make([]types.Team, 0, 10)
				selector = bson.M{
					"_id": bson.M{
						"$in": user.TeamIds,
					},
				}
				//查找该用户所在Team
				if err := cs["team"].Find(selector).All(&teams); err != nil {
					if err == mgo.ErrNotFound {
						HttpError(w, "not found params", http.StatusNotFound)
						return
					}
					HttpError(w, err.Error(), http.StatusNotFound)
					return
				}

				for _, team := range teams {
					poolIds = append(poolIds, team.PoolIds...)
				}
			}
			//将授权给用户的pool id也加入查询条件
			poolIds = append(poolIds, user.PoolIds...)

			selector = bson.M{
				"_id": bson.M{
					"$in": poolIds,
				},
				"name": container.PoolName,
			}

			//批量查找出Pool数据
			if c, err := cs["pool"].Find(selector).Count(); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, "the container is not permit access.", http.StatusUnauthorized)
					return
				}
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			} else if c <= 0 {
				HttpError(w, "the container is not permit access.", http.StatusUnauthorized)
				return
			} else {
				//允许访问
			}
		}

		//如果用户对该Container有权操作
		//则生成临时的token给用户
		var token string
		if token, err = utils.CreateSSHSession(ctx, container, user); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ssh := utils.GenerateSSHToken(user.Name, token)
		rsp := CreateSSHSessionResponse{
			Token: ssh,
		}
		HttpOK(w, rsp)

	})
}

/*
/audit/login
*/

type ValidateSSHSessionRequest struct {
	Token     string
	User      string
	Timestamp string
}

type ValidateSSHSessionResponse struct {
	Result    string
	Status    int
	Container string
}

func validateSSHSession(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := ValidateSSHSessionRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Token) <= 0 || len(req.User) <= 0 || len(req.Timestamp) <= 0 {
		HttpError(w, "request with invalidate params.", http.StatusBadRequest)
		return
	}

	token, err := utils.ParseSSHToken(req.Token)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info, err := utils.FetchContainerFromSSHCache(ctx, token)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*
		如果鉴权成功
		则生成用于本次会话的用户跟踪模型
	*/

	uid := info["uid"]     //用户ID
	uname := info["uname"] //用户名称
	cid := info["cid"]     //容器ID
	cname := info["cname"] //容器名称
	pname := info["pname"] //集群名称
	url := info["url"]     //tunnel的URL

	//当前操作用户
	u := types.ContainerAuditUser{
		Id:   bson.ObjectIdHex(uid),
		Name: uname,
	}

	//容器
	c := types.ContainerAuditContainer{
		Id:   bson.ObjectIdHex(cid),
		Name: cname,
	}

	utils.GetMgoCollections(ctx, w, []string{"pool", "container_audit_trace", "container"}, func(cs map[string]*mgo.Collection) {
		pool := &types.PoolInfo{}
		//找到容器所属的集群
		if err := cs["pool"].Find(bson.M{"name": pname}).One(&pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a pool", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		//集群
		p := types.ContainerAuditPool{
			Id:   pool.Id,
			Name: pool.Name,
		}

		trace := types.ContainerAuditTrace{
			Id:          bson.NewObjectId(),
			Token:       token,
			UserId:      bson.ObjectIdHex(uid),
			User:        u,
			ContainerId: bson.ObjectIdHex(cid),
			Container:   c,
			PoolId:      pool.Id,
			Pool:        p,

			CreatedTime: time.Now().Unix(),
		}

		if err := cs["container_audit_trace"].Insert(trace); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rlt := ValidateSSHSessionResponse{
			Result:    "OK",
			Status:    1,
			Container: url,
		}

		HttpOK(w, rlt)
	})
}

/*
/audit/log
*/

//{
//"Token": "ac94970cd14940d59b303ff2c2a68bff",	// 复用Token用作trace id
//"User": "1.1.1.1:11011",	// 用户访问客户端IP
//"Command": "ls /root",	// 用户本次执行的命令（长度最长为1024 * 1024个字符）
//"Output": "",	// 预留字段。用户本次执行的命令输出
//"Timestamp": "2017-07-11 18:11:11"	// 执行命令的时间
//}

type CreateAuditLogRequest struct {
	Token     string
	Ip        string
	Command   string
	Output    string
	Timestamp string
}

type CreateAuditLogResponse struct {
	Result string
}

func createAuditLog(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := CreateAuditLogRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Token) <= 0 || len(req.Ip) <= 0 || len(req.Command) <= 0 {
		HttpError(w, "request with invalidate params.", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"container_audit", "container_audit_trace"}, func(cs map[string]*mgo.Collection) {

		//验证是否Token的合法性
		//考虑性能的话，可以不做校验，但会增加垃圾数据
		if c, err := cs["container_audit_trace"].Find(bson.M{"token": req.Token}).Count(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a trace", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		} else if c <= 0 {
			HttpError(w, "no such a trace", http.StatusNotFound)
			return
		}

		//将调用方上传过来的命令行
		//解析为命令文件和参数数组
		cmds := strings.Split(req.Command, " ")
		var cmd string
		args := make([]string, 0, 10)
		for _, c := range cmds {
			if len(c) > 0 {
				//如果CMD没有被初始化过
				if len(cmd) <= 0 {
					cmd = c
				} else {
					//保存到参数
					args = append(args, c)
				}
			}
		}

		audit := types.ContainerAuditLog{
			Id:        bson.NewObjectId(),
			Ip:        req.Ip,
			TraceId:   req.Token,
			Cmd:       cmd,
			Arguments: args,
			Stdout:    req.Output,

			CreatedTime: time.Now().Unix(),
		}

		if err := cs["container_audit"].Insert(audit); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp := CreateAuditLogResponse{
			Result: "OK",
		}

		HttpOK(w, rsp)
	})
}

/*
/audit/log/update
*/
func updateAuditLog(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//TODO
}

/*
/audit/list
*/

type GetAuditListRequest struct {
	StartTime     time.Time
	EndTime       time.Time
	UserId        string
	IP            string
	Operation     string
	ApplicationId string
	ServiceName   string
	ContainerId   string
	PageSize      int
	Page          int
}

type GetAuditListResponseData struct {
	/*
		从ContainerAuditTrace模型中获取
	*/
	Id            string
	UserId        bson.ObjectId
	User          types.ContainerAuditUser
	PoolId        bson.ObjectId
	Pool          types.ContainerAuditPool
	ApplicationId bson.ObjectId
	Application   types.ContainerAuditApplication
	ServiceId     bson.ObjectId
	Service       types.ContainerAuditService
	ContainerId   bson.ObjectId
	Container     types.ContainerAuditContainer

	/*
		从ContainerAuditLog模型中获取
	*/
	Ip        string
	Cmd       string
	Arguments []string
	Stderr    string
	Stdout    string
	Stdin     string
	ExitCode  int8

	CreatedTime int64 `json:",omitempty"`
}
type GetAuditListResponse struct {
	Total     int
	PageCount int
	PageSize  int
	Page      int
	Data      []GetAuditListResponseData
}

func getAuditList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := GetAuditListRequest{}
	var page, pageSize int

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	selector := bson.M{}

	//设定时间查询条件
	if !req.StartTime.IsZero() || !req.EndTime.IsZero() {
		createdtime := bson.M{}

		if !req.StartTime.IsZero() {
			createdtime["$gte"] = req.StartTime
		}
		if !req.EndTime.IsZero() {
			createdtime["$lt"] = req.EndTime
		}

		selector["createdtime"] = createdtime
	}

	if len(req.UserId) > 0 {
		selector["userid"] = bson.ObjectIdHex(req.UserId)
	}

	if len(req.IP) > 0 {
		selector["ip"] = req.IP
	}

	if len(req.ApplicationId) > 0 {
		selector["applicationid"] = bson.ObjectIdHex(req.ApplicationId)
	}

	if len(req.ServiceName) > 0 {
		//查询HashMap中的某个字段值
		selector["service.name"] = req.ServiceName
	}

	if len(req.ContainerId) > 0 {
		selector["containerid"] = bson.ObjectIdHex(req.ContainerId)
	}

	if req.Page != 0 {
		//前端page第一页从1开始计数
		page = req.Page - 1
	}

	if req.PageSize == 0 {
		//默认每页20条
		pageSize = 20
	}

	utils.GetMgoCollections(ctx, w, []string{"container_audit_log"}, func(cs map[string]*mgo.Collection) {
		data := make([]types.ContainerAuditLog, 0, pageSize)

		if err := cs["container_audit_log"].Find(selector).Sort("-createdtime").Skip(page * pageSize).Limit(pageSize).All(&data); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})
}
