package api

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	//"knative.dev/client/pkg/kn/commands/service"
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
	r.HandleFunc("/v1/apps", createApp).Methods("POST")
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

type App struct {
	Name  string `json:"name"`
	Space string `json:"space"`
	Image string `json:"image"`
}

func createApp(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
	fmt.Printf("vars : %v\n", vars)

	app := App{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "while reading data in request body")
		return
	}

	err = json.Unmarshal(body, &app)
	if err != nil {
		log.Error(err, "while unmarhsalling request body data")
		return
	}

	fmt.Printf("Name: %s, space: %s, image: %s\n", app.Name, app.Space, app.Image)

	err = knative.CreateApp(util.Kubeconfig, app.Name, app.Space, app.Image)
	if err != nil {
		log.Error(err, "while creating app")
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusOK)
}
