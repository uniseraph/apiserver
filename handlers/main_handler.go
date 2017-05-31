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
)

var mainRoutes = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
	},
	"POST": {

		"/pools/register":             mgoSessionAware( postPoolsRegister),

	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": optionsHandler,
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
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}


	var (

		name                    = r.Form.Get("name")
	)

	req := PoolsRegisterRequest{
		store.PoolInfo{
			Name:name,
		},
	}

	if err:= json.NewDecoder(r.Body).Decode(&req) ; err!=nil {
		httpError(w, err.Error(),http.StatusBadRequest)
		return
	}

	mgoSession  , ok := ctx.Value(utils.KEY_MGO_SESSION).(*mgo.Session)

	if !ok {
		//走不到这里的
		httpError(w, "cant get mgo session" , http.StatusInternalServerError)
		return
	}

	mgoDB  , ok := ctx.Value(utils.KEY_MGO_DB).(string)
	if !ok {
		httpError(w, "cant get mgo db" , http.StatusInternalServerError)
		return
	}
	c := mgoSession.DB(mgoDB).C("pool" )

	err := c.Insert(&req.PoolInfo)
	if err!=nil {
		httpError(w, err.Error(),http.StatusInternalServerError)
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
			httpError(w, "no mogodburl in ctx" , http.StatusInternalServerError)
			return
		}

		session, err := mgo.Dial(mgoURLS)
		if err !=nil {
			httpError(w, err.Error(),http.StatusInternalServerError)
			return
		}

		defer  func() {
			logrus.Debug("close mgo session")
			session.Close()
		} ()

		session.SetMode(mgo.Monotonic, true)

		c := context.WithValue(ctx, "mgoSession" , ctx)

		h(c , w , r)




	}
}