package handlers

import (
	"context"
	"net/http"
	"github.com/zanecloud/apiserver/store"


	"encoding/json"
	"gopkg.in/mgo.v2"
	"github.com/zanecloud/apiserver/utils"
	"github.com/Sirupsen/logrus"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"

)

func getPoolJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	mgoSession  , ok := ctx.Value(utils.KEY_MGO_SESSION).(*mgo.Session)

	if !ok {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, "cant get mgo session" , http.StatusInternalServerError)
		return
	}

	mgoDB  , ok := ctx.Value(utils.KEY_MGO_DB).(string)
	if !ok {
		HttpError(w, "cant get mgo db" , http.StatusInternalServerError)
		return
	}

	c := mgoSession.DB(mgoDB).C("pool" )


	result := store.PoolInfo{}
	if err := c.Find(bson.M{"name": name}).One(&result) ; err!=nil {


		if err==mgo.ErrNotFound {
			// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
			HttpError(w,err.Error(),http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(),http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)



}

const POOL_LABEL = "com.zanecloud.omega.pool"



var mainRoutes = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
		"/pools/{name:.*}/inspect":        mgoSessionAware(  getPoolJSON),
	},
	"POST": {

		"/pools/register":             mgoSessionAware( postPoolsRegister),

	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": OptionsHandler,
	},
}



func NewMainHandler(ctx context.Context ) http.Handler{
	return NewHandler(ctx , mainRoutes)
}

type PoolsRegisterRequest struct {
	store.PoolInfo
}



func postPoolsRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}


	var (

		name                    = r.Form.Get("name")
	)

	req := PoolsRegisterRequest{
		store.PoolInfo{
			Name:name,
			Id : bson.NewObjectId(),
		},
	}

	if err:= json.NewDecoder(r.Body).Decode(&req) ; err!=nil {
		HttpError(w, err.Error(),http.StatusBadRequest)
		return
	}

	mgoSession  , ok := ctx.Value(utils.KEY_MGO_SESSION).(*mgo.Session)

	if !ok {
		//走不到这里的
		HttpError(w, "cant get mgo session" , http.StatusInternalServerError)
		return
	}

	mgoDB  , ok := ctx.Value(utils.KEY_MGO_DB).(string)
	if !ok {
		HttpError(w, "cant get mgo db" , http.StatusInternalServerError)
		return
	}

	c := mgoSession.DB(mgoDB).C("pool" )


	n , err := c.Find(bson.M{"name": name}).Count()
	if err != nil {
		HttpError(w , err.Error() , http.StatusInternalServerError)
		return
	}

	if n>=1 {
		HttpError(w , "the pool is exist" , http.StatusConflict)
		return
	}

	if err = c.Insert(&req.PoolInfo) ; err!=nil {
		HttpError(w, err.Error(),http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Name", name)

}



func  mgoSessionAware( h Handler )  Handler {

	return  func(ctx context.Context , w http.ResponseWriter , r *http.Request){

		mgoURLS, ok := ctx.Value(utils.KEY_MGO_URLS).(string)
		if !ok {
			// context 里面没有 mongourl，这是不应该的
			logrus.Errorf("no mogodburl in ctx , ctx is #%v" , ctx)
			HttpError(w, "no mogodburl in ctx" , http.StatusInternalServerError)
			return
		}

		session, err := mgo.Dial(mgoURLS)
		if err !=nil {
			HttpError(w, err.Error(),http.StatusInternalServerError)
			return
		}

		defer  func() {
			logrus.Debug("close mgo session")
			session.Close()
		} ()

		session.SetMode(mgo.Monotonic, true)

		c := context.WithValue(ctx, utils.KEY_MGO_SESSION , session)

		logrus.Debugf("ctx is %#v" , c)

		h(c , w , r)




	}
}