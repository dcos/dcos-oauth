package common

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) *HttpError

func NewRouter(ctx context.Context, routes map[string]map[string]Handler) *mux.Router {
	r := mux.NewRouter()

	for method, mappings := range routes {
		for route, fct := range mappings {
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				log.WithFields(log.Fields{"method": r.Method, "uri": r.RequestURI}).Info("HTTP request received")

				err := localFct(ctx, w, r)
				if err != nil {
					log.WithFields(log.Fields{"method": r.Method, "uri": r.RequestURI}).Info(err.Description)
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.Header().Set("X-Content-Type-Options", "nosniff")
					w.WriteHeader(err.Status)
					enc := json.NewEncoder(w)
					enc.Encode(err)
					return
				}
			}

			r.Path(route).Methods(method).HandlerFunc(wrap)
		}
	}
	return r
}
