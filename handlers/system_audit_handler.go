package handlers

import (
	"context"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

type GetSystemAuditListRequest struct {
	StartTime int64 `json:",int"`
	EndTime   int64 `json:",int"`
	UserId    string
	PoolId    string
	IP        string
	Module    types.SystemAuditModuleType
	Operation types.SystemAuditModuleOperationType
	PageSize  int
	Page      int
}

/*
Id,
CreatedTime,
UserId,
User: { Id, Name },
PoolId,
Pool: { Id, Name },
ApplicationId,
Application: { Id, Title, Name, Version },
IP,
Module:
Operation:
Detail: 上述JSON格式
*/

type GetSystemAuditListResponseUserData struct {
	Id   string
	Name string
}

type GetSystemAuditListResponsePoolData struct {
	Id   string
	Name string
}

type GetSystemAuditListResponseApplicationData struct {
	Id      string
	Title   string
	Name    string
	Version string
}

type GetSystemAuditListResponseData struct {
	/*
		从SystemAuditLog模型中获取
	*/
	Id            string
	RequestURI    string
	CreatedTime   int64 `json:",omitempty"`
	UserId        string
	User          GetSystemAuditListResponseUserData
	PoolId        string
	Pool          GetSystemAuditListResponsePoolData
	ApplicationId string
	Application   GetSystemAuditListResponseApplicationData
	IP            string
	Module        types.SystemAuditModuleType
	Operation     types.SystemAuditModuleOperationType
	Detail        interface{}
}
type GetSystemAuditListResponse struct {
	Total     int
	PageCount int
	PageSize  int
	Page      int
	Data      []GetSystemAuditListResponseData
}

func getSystemAuditList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := GetSystemAuditListRequest{}
	var page, pageSize int

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	selector := bson.M{}

	//设定时间查询条件
	//if len(req.StartTime) > 0 || len(req.EndTime) > 0 {
	if req.StartTime > 0 || req.EndTime > 0 {
		createdtime := bson.M{}

		if req.StartTime > 0 {
			tm := time.Unix(req.StartTime, 0)
			createdtime["$gte"] = tm
		}

		if req.EndTime > 0 {
			tm := time.Unix(req.EndTime, 0)
			createdtime["$lt"] = tm
		}

		selector["createdtime"] = createdtime
	}

	if len(req.UserId) > 0 {
		selector["userid"] = bson.ObjectIdHex(req.UserId)
	}

	if len(req.PoolId) > 0 {
		selector["poolid"] = bson.ObjectIdHex(req.PoolId)
	}

	if len(req.IP) > 0 {
		selector["ip"] = req.IP
	}

	if req.Module > 0 {
		selector["module"] = req.Module
	}

	if req.Operation > 0 {
		selector["operation"] = req.Operation
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

	utils.GetMgoCollections(ctx, w, []string{"system_audit_log", "user", "pool", "application"}, func(cs map[string]*mgo.Collection) {
		logs := make([]types.SystemAuditLog, 0, pageSize)

		if err := cs["system_audit_log"].Find(selector).Sort("-createdtime").Skip(page * pageSize).Limit(pageSize).All(&logs); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var pageCount, total int
		if t, err := cs["system_audit_log"].Find(selector).Count(); err != nil {
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

		/*
			组装数据
		*/
		data := make([]GetSystemAuditListResponseData, 0, len(logs))

		/*
			做一个Cache，避免每次查询都走数据库
		*/
		userCache := make(map[bson.ObjectId]types.User)
		poolCache := make(map[bson.ObjectId]types.PoolInfo)
		appCache := make(map[bson.ObjectId]types.Application)

		//整理数据
		//每条数据由如下组成：一条log信息，及log对应的trace信息
		for _, log := range logs {
			d := GetSystemAuditListResponseData{
				Id:            log.Id.Hex(),
				RequestURI:    log.RequestURI,
				UserId:        log.UserId.Hex(),
				PoolId:        log.PoolId.Hex(),
				ApplicationId: log.ApplicationId.Hex(),
				IP:            log.IP,
				Module:        log.Module,
				Operation:     log.Operation,
				Detail:        log.Detail,
				CreatedTime:   log.CreatedTime,
			}
			if log.UserId != "" {
				user := types.User{}

				user, ok := userCache[log.UserId]
				if ok {
					d.User = GetSystemAuditListResponseUserData{
						Id:   user.Id.Hex(),
						Name: user.Name,
					}
				} else {
					if err := cs["user"].FindId(log.UserId).One(&user); err != nil {
						logrus.Errorf("Sysstem autid list: Could not fount user with id: %s", log.UserId.Hex())
					} else {
						userCache[log.UserId] = user

						d.User = GetSystemAuditListResponseUserData{
							Id:   user.Id.Hex(),
							Name: user.Name,
						}
					}
				}

			}

			if log.PoolId != "" {
				pool := types.PoolInfo{}

				pool, ok := poolCache[log.PoolId]
				if ok {
					d.Pool = GetSystemAuditListResponsePoolData{
						Id:   pool.Id.Hex(),
						Name: pool.Name,
					}
				} else {
					if err := cs["pool"].FindId(log.PoolId).One(&pool); err != nil {
						logrus.Errorf("Sysstem autid list: Could not fount pool with id: %s", log.PoolId.Hex())
					} else {
						poolCache[log.PoolId] = pool

						d.Pool = GetSystemAuditListResponsePoolData{
							Id:   pool.Id.Hex(),
							Name: pool.Name,
						}
					}
				}

			}

			if log.ApplicationId != "" {
				app := types.Application{}

				app, ok := appCache[log.ApplicationId]
				if ok {
					d.Application = GetSystemAuditListResponseApplicationData{
						Id:      app.Id.Hex(),
						Title:   app.Title,
						Name:    app.Name,
						Version: app.Version,
					}
				} else {
					if err := cs["application"].FindId(log.ApplicationId).One(&app); err != nil {
						logrus.Errorf("Sysstem autid list: Could not fount application with id: %s", log.ApplicationId.Hex())
					} else {
						appCache[log.ApplicationId] = app

						d.Application = GetSystemAuditListResponseApplicationData{
							Id:      app.Id.Hex(),
							Title:   app.Title,
							Name:    app.Name,
							Version: app.Version,
						}
					}
				}

			}
			data = append(data, d)
		}

		//整理返回值
		rlt := GetSystemAuditListResponse{
			Total:     total,
			PageCount: pageCount,
			PageSize:  pageSize,
			Page:      page,
			Data:      data,
		}

		HttpOK(w, rlt)

	})
}
