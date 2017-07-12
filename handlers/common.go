package handlers

import (
	"net/http"
)

type PageRequest struct {
	Keyword  string
	PageSize int
	Page     int
}

type PageResponse struct {
	PageRequest
	Total     int
	PageCount int
}

func PageQuery(req *PageRequest, res *PageResponse, w http.ResponseWriter, r *http.Request) {
	//
	//if err := json.NewDecoder(r.Body).Decode(req); err != nil {
	//	HttpError(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	//if req.Page == 0 {
	//	HttpError(w, "从第一页开始", http.StatusBadRequest)
	//	return
	//}
	//
	//if req.PageSize == 0 {
	//	req.PageSize = 20
	//}
	//
	//mgoSession, err := utils.GetMgoSessionClone(ctx)
	//if err != nil {
	//	HttpError(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//defer mgoSession.Close()
	//
	//config := utils.GetAPIServerConfig(ctx)
	//
	//c := mgoSession.DB(config.MgoDB).C("template")
	//
	//
	//pattern := fmt.Sprintf("^%s", req.Keyword)
	//
	//regex1 := bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: pattern, Options: "i"}}}
	//
	//regex2 := bson.M{"title": bson.M{"$regex": bson.RegEx{pattern, "i"}}}
	//
	//selector := bson.M{"$or": []bson.M{regex1, regex2}}
	//
	//logrus.Debugf("getTemplateList::过滤条件为%#v", regex1)
	//
	//if res.Total, err = c.Find(selector).Count(); err != nil {
	//	HttpError(w, fmt.Sprintf("查询记录数出错，%s", err.Error()), http.StatusInternalServerError)
	//	return
	//}
	//
	//logrus.Debugf("getTemplateList::符合条件的template有%d个", res.Total)
	//
	//if err := c.Find(selector).Sort("title").Limit(req.PageSize).Skip(req.PageSize * (req.Page - 1)).All(&res.); err != nil {
	//	HttpError(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//result.Keyword = req.Keyword
	//result.Page = req.Page
	//result.PageSize = req.PageSize
	//result.PageCount = result.Total / result.PageSize
	//
	//httpJsonResponse(w, &result)
	//

}
