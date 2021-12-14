package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	//"knative.dev/client/pkg/kn/commands/service"
	"github.com/platform9/fast-path/pkg/knative"
	"github.com/platform9/fast-path/pkg/util"

	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	b64 "encoding/base64"
)

type DB struct {
	Name      string
	Email     string
	NameSpace string
	NickName  string
	//TokenExp  float64
}

type Header struct {
	Algorithm     string `json:"alg"`
	TokenType     string `json:"typ"`
	KeyIdentifier string `json:"kid"`
}
type Payload struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Issuer    string    `json:"iss"`
	SubIssuer string    `json:"sub"`
	Audience  string    `json:"aud"`
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
}

var LocalDB []DB

// New returns new API router for fast-path
func New() *mux.Router {
	r := mux.NewRouter()
	/*
		//Add Authentication methods here when applicable
		if options.IsAuthEnabled() {
		}
	*/

	r.HandleFunc("/v1/apps/{space}", getApp).Methods("GET")
	r.HandleFunc("/v1/apps/{space}/{name}", getAppByName).Methods("GET")
	r.HandleFunc("/v1/apps", createApp).Methods("POST")
	r.HandleFunc("/v1/apps/login", loginApp).Methods("POST")
	r.HandleFunc("/v1/apps/{space}/{name}", deleteApp).Methods("DELETE")

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

/*
1. User fires appctl deploy command with name, image, token as bearer.
2. Validate the token, get user name, token expiry.
3. If token is expired return message token expired. Else continue..
	4. Fetch Username - Namespace map in DB. Use this as nameSpace.
	5. CreateApp.
*/

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

func getAppByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	space := vars["space"]
	appName := vars["name"]

	appList, err := knative.GetAppByName(util.Kubeconfig, space, appName)

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

func deleteApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deleteAppSpace := vars["space"]
	deleteAppName := vars["name"]

	fmt.Printf("vars : %v\n", vars)

	fmt.Printf("Name: %s, space: %s", deleteAppName, deleteAppSpace)

	errdel := knative.DeleteApp(util.Kubeconfig, deleteAppSpace, deleteAppName)
	if errdel != nil {
		log.Error(errdel, "while deleting app")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func loginApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("vars : %v\n", vars)

	// Fetch the token.
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	fmt.Printf("The token is %v\n\n", reqToken)

	//Validate the token and decode it and get userinfo.
	//ValidateToken(reqToken)
	header, payload, err := DecodeToken(reqToken)
	if err != nil {
		fmt.Printf("Faield to decode the token\n")
		return
	}
	name := fmt.Sprintf("%v", payload["name"])
	email := fmt.Sprintf("%v", payload["email"])
	aud := fmt.Sprintf("%v", payload["aud"])
	nickname := fmt.Sprintf("%v", payload["nickname"])
	sub := fmt.Sprintf("%v", payload["sub"])

	// Doing simple validation i.e if audiance == auth0 clientID
	if aud == util.ClientId {
		fmt.Printf("Its valid token\n")
	}

	fmt.Printf("The token decoded is header: %v, payload %v\n\n", header, payload)

	// If user exists in DB, then check if its new expiry.
	var UserExists bool
	for _, val := range LocalDB {

		if val.Email == email {
			UserExists = true
			/*if val.TokenExp != payload["exp"] {
				//LocalDB[key].TokenExp =
			}*/
		}
		// For Github case
		if val.NickName == nickname {
			UserExists = true
		}
	}
	var nameSpace string
	if !UserExists {

		if strings.Contains(sub, "github") {
			nameSpace = nickname + "Random 6 digit code"
		} else {
			nameSpace = strings.Split(email, "@")[0] + "Random 6 digit code."
		}

		// Check if namespace is valid. bcz only regex valid "^*[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
		if util.RegexValidate(nameSpace) {
			fmt.Printf("Valid Namespace")
		} else {
			// Have to create a unique namespace name.
		}

		// Fetch the details and check for namespace.
		// Create new namespace.
		errNamespace := CreateNamespace(nameSpace)
		if errNamespace != nil {
			fmt.Printf("Failed to create a namespace with name %v\n", nameSpace)
			return
		}

		fmt.Printf("Successfully created namespace %v\n", nameSpace)

		// Add Userinfo to DB.
		LocalDB = append(LocalDB, DB{Name: name, Email: email, NameSpace: nameSpace, NickName: nickname})
		fmt.Printf("The local database is %v\n", LocalDB)
	}

	w.WriteHeader(http.StatusOK)
}

// To validate given token.
func ValidateToken(token string) {
	// Logic to be implemented.
}

func CreateNamespace(nameSpace string) error {
	config, err := clientcmd.BuildConfigFromFlags("", util.Kubeconfig)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}
	// Check if the creating namespace already exist.
	ns, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Cant list namespaces ")
		return err
	}

	for _, val := range (*ns).Items {
		if val.Name == nameSpace {
			fmt.Printf("Namespace found")
			nameSpace = nameSpace + string(rand.Intn(10))
		}
	}
	// Create a new Namespace for first time login user.
	nsName := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nameSpace,
		},
	}

	_, errCreate := clientset.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
	if errCreate != nil {
		fmt.Printf("Failed to create a new namespace %v", nameSpace)
		return errCreate
	}
	return nil
}

// Simple Decode Token.
func DecodeToken(token string) (map[string]interface{}, map[string]interface{}, error) {
	rawHeader := strings.Split(token, ".")[0]
	rawPayload := strings.Split(token, ".")[1]

	decodedHeader, _ := b64.StdEncoding.DecodeString(rawHeader)
	decodedPayload, _ := b64.StdEncoding.DecodeString(rawPayload)

	fmt.Printf("Decoded Header: %s\nDecoded Payload:%s", decodedHeader, decodedPayload)

	var headerStruct map[string]interface{}
	var payloadStruct map[string]interface{}

	errHead := json.Unmarshal(decodedHeader, &headerStruct)
	if errHead != nil {
		return nil, nil, errHead
	}

	errPay := json.Unmarshal([]byte(string(decodedPayload)+"}"), &payloadStruct)
	if errPay != nil {
		return nil, nil, errPay
	}

	return headerStruct, payloadStruct, nil
}

// Check if username - namespace map exist in DB.
func NamespaceExist(nameSpace string) (bool, error) {
	return false, nil
}
