package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

type TemplateListRequest struct {
	Keyword  string
	PageSize int
	Page     int
}

type TemplateListResponse struct {
	TemplateListRequest
	Total     int
	PageCount int
	Data      []types.Template
}

func getTemplateList(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &TemplateListRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Page == 0 {
		HttpError(w, "从第一页开始", http.StatusBadRequest)
		return
	}

	if req.PageSize == 0 {
		req.PageSize = 20
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("template")

	result := TemplateListResponse{
		Data: make([]types.Template, 0, 100),
	}

	pattern := fmt.Sprintf("^%s", req.Keyword)

	regex1 := bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: pattern, Options: "i"}}}

	regex2 := bson.M{"title": bson.M{"$regex": bson.RegEx{pattern, "i"}}}

	selector := bson.M{"$or": []bson.M{regex1, regex2}}

	logrus.Debugf("getTemplateList::过滤条件为%#v", regex1)

	if result.Total, err = c.Find(selector).Count(); err != nil {
		HttpError(w, fmt.Sprintf("查询记录数出错，%s", err.Error()), http.StatusInternalServerError)
		return
	}

	logrus.Debugf("getTemplateList::符合条件的template有%d个", result.Total)

	if err := c.Find(selector).Sort("title").Limit(req.PageSize).Skip(req.PageSize * (req.Page - 1)).All(&result.Data); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result.Keyword = req.Keyword
	result.Page = req.Page
	result.PageSize = req.PageSize
	result.PageCount = result.Total / result.PageSize

	httpJsonResponse(w, &result)
}

//		"/templates/{id:.*}/inspect":&MyHandler{h: getTemplate} ,
func getTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("template")

	result := types.Template{}

	if err := c.FindId(bson.ObjectIdHex(id)).One(&result); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, "模版不存在", http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpJsonResponse(w, &result)
}

type TemplateCreateRequest struct {
	types.Template
}
type TemplateCreateResponse struct {
	Id                                string
	Title, Name, Version, Description string
}

func createTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &TemplateCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//TODO 检查Name规则 ^[a-zA-Z]+\w*$
	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("template")

	n, err := c.Find(bson.M{"name": req.Name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n >= 1 {
		HttpError(w, "模版已经存在", http.StatusBadRequest)
		return
	}

	user, err := utils.GetCurrentUser(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	req.Template.Id = bson.NewObjectId()
	req.CreatorId = user.Id.Hex()
	req.UpdaterId = user.Id.Hex()
	req.CreatedTime = time.Now().Unix()
	req.UpdatedTime = req.CreatedTime

	if err := c.Insert(req.Template); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rsp := &TemplateCreateResponse{}
	rsp.Id = req.Id.Hex()
	rsp.Version = req.Version
	rsp.Name = req.Name
	rsp.Description = req.Description
	rsp.Title = req.Title

	httpJsonResponse(w, rsp)

}

type TemplateCopyRequest struct {
	Title string
}

type TemplateCopyResponse struct {
	TemplateCreateResponse
}

func copyTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &TemplateCopyRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]

	//TODO 检查Name规则 ^[a-zA-Z]+\w*$
	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("template")

	result := &types.Template{}
	err = c.FindId(bson.ObjectIdHex(id)).One(result)
	if err == mgo.ErrNotFound {
		HttpError(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := utils.GetCurrentUser(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	result.Id = bson.NewObjectId()
	result.Title = req.Title
	result.CreatorId = user.Id.Hex()
	result.UpdaterId = user.Id.Hex()
	result.CreatedTime = time.Now().Unix()
	result.UpdatedTime = result.CreatedTime

	if err := c.Insert(result); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rsp := TemplateCopyResponse{}
	rsp.Id = result.Id.Hex()
	rsp.Name = result.Name
	rsp.Title = result.Title
	rsp.Description = result.Description
	rsp.Version = result.Version

	httpJsonResponse(w, rsp)

}

type TemplateUpdateRequest struct {
	types.Template
}

type TemplateUpdateResponse struct {
	TemplateCreateResponse
}

func updateTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &TemplateUpdateRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]

	//TODO 检查Name规则 ^[a-zA-Z]+\w*$
	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("template")

	result := &types.Template{}
	err = c.FindId(bson.ObjectIdHex(id)).One(result)
	if err == mgo.ErrNotFound {
		HttpError(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := utils.GetCurrentUser(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	result.Id = bson.NewObjectId()
	result.CreatorId = user.Id.Hex()
	result.UpdaterId = user.Id.Hex()
	result.CreatedTime = time.Now().Unix()
	result.UpdatedTime = result.CreatedTime

	if err := c.UpdateId(bson.ObjectIdHex(id), result); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rsp := TemplateUpdateResponse{}
	rsp.Id = result.Id.Hex()
	rsp.Name = result.Name
	rsp.Title = result.Title
	rsp.Description = result.Description
	rsp.Version = result.Version

	httpJsonResponse(w, rsp)

}

func removeTemplate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	//TODO 检查Name规则 ^[a-zA-Z]+\w*$
	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("template")

	if err := c.RemoveId(bson.ObjectIdHex(id)); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
