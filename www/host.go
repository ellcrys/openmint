package www

import(
	"os"
	"log"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/lib"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/ellcrys/util"
	"golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    vision "google.golang.org/api/vision/v1"
)

var (

	// bucket name
	bucketName 				= util.Env("BUCKET_NAME", "")

	// google storage credential file path
	googleStorageCredPath 	= util.Env("GOOGLE_STORAGE_CREDENTIALS", "")
	googleVisionCredPath 	= util.Env("GOOGLE_VISION_CREDENTIALS", "")

	// Config params
	configHost 				= util.Env("CONFIG_HOST", "")
	configAuthToken 		= util.Env("CONFIG_AUTH_TOKEN", "")

	// mongo params
	MongoDBHosts 			= ""
	MongoUsername 			= ""
	MongoPassword 			= ""
	MongoDatabase 			= ""

	// mongo collections
	CurrencyColName  		= 	util.Env("MONGO_CURRENCY_COL", "currency")
)

// fetch application config
func fetchConfig() {

	names := []string{ 
		"MONGO_DB_HOST", 
		"MONGO_DB_NAME", 
		"MONGO_USERNAME", 
		"MONGO_PASSWORD", 
	}

	keys := strings.Join(names, ",")
	var url = fmt.Sprintf("%s/v1/keys/%s", configHost, keys)
	var headers = map[string]string{
		"Authorization": "Bearer " + configAuthToken,
	}
	
	resp, err := util.NewGetRequest(url, headers)
	if err != nil {
		log.Println("failed to fetch config")
		os.Exit(1)
	}

	defer resp.Body.Close()
    contents, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("failed to read response from config host")
		os.Exit(1)
    }

    data, err := util.DecodeJSONToMap(string(contents))
    if err != nil {
    	log.Println("failed to parse malformed config response")
		os.Exit(1)
    }

    if resp.StatusCode != 200 {
    	log.Println("failed: " + data["message"].(string))
		os.Exit(1)
    }

    configs 			:= data["values"].(map[string]interface{})
    MongoDBHosts 		= configs["MONGO_DB_HOST"].(string)
    MongoUsername 		= configs["MONGO_USERNAME"].(string)
    MongoPassword 		= configs["MONGO_PASSWORD"].(string)
    MongoDatabase 		= configs["MONGO_DB_NAME"].(string)
} 

// setup middleware, logger etc
func configRouter(router *echo.Echo, testMode bool) {
	if testMode {
		log.SetOutput(ioutil.Discard)
		return
	}
}

// Defines and return an array of policies to pass
// to routes that require authentication to access
func UseAuthPolicy(policyCntrl *lib.PolicyController) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		extend.MiddlewareHandle(policyCntrl.Authenticate),
	}
}

// Creates google cloud storage client
// and any other client required.
func CreateGoogleClients() (*http.Client, *http.Client) {

	// get google storage crendentials
	gStorageCredData, err := ioutil.ReadFile(googleStorageCredPath)
	if err != nil {
		util.Println("Failed to read storage credential from ", "GOOGLE_STORAGE_CREDENTIALS environment variable")
	    log.Fatal(err)
	}
	
	scope := "https://www.googleapis.com/auth/devstorage.full_control"
	conf, err := google.JWTConfigFromJSON(gStorageCredData, scope)
	if err != nil {
	    log.Fatal(err)
	}

	// Initiate an http.Client
	gStorageClient := conf.Client(oauth2.NoContext)

	// get google vision credentials
	gVisionCredData, err := ioutil.ReadFile(googleVisionCredPath)
	if err != nil {
		util.Println("Failed to read vision credential from ", "GOOGLE_VISION_CREDENTIALS environment variable")
	    log.Fatal(err)
	}

	// parse client credential file
	config, err := google.JWTConfigFromJSON(gVisionCredData, vision.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	gVisionClient := config.Client(oauth2.NoContext)

	return gStorageClient, gVisionClient;
}

// Fatally exits if an environment variable is unset
func requiresEnv(envName string) {
	if strings.TrimSpace(util.Env(envName, "")) == "" {
		log.Fatal(envName + " environment variable is unset")
	}
}

func App(testMode, runSeed bool) (*echo.Echo) {

	// create new server and router
	router := echo.New()
	var v1 = router.Group("/v1")	
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	config.HandleError(router)

	// fetch config
	fetchConfig()

	// setup router
	configRouter(router, testMode)

	// bucket name must be set
	requiresEnv("BUCKET_NAME")
	requiresEnv("CONFIG_HOST")

	// create google service clients
	gStorageClient, gVisionClient := CreateGoogleClients()

	// add some data in global config
	config.C.Add("bucket_name", bucketName)
	config.C.Add("mongo_database", MongoDatabase)
	config.C.Add("mongo_currency_collection", CurrencyColName)

	// mongo connection
	mongoSession, err := GetMongoSession(MongoDBHosts, MongoDatabase, MongoUsername, MongoPassword)
	if err != nil {
		util.Println("could not connect to mongo database")
		os.Exit(1)
	}

	// initialize controllers
	appCntrl 	:= lib.NewAppController()
	policyCntrl := lib.NewPolicyController(appCntrl)
	mintCntrl   := lib.NewMintController(mongoSession, gStorageClient, gVisionClient)

	// app management related route
	router.GET("/", extend.Handle(appCntrl.Index), UseAuthPolicy(policyCntrl)...)

	// currency processing route
	var mintRoute = v1.Group("/mint")
	mintRoute.POST("/new", extend.Handle(mintCntrl.Process))

	return router
}