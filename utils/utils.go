package utils

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"io"
	"net/http"
	"time"
)

//计算MD5
func Md5(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	str := fmt.Sprintf("%x", h.Sum(nil))
	logrus.Debugf("MD5 for string: %s, hash is %s", data, str)
	return str
}

func HttpError(w http.ResponseWriter, err string, status int) {
	logrus.WithField("status", status).Errorf("HTTP error: %v", err)
	http.Error(w, err, status)
}

//请求处理结果成功的标准操作
func HttpOK(w http.ResponseWriter, result interface{}) {
	//body := map[string]interface{}{
	//	"status" : "0",
	//	"msg"	 : "Success",
	//}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	//if result != nil {
	//	body["data"] = result
	//	json.NewEncoder(w).Encode(body)
	//}else{
	//	json.NewEncoder(w).Encode(body)
	//}
	json.NewEncoder(w).Encode(result)
}

//生成随机字符串，长度为n
func RandomStr(n int) string {
	if n > 0 {
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		s := fmt.Sprintf("%X", b)

		return s
	}
	return ""
}

//当前时间的int64格式返回值
func TimeNow64() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//将解析HTTP请求的body解析成JSON对象
//存储到对应的req模型中
//func HttpRequestBodyJsonParse(w http.ResponseWriter, r *http.Request, req interface{})  {
//	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
//		HttpError(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//}

//得到数据库表连接后的HTTP请求回调函数
type mgoCollectionsCallback func(cs map[string]*mgo.Collection)

//统一管理数据库
//批量获取表连接
//使用闭包处理API的业务逻辑
func GetMgoCollections(ctx context.Context, w http.ResponseWriter, names []string, cb mgoCollectionsCallback) {
	mgoSession, err := GetMgoSessionClone(ctx)
	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := GetAPIServerConfig(ctx).MgoDB

	var cs = make(map[string]*mgo.Collection)
	for _, name := range names {
		c := mgoSession.DB(mgoDB).C(name)
		cs[name] = c
	}

	cb(cs)
}
