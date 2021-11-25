package api

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/platform9/fast-path/pkg/knative"
	"github.com/platform9/fast-path/pkg/util"
)

// New returns new API router for fast-path
func New() *mux.Router {
	r := mux.NewRouter()
	/*
	//Add Authentication methods here when applicable
	if options.IsAuthEnabled() {
	}
	*/
	r.HandleFunc("/v1/apps/{space}", getApp).Methods("GET")
	return r
}


func getApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	spaceName := vars["space"]

	appList, err := knative.GetApps(util.Kubeconfig, spaceName)

	if err != nil {
		log.Error(err, "while listing app")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := []byte(appList)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(data); err != nil {
		log.Error(err, "while responding over http")
	}
}
