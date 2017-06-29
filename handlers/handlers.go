package handlers

import (
	"net/http"
	"strings"

	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/utils"
)

type Handler func(c context.Context, w http.ResponseWriter, r *http.Request)

func NewHandler(ctx context.Context, rs map[string]map[string]Handler) http.Handler {

	r := mux.NewRouter()
	SetupPrimaryRouter(r, ctx, rs)
	return r

}

func SetupPrimaryRouter(r *mux.Router, ctx context.Context, rs map[string]map[string]Handler) {
	for method, mappings := range rs {
		for route, fct := range mappings {
			log.WithFields(log.Fields{"method": method, "route": route}).Debug("Registering HTTP route")


			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				log.WithFields(log.Fields{"method": r.Method, "uri": r.RequestURI}).Debug("HTTP request received")

				localFct(ctx, w, r)
			}
			localMethod := method

			r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			r.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}
}


func BoolValue(r *http.Request, k string) bool {
	s := strings.ToLower(strings.TrimSpace(r.FormValue(k)))
	return !(s == "" || s == "0" || s == "no" || s == "false" || s == "none")
}

// Default handler for methods not supported by clustering.
func notImplementedHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.HttpError(w, "Not supported in clustering mode.", http.StatusNotImplemented)
}

func OptionsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HttpError(w http.ResponseWriter, err string, status int) {
	utils.HttpError(w,err,status)
}
