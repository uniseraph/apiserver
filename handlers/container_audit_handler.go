package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/proxy/swarm"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
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
	Command string
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

	utils.GetMgoCollections(ctx, w, []string{"pool", "container"}, func(cs map[string]*mgo.Collection) {
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

		/*
			验证用户是否有权访问该容器
		*/

		//检查当前用户是否有权限操作该容器
		if user.RoleSet&types.ROLESET_SYSADMIN == types.ROLESET_SYSADMIN {
			//如果用户是系统管理员
			//则不需要校验用户对该机器的权限
			goto AUTHORIZED
		}

		//如果当前容器所在集群
		//已经给当前用户授权过
		//则验证通过
		for _, id := range user.PoolIds {
			if id.Hex() == container.PoolId {
				goto AUTHORIZED
			}
		}

		//如果该用户加入过某些团队
		//则该团队能查看的pool
		//该用户也可以查看
		//则验证通过
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

			//如果用户所在的某个TEAM
			//拥有对该集群的授权
			//则验证通过
			for _, team := range teams {
				for _, id := range team.PoolIds {
					if id.Hex() == container.PoolId {
						goto AUTHORIZED
					}
				}
			}
		}

	AUTHORIZED:

		pool := &types.PoolInfo{}
		if err := cs["pool"].FindId(bson.ObjectIdHex(container.PoolId)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, fmt.Sprintf("no such id for pool: %s", container.PoolId), http.StatusNotFound)
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//如果用户对该Container有权操作
		//则生成临时的token给用户
		var token string
		if token, err = utils.CreateSSHSession(ctx, container.Name, container.Id.Hex(), container.ContainerId, container.ApplicationId, container.Service, user, pool); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ssh := utils.GenerateSSHToken(token, pool)
		rsp := CreateSSHSessionResponse{
			Command: ssh,
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

	token := req.Token

	//创建会话需要一条操作记录
	//记录这个会话创建动作的结果
	log := types.ContainerAuditLog{
		Id:     bson.NewObjectId(),
		Token:  req.Token,
		Ip:     req.User,
		Detail: types.ContainerAuditLogOperationDetail{},
	}

	//需要检查当Redis的KEY不存在时
	//info是nil还是空的map
	info, err := utils.FetchContainerFromSSHCache(ctx, token)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//如果验证登陆不成功
	//要记录OPERATION是不成功的log实例
	if len(info) <= 0 {
		utils.GetMgoCollections(ctx, w, []string{"container_audit_log"}, func(cs map[string]*mgo.Collection) {
			if err := validateSSHSessionFailedLog(cs, fmt.Sprintf("Token is invalid: %s", token), log); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
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
	scid := info["scid"]   //Swarm Container Id
	pname := info["pname"] //集群名称
	pid := info["pid"]     //集群ID
	aid := info["aid"]     //应用ID
	sname := info["sname"] //Service名称

	//当前操作用户
	u := types.ContainerAuditUser{
		Id:   bson.ObjectIdHex(uid),
		Name: uname,
	}

	//容器
	c := types.ContainerAuditContainer{
		Id:               bson.ObjectIdHex(cid),
		Name:             cname,
		SwarmContainerId: scid,
	}

	utils.GetMgoCollections(ctx, w, []string{"application", "container_audit_trace", "container_audit_log", "container"}, func(cs map[string]*mgo.Collection) {
		app := types.Application{}

		if err := cs["application"].FindId(bson.ObjectIdHex(aid)).One(&app); err != nil {
			if e := validateSSHSessionFailedLog(cs, err.Error(), log); err != nil {
				HttpError(w, e.Error(), http.StatusInternalServerError)
				return
			}
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a application", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
		}

		//Service
		var service *types.Service
		for _, s := range app.Services {
			if s.Name == sname {
				service = &s
				break
			}
		}
		if service == nil {
			if err := validateSSHSessionFailedLog(cs, fmt.Sprintf("could not found service with name: %s", sname), log); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			HttpError(w, fmt.Sprintf("could not found service with name: %s", sname), http.StatusNotFound)
			return
		}

		//应用
		a := types.ContainerAuditApplication{
			Id:      bson.ObjectIdHex(aid),
			Name:    app.Name,
			Title:   app.Title,
			Version: app.Version,
		}

		//集群
		p := types.ContainerAuditPool{
			Id:   bson.ObjectIdHex(pid),
			Name: pname,
		}

		//服务
		s := types.ContainerAuditService{
			Name:  service.Name,
			Title: service.Title,
		}

		//创建一次Tunnel的会话过程记录
		//记住环境信息，但不包含会话过程中的交互信息
		//会话过程中的交互信息存在Log表中
		//Trace has many Logs
		trace := types.ContainerAuditTrace{
			Id:            bson.NewObjectId(),
			Token:         token,
			UserId:        bson.ObjectIdHex(uid),
			User:          u,
			ContainerId:   bson.ObjectIdHex(cid),
			Container:     c,
			PoolId:        bson.ObjectIdHex(pid),
			Pool:          p,
			ApplicationId: a.Id,
			Application:   a,
			Service:       s,

			CreatedTime: time.Now().Unix(),
		}

		if err := cs["container_audit_trace"].Insert(trace); err != nil {
			if e := validateSSHSessionFailedLog(cs, err.Error(), log); err != nil {
				HttpError(w, e.Error(), http.StatusInternalServerError)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//认证成功后
		//删除该Token
		//避免被重复使用
		if err := utils.RemoveSSHSession(ctx, token); err != nil {
			if e := validateSSHSessionFailedLog(cs, err.Error(), log); err != nil {
				HttpError(w, e.Error(), http.StatusInternalServerError)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//到此为止登录成功
		//记录登录成功的操作
		if e := validateSSHSessionSuccessLog(cs, log); err != nil {
			HttpError(w, e.Error(), http.StatusInternalServerError)
			return
		}

		rlt := ValidateSSHSessionResponse{
			Result:    "OK",
			Status:    1,
			Container: scid,
		}

		HttpOK(w, rlt)
	})
}

//写入一条登录记录
//将登录的输入和失败结果入库
func validateSSHSessionFailedLog(cs map[string]*mgo.Collection, msg string, log types.ContainerAuditLog) (err error) {
	log.Detail.Reason = msg
	log.Operation = "LoginFailed"
	if err := cs["container_audit_log"].Insert(log); err != nil {
		return err
	}
	return nil
}

//写入一条登录记录
//将登录的输入和成功结果入库
func validateSSHSessionSuccessLog(cs map[string]*mgo.Collection, log types.ContainerAuditLog) (err error) {
	log.Operation = "Logined"
	if err := cs["container_audit_log"].Insert(log); err != nil {
		return err
	}
	return nil
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

		var output string
		//处理输入字符串特别长的情况
		if len(req.Output) > 8000 {
			//不考虑中文字符被截断的情况
			//顶多被前后各截断1个汉字
			//没必要转换成rune的slice然后再处理，性能开销的性价比不高

			//取前4k字符
			prefixStr := req.Output[:4000]
			//取后4k字符
			suffixStr := req.Output[(len(req.Output) - 4000):]

			output = prefixStr + "\n......\n" + suffixStr
		}

		audit := types.ContainerAuditLog{
			Id:        bson.NewObjectId(),
			Ip:        req.Ip,
			Token:     req.Token,
			Operation: "ExecCmd",
			Detail: types.ContainerAuditLogOperationDetail{
				Command:   cmd,
				Arguments: args,
				Stdout:    output,
			},

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
	StartTime     string `json:",int"`
	EndTime       string `json:",int"`
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
	UserId        string
	User          types.ContainerAuditUser
	PoolId        string
	Pool          types.ContainerAuditPool
	ApplicationId string
	Application   types.ContainerAuditApplication
	Service       types.ContainerAuditService
	ContainerId   string
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
	if len(req.StartTime) > 0 || len(req.EndTime) > 0 {
		createdtime := bson.M{}

		if len(req.StartTime) > 0 {
			i, err := strconv.ParseInt(req.StartTime, 10, 64)
			if err != nil {
				panic(err)
			}
			tm := time.Unix(i, 0)
			createdtime["$gte"] = tm
		}
		if len(req.EndTime) > 0 {
			i, err := strconv.ParseInt(req.EndTime, 10, 64)
			if err != nil {
				panic(err)
			}
			tm := time.Unix(i, 0)
			createdtime["$lt"] = tm
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

	page = req.Page - 1

	if page <= 0 {
		page = 0
	}

	pageSize = req.PageSize

	if pageSize == 0 {
		//默认每页20条
		pageSize = 20
	}

	utils.GetMgoCollections(ctx, w, []string{"container_audit_log", "container_audit_trace"}, func(cs map[string]*mgo.Collection) {
		logs := make([]types.ContainerAuditLog, 0, pageSize)

		if err := cs["container_audit_log"].Find(selector).Sort("-createdtime").Skip(page * pageSize).Limit(pageSize).All(&logs); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var pageCount, total int
		if t, err := cs["container_audit_log"].Find(selector).Count(); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			total = t
		}

		if total%pageSize == 0 {
			pageCount = total / pageSize
		} else {
			pageCount = total/pageSize + 1
		}

		//根据Log集合的tokens
		//找到每个log对应的trace记录
		tokens := make(map[string]string)
		for _, log := range logs {
			tokens[log.Token] = "ok"
		}
		tokenKeys := make([]string, 0, 20)
		for k := range tokens {
			tokenKeys = append(tokenKeys, k)
		}

		selector = bson.M{
			"token": bson.M{
				"$in": tokenKeys,
			},
		}

		//查询traces记录
		traces := make([]types.ContainerAuditTrace, 0, 20)
		if err := cs["container_audit_trace"].Find(selector).All(&traces); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//将trace记录做成map
		//以便整理数据的时候
		//可以根据TOKEN作为KEY，来查找log对应的trace
		tracesMap := make(map[string]types.ContainerAuditTrace)
		for _, t := range traces {
			tracesMap[t.Token] = t
		}

		data := make([]GetAuditListResponseData, 0, len(logs))

		//整理数据
		//每条数据由如下组成：一条log信息，及log对应的trace信息
		for _, log := range logs {
			t, ok := tracesMap[log.Token]
			if !ok {
				logrus.Errorf("no trace found for token: %s", log.Token)
				continue
			}
			d := GetAuditListResponseData{
				Id:            log.Id.Hex(),
				UserId:        t.UserId.Hex(),
				User:          t.User,
				PoolId:        t.PoolId.Hex(),
				Pool:          t.Pool,
				ApplicationId: t.ApplicationId.Hex(),
				Application:   t.Application,
				Service:       t.Service,
				ContainerId:   t.ContainerId.Hex(),
				Container:     t.Container,

				Ip:        log.Ip,
				Cmd:       log.Detail.Command,
				Arguments: log.Detail.Arguments,
				Stdout:    log.Detail.Stdout,
				Stderr:    log.Detail.Stderr,
				Stdin:     log.Detail.Stdin,
				ExitCode:  log.Detail.ExitCode,

				CreatedTime: log.CreatedTime,
			}
			data = append(data, d)
		}

		//整理返回值
		rlt := GetAuditListResponse{
			Total:     total,
			PageCount: pageCount,
			PageSize:  pageSize,
			Page:      page,
			Data:      data,
		}

		HttpOK(w, rlt)

	})
}
