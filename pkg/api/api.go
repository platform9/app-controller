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

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"github.com/MicahParks/keyfunc"

	"github.com/gorilla/mux"

	"github.com/platform9/fast-path/pkg/db"
	"github.com/platform9/fast-path/pkg/knative"
	"github.com/platform9/fast-path/pkg/objects"
	"github.com/platform9/fast-path/pkg/options"
	"github.com/platform9/fast-path/pkg/util"
	corev1 "k8s.io/api/core/v1"

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
	zap.S().Info("***** Get Apps *****")
	// Validate the token, and get claims.
	claims, err := ValidateToken(r)
	if err != nil {
		if findStrInSlice(err.Error(), util.ErrorsToken) {
			zap.S().Errorf("Token validation Error: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		zap.S().Errorf("Error is: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch user information from claims
	userInfo, err := GetUserClaims(claims)
	if err != nil {
		zap.S().Errorf("Failed to get user information. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		zap.S().Errorf("Failed to get Namespace. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	appList, err := knative.GetApps(util.Kubeconfig, nameSpace)
	if err != nil {
		zap.S().Errorf("Error while listing app. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	zap.S().Infof("App List, successful. Space: %v", nameSpace)

	data := []byte(appList)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(data); err != nil {
		zap.S().Errorf("Error while responding over http. Error: %v", err)
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
	SecretName string `json:"secretname"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

/*
1. User fires appctl deploy command with name, image, token as bearer.
2. Validate the token, get user info.
3. If token is valid,
	4. Fetch Username - Namespace map in DB.
	5. CreateApp using above fetched namespace.
*/

func createApp(w http.ResponseWriter, r *http.Request) {
	zap.S().Info("***** Create App *****")
	// Validate the token, and get claims.
	claims, err := ValidateToken(r)
	if err != nil {
		if findStrInSlice(err.Error(), util.ErrorsToken) {
			zap.S().Errorf("Token validation Error: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		zap.S().Errorf("Error is: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch user information from claims
	userInfo, err := GetUserClaims(claims)
	if err != nil {
		zap.S().Errorf("Failed to get user information. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		zap.S().Errorf("Failed to get Namespace. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	app := App{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		zap.S().Errorf("Error while reading data in request body. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &app)
	if err != nil {
		zap.S().Errorf("Error while unmarhsalling request body data. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	zap.S().Debugf("app: %v\n", app)

	envVars := []corev1.EnvVar{}
	envVar := corev1.EnvVar{}
	for _, env := range app.Envs {
		envVar.Name = env.Key
		envVar.Value = env.Value
		envVars = append(envVars, envVar)
	}

	err = knative.CreateApp(util.Kubeconfig, app.Name, nameSpace, app.Image, envVars, app.Port,
				app.SecretName, app.UserName, app.Password)
	if err != nil {
		if err.Error() == util.MaxAppDeployError {
			zap.S().Errorf("Maximum App deployed limit reached!! Namespace: %v", nameSpace)
			w.WriteHeader(util.MaxAppDeployStatusCode)
			return
		} else if strings.Contains(err.Error(), util.Errors[0]) {
			zap.S().Errorf("Error while creating app. Error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		zap.S().Errorf("Error while creating app. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	zap.S().Infof("App Name: %v, Image: %v, created successfully in Space: %v", app.Name, app.Image, nameSpace)
	w.WriteHeader(http.StatusOK)
}

func getAppByName(w http.ResponseWriter, r *http.Request) {

	zap.S().Info("***** Get App by name *****")

	// Validate the token, and get claims.
	claims, err := ValidateToken(r)
	if err != nil {
		if findStrInSlice(err.Error(), util.ErrorsToken) {
			zap.S().Errorf("Token validation Error: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		zap.S().Errorf("Error is: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch user information from claims
	userInfo, err := GetUserClaims(claims)
	if err != nil {
		zap.S().Errorf("Failed to get user information. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		zap.S().Errorf("Failed to get Namespace. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	appName := vars["name"]

	appList, err := knative.GetAppByName(util.Kubeconfig, nameSpace, appName)
	if err != nil {
		zap.S().Errorf("Error while listing app. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	zap.S().Infof("Get app by name successful. Name: %v, Space: %v", appName, nameSpace)

	data := []byte(appList)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(data); err != nil {
		zap.S().Errorf("Error while responding over http. Error: %v", err)
	}
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	zap.S().Info("***** Delete App *****")

	// Validate the token, and get claims.
	claims, err := ValidateToken(r)
	if err != nil {
		if findStrInSlice(err.Error(), util.ErrorsToken) {
			zap.S().Errorf("Token validation Error: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		zap.S().Errorf("Error is: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch user information from claims
	userInfo, err := GetUserClaims(claims)
	if err != nil {
		zap.S().Errorf("Failed to get user information. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Get Namespace from DB
	nameSpace, err := GetNamespace(*userInfo)
	if err != nil {
		zap.S().Errorf("Failed to get Namespace. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	deleteAppName := vars["name"]

	zap.S().Debugf("vars : %v", vars)

	zap.S().Infof("Name: %s, space: %s", deleteAppName, nameSpace)

	errDel := knative.DeleteApp(util.Kubeconfig, nameSpace, deleteAppName)
	if errDel != nil {
		zap.S().Errorf("Error while deleting app. Error: %v", errDel)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	zap.S().Infof("Delete app successful. Name: %v, Space: %v", deleteAppName, nameSpace)
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
	zap.S().Info("***** Login *****")

	// Validate the token, and get claims.
	claims, err := ValidateToken(r)
	if err != nil {
		if findStrInSlice(err.Error(), util.ErrorsToken) {
			zap.S().Errorf("Token validation Error: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		zap.S().Errorf("Error is: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch user information from claims
	userInfo, err := GetUserClaims(claims)
	if err != nil {
		zap.S().Errorf("Failed to get user information. Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
			zap.S().Errorf("DB Error: %v", errDB)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if userDB.Name == userInfo.NickName {
			UserExists = true
		}
	} else {
		errDB := que.GetUserByEmail(userInfo.Email, &userDB)
		if errDB != nil {
			zap.S().Errorf("Get user info from DB. Error: %v", errDB)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if userDB.Email == userInfo.Email {
			UserExists = true
		}
	}

	var NameSpace, createdNS string
	if !UserExists {
		zap.S().Info("User doesn't exist's in DB, starting creation of namespace.")
		if strings.Contains(userInfo.Sub, "github") {
			NameSpace = userInfo.NickName + CreateRandomCode(6)
		} else {
			NameSpace = strings.Split(userInfo.Email, "@")[0] + CreateRandomCode(6)
		}

		// Check if namespace is valid. bcz only regex valid "^*[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
		if util.RegexValidate(NameSpace) {
			zap.S().Debug("Its a Valid Namespace")
		} else {
			/* If namespace is invalid:
			1. Remove all special characters other than [a-zA-Z0-9] from namespace.
			2. Convert if any uppercase characters exists in string to lowercase.
			*/
			zap.S().Debug("Namespace given is not valid, so formating as per valid regex.")
			NameSpace, err = RemoveSpecialChars(NameSpace)
			if err != nil {
				zap.S().Errorf("Notable to remove special characters from given string. Error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			NameSpace = NameSpace + CreateRandomCode(6)
		}

		// Create new namespace.
		createdNS, err = CreateNamespace(NameSpace)
		if err != nil {
			zap.S().Errorf("Failed to create a namespace with name %v. Error: ", createdNS, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		zap.S().Infof("Successfully created namespace %v", createdNS)

		// Add Userinfo to DB.
		var user objects.User
		user.Name = userInfo.NickName
		user.Email = userInfo.Email
		user.Space = createdNS

		errDB := que.AddUser(&user)
		if errDB != nil {
			zap.S().Errorf("Adding user information to DB. Error: %v", errDB)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		zap.S().Infof("Added user information to DB. Name: %v, Email: %v, Space: %v", userInfo.NickName, userInfo.Email, createdNS)
	} else {
		//Get Namespace from DB
		nameSpace, err := GetNamespace(*userInfo)
		if err != nil {
			zap.S().Errorf("Failed to get Namespace. Error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		zap.S().Infof("Login successful. Existing-User: %v, Email: %v, Space: %v", userInfo.NickName, userInfo.Email, nameSpace)
	}
	w.WriteHeader(http.StatusOK)
}

// To create a new namespace as part of login.
func CreateNamespace(nameSpace string) (string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", util.Kubeconfig)
	if err != nil {
		zap.S().Errorf("Error building config: %v", err)
		return "", err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		zap.S().Errorf("Error creating kubernetes config: %v", err)
		return "", err
	}

	// Check if the creating namespace already exist.
	_, errGetNs := clientset.CoreV1().Namespaces().Get(context.TODO(), nameSpace, metav1.GetOptions{})
	if errGetNs == nil {
		nameSpace = nameSpace + CreateRandomCode(4)
		zap.S().Debugf("Namespace already exists. Generating new namespace: %v", nameSpace)
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
		zap.S().Errorf("Failed to create a new namespace %v. Error: %v", nameSpace, errCreate)
		return "", errCreate
	}
	return nameSpace, nil
}

// Get the namespace for user, from DB.
func GetNamespace(userInfo UserInfo) (string, error) {
	//Database User object.
	var userDB objects.User
	que := db.Get()
	if strings.Contains(userInfo.Sub, "github") {
		errDB := que.GetUserByName(userInfo.NickName, &userDB)
		if errDB != nil {
			zap.S().Errorf("DB Error: %v", errDB)
			return "", fmt.Errorf("Failed to get Namespace. Error: %v", errDB)
		}
		if userDB.Space != "" {
			zap.S().Debugf("Namespace found is: %v", userDB.Space)
			return userDB.Space, nil
		}
	} else {
		errDB := que.GetUserByEmail(userInfo.Email, &userDB)
		if errDB != nil {
			zap.S().Errorf("Get user info from DB. Error: %v", errDB)
			return "", fmt.Errorf("Failed to get Namespace. Error: %v", errDB)
		}
		if userDB.Space != "" {
			zap.S().Debugf("Namespace found is: %v", userDB.Space)
			return userDB.Space, nil
		}
	}
	return "", fmt.Errorf("Failed to get Namespace")
}

// Create a random code of given length.
func CreateRandomCode(lenCode int) string {
	var letter = []rune(util.AllCharSet)

	rand.Seed(time.Now().UnixNano())
	code := make([]rune, lenCode)
	for i := range code {
		code[i] = letter[rand.Intn(len(letter))]
	}
	return string(code)
}

// Remove all special characters and convert to lowercase alphanumeric string.
func RemoveSpecialChars(specialChar string) (string, error) {
	regex, err := regexp.Compile(util.NoSpecialChar)
	if err != nil {
		zap.S().Errorf("Error while removing special characters: %v", err)
		return "", fmt.Errorf("%v\n", err)
	}
	formattedString := regex.ReplaceAllString(specialChar, "")
	return strings.ToLower(formattedString), nil
}

// Get the UserInfo from claims.
func GetUserClaims(claims jwt.Claims) (*UserInfo, error) {

	var user UserInfo
	errStru := mapstructure.Decode(claims, &user)
	if errStru != nil {
		zap.S().Error("Failed to convert map to struct. Error: %v", errStru)
		return &UserInfo{}, fmt.Errorf("%v", errStru)
	}

	zap.S().Infof("The User info is %+v", user)
	return &user, nil
}

// Fetch the token and validate it.
func ValidateToken(r *http.Request) (jwt.Claims, error) {

	// Fetch the token.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return jwt.MapClaims{}, fmt.Errorf(util.ErrorsToken[2])
	}

	bearerToken := strings.Split(authHeader, "Bearer ")
	if len(bearerToken) != 2 {
		return jwt.MapClaims{}, fmt.Errorf(util.ErrorsToken[2])
	}

	// Create the JWKS from the resource at the given URL.
	jwks, err := keyfunc.Get(options.GetJWKSURL(), keyfunc.Options{})
	if err != nil {
		zap.S().Errorf("Failed to create JWKS from URL.\nError: %s", err.Error())
		return jwt.MapClaims{}, fmt.Errorf("Failed to create JWKS from URL. Error: %s", err.Error())
	}

	// Parse the token.
	token, err := jwt.Parse(bearerToken[1], jwks.Keyfunc)
	if err != nil {
		zap.S().Errorf("Error is %v\n", err)
		if err.Error() == util.ErrorsToken[0] {
			return jwt.MapClaims{}, fmt.Errorf(util.ErrorsToken[0])
		}
		zap.S().Errorf("Falied to parse token. Error: %s", err.Error())
		return jwt.MapClaims{}, fmt.Errorf("Falied to parse token. Error: %s", err.Error())
	}

	//Fetch Claims
	claims, _ := token.Claims.(jwt.MapClaims)

	// Audiance validation i.e if audiance == auth0 clientID
	if !token.Valid || claims["aud"] != options.GetAuth0ClientId() {
		return jwt.MapClaims{}, fmt.Errorf(util.ErrorsToken[1])
	}

	return claims, nil
}

// To check if a string exists in a slice of strings.
func findStrInSlice(msg string, list []string) bool {
	for _, value := range list {
		if value == msg {
			return true
		}
	}
	return false
}
