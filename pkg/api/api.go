package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/platform9/fast-path/pkg/db"
	corev1 "k8s.io/api/core/v1"
	"github.com/platform9/fast-path/pkg/knative"
	"github.com/platform9/fast-path/pkg/objects"
	"github.com/platform9/fast-path/pkg/util"

	"context"

	"github.com/mitchellh/mapstructure"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type UserInfo struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	NickName string  `json:"nickname"`
	Aud      string  `json:"aud"`
	Sub      string  `json:"sub"`
	Exp      float64 `json:"exp"`
}

// New returns new API router for fast-path
func New() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/v1/apps", getApp).Methods("GET")
	r.HandleFunc("/v1/apps/{name}", getAppByName).Methods("GET")
	r.HandleFunc("/v1/apps", createApp).Methods("POST")
	r.HandleFunc("/v1/apps/login", loginApp).Methods("POST")
	r.HandleFunc("/v1/apps/{name}", deleteApp).Methods("DELETE")

	return r
}

func getApp(w http.ResponseWriter, r *http.Request) {

	// Validate a token, and get user claims.
	userInfo, validToken, expired, err := ValidateToken(r)

	// Continue only if token is valid, token is not expired and no error occured while validating token.
	if !(validToken && !expired && err == nil) {
		fmt.Printf("Token Validation Error: %v\n", err)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		fmt.Printf("Failed to get Namespace. Error: %v\n", err)
		return
	}

	appList, err := knative.GetApps(util.Kubeconfig, nameSpace)

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
	Image string `json:"image"`
	Port  string `json:"port"`
	Envs  []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"envs"`
}

/*
1. User fires appctl deploy command with name, image, token as bearer.
2. Validate the token, get user name, token expiry.
3. If token is expired return message token expired. Else continue..
	4. Fetch Username - Namespace map in DB. Use this as nameSpace.
	5. CreateApp.
*/

func createApp(w http.ResponseWriter, r *http.Request) {
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

	fmt.Printf("app: %v", app)

	// Validate a token, and get user claims.
	userInfo, validToken, expired, err := ValidateToken(r)

	// Continue only if token is valid, token is not expired and no error occured while validating token.
	if !(validToken && !expired && err == nil) {
		fmt.Printf("Token Validation Error: %v\n", err)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		fmt.Printf("Failed to get Namespace. Error: %v", err)
		return
	}

	envVars := []corev1.EnvVar{}
	envVar := corev1.EnvVar{}
	for _, env := range app.Envs {
		envVar.Name = env.Key
		envVar.Value = env.Value
		envVars = append(envVars, envVar)
	}

	err = knative.CreateApp(util.Kubeconfig, app.Name, nameSpace, app.Image, envVars, app.Port)
	if err != nil {
		log.Error(err, "while creating app")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getAppByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["name"]

	// Validate a token, and get user claims.
	userInfo, validToken, expired, err := ValidateToken(r)

	// Continue only if token is valid, token is not expired and no error occured while validating token.
	if !(validToken && !expired && err == nil) {
		fmt.Printf("Token Validation Error: %v\n", err)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		fmt.Printf("Failed to get Namespace. Error: %v", err)
		return
	}

	appList, err := knative.GetAppByName(util.Kubeconfig, nameSpace, appName)

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
	deleteAppName := vars["name"]

	// Validate a token, and get user claims.
	userInfo, validToken, expired, err := ValidateToken(r)

	// Continue only if token is valid, token is not expired and no error occured while validating token.
	if !(validToken && !expired && err == nil) {
		fmt.Printf("Token Validation Error: %v\n", err)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		fmt.Printf("Failed to get Namespace. Error: %v", err)
		return
	}

	fmt.Printf("vars : %v\n", vars)

	fmt.Printf("Name: %s, space: %s", deleteAppName, nameSpace)

	errdel := knative.DeleteApp(util.Kubeconfig, nameSpace, deleteAppName)
	if errdel != nil {
		log.Error(errdel, "while deleting app")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/*
1. Obtain token from header.
2. GetUserInfo after validating the token signature.
3. Check if user exists in DB.
	4. If exists then check expiry and do necessary action if exipred.
	5. Else, create a userNamespace and update the DB.
*/
func loginApp(w http.ResponseWriter, r *http.Request) {

	// Validate a token, and get user claims.
	userInfo, validToken, expired, err := ValidateToken(r)

	// Continue only if token is valid, token is not expired and no error occured while validating token.
	if !(validToken && !expired && err == nil) {
		fmt.Printf("Token Validation Error: %v\n", err)
		return
	}

	//Database User object.
	var userDB objects.User
	que := db.Get()

	// Check if user exists in DB.
	var UserExists bool = false
	if strings.Contains(userInfo.Sub, "github") {
		errDB := que.GetUserByName(userInfo.NickName, &userDB)
		if errDB != nil {
			fmt.Printf("DB Error: %v\n", errDB)
		}
		if userDB.Name == userInfo.NickName {
			UserExists = true
		}
	} else {
		errDB := que.GetUserByEmail(userInfo.Email, &userDB)
		if errDB != nil {
			fmt.Printf("Get user info from DB. Error: %v\n", errDB)
		}
		if userDB.Email == userInfo.Email {
			UserExists = true
		}
	}

	var NameSpace string
	if !UserExists {

		if strings.Contains(userInfo.Sub, "github") {
			NameSpace = userInfo.NickName + CreateRandomCode(6)
		} else {
			NameSpace = strings.Split(userInfo.Email, "@")[0] + CreateRandomCode(6)
		}

		// Check if namespace is valid. bcz only regex valid "^*[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
		if util.RegexValidate(NameSpace) {
			fmt.Printf("Its a Valid Namespace\n")
		} else {
			/* If namespace is invalid:
			1. Remove all special characters other than [a-zA-Z0-9] from namespace.
			2. Convert if any uppercase characters exists in string to lowercase.
			*/
			fmt.Printf("Namespace given is not valid, so formating as per valid regex.\n")
			NameSpace, err = RemoveSpecialChars(NameSpace)
			if err != nil {
				fmt.Printf("Notable to remove special characters from given string. Error: %v", err)
			}
			NameSpace = NameSpace + CreateRandomCode(6)
		}

		// Create new namespace.
		errNamespace := CreateNamespace(NameSpace)
		if errNamespace != nil {
			fmt.Printf("Failed to create a namespace with name %v\n", NameSpace)
			return
		}

		fmt.Printf("Successfully created namespace\n")

		// Add Userinfo to DB.
		var user objects.User
		user.Name = userInfo.NickName
		user.Email = userInfo.Email
		user.Space = NameSpace

		errDB := que.AddUser(&user)
		if err != nil {
			fmt.Printf("Adding Info to DB. Error: %v\n", errDB)
		}
	} else {
		// User already exists in DB, then check expiry of token.
		expiryTime := time.Unix(int64(userInfo.Exp), 0)
		if expiryTime.Before(time.Now()) {
			fmt.Printf("Login expired. Please login again using command `appctl login`\n")
		}
	}
	w.WriteHeader(http.StatusOK)
}

func CreateNamespace(nameSpace string) error {
	config, err := clientcmd.BuildConfigFromFlags("", util.Kubeconfig)
	if err != nil {
		fmt.Printf("Error building config: %v", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating kubernetes config: %v", err)
		return err
	}

	// Check if the creating namespace already exist.
	_, errGetNs := clientset.CoreV1().Namespaces().Get(context.TODO(), nameSpace, metav1.GetOptions{})
	if errGetNs == nil {
		nameSpace = nameSpace + CreateRandomCode(4)
		fmt.Printf("Namespace already exists. Generating new namespace: %v\n", nameSpace)
	}

	// Namespace metaobject.
	nsName := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nameSpace,
		},
	}

	//Create a namespace.
	_, errCreate := clientset.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
	if errCreate != nil {
		fmt.Printf("Failed to create a new namespace %v\n", nameSpace)
		return errCreate
	}
	return nil
}

// Get the namespace for user, from DB.
func GetNamespace(userInfo UserInfo) (string, error) {
	//Database User object.
	var userDB objects.User
	que := db.Get()
	if strings.Contains(userInfo.Sub, "github") {
		errDB := que.GetUserByName(userInfo.NickName, &userDB)
		if errDB != nil {
			fmt.Printf("DB Error: %v\n", errDB)
			return "", fmt.Errorf("Failed to get Namespace. Error: %v", errDB)
		}
		if userDB.Space != "" {
			return userDB.Space, nil
		}
	} else {
		errDB := que.GetUserByEmail(userInfo.Email, &userDB)
		if errDB != nil {
			fmt.Printf("Get user info from DB. Error: %v\n", errDB)
			return "", fmt.Errorf("Failed to get Namespace. Error: %v", errDB)
		}
		if userDB.Space != "" {
			return userDB.Space, nil
		}
	}
	return "", fmt.Errorf("Failed to get Namespace")
}

func CreateRandomCode(lenCode int) string {
	var letter = []rune(util.AllCharSet)

	rand.Seed(time.Now().UnixNano())
	code := make([]rune, lenCode)
	for i := range code {
		code[i] = letter[rand.Intn(len(letter))]
	}
	return string(code)
}

// Check if token is expired or not.
func CheckTokenExpired(expiry float64) bool {
	// User already exists in DB.
	expiryTime := time.Unix(int64(expiry), 0)
	if expiryTime.Before(time.Now()) {
		fmt.Printf("Login expired. Please login again using command `appctl login`\n")
		return true
	}
	return false
}

// Remove all special characters and convert to lowercase alphanumeric string.
func RemoveSpecialChars(specialChar string) (string, error) {
	regex, err := regexp.Compile(util.NoSpecialChar)
	if err != nil {
		return "", fmt.Errorf("%v\n", err)
	}
	formattedString := regex.ReplaceAllString(specialChar, "")
	return strings.ToLower(formattedString), nil
}

// Validate token, check if token is expired or not and return user claims as userInfo.
func ValidateToken(r *http.Request) (*UserInfo, bool, bool, error) {

	// Fetch the token.
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]

	// Parse the token.
	tokens, err := jwt.Parse(reqToken, nil)
	if tokens == nil {
		fmt.Printf("Empty with error :%v", err)
	}

	//Fetch Claims
	claims, _ := tokens.Claims.(jwt.MapClaims)

	// Doing simple additional validation i.e if audiance == auth0 clientID
	if claims["aud"] != util.ClientId {
		return &UserInfo{}, false, false, fmt.Errorf("Token is invalid")
	}

	if expiry, ok := claims["exp"].(float64); ok {
		// Check if token is expired.
		if CheckTokenExpired(expiry) {
			fmt.Printf("Token is expired\n")
			return &UserInfo{}, true, true, fmt.Errorf("Token is Expired\n")
		}
	} else {
		return &UserInfo{}, true, false, fmt.Errorf("Can't fetch token expiryAt time.\n")
	}

	var user UserInfo
	errStru := mapstructure.Decode(claims, &user)
	if errStru != nil {
		fmt.Printf("Failed to convert map to struct\n")
		return &UserInfo{}, true, false, fmt.Errorf("%v", errStru)
	}

	fmt.Printf("The User info is %+v\n", user)
	return &user, true, false, nil
}
