package www

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/lib"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/util"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	vision "google.golang.org/api/vision/v1"
	"gopkg.in/mgo.v2"
)

var (

	// bucket name
	bucketName = util.Env("BUCKET_NAME", "")

	// google storage credential file path
	googleStorageCredPath = util.Env("GOOGLE_STORAGE_CREDENTIALS", "")
	googleVisionCredPath  = util.Env("GOOGLE_VISION_CREDENTIALS", "")

	// Config params
	configHost      = util.Env("CONFIG_HOST", "")
	configAuthToken = util.Env("CONFIG_AUTH_TOKEN", "")

	// mongo params
	MongoDBHosts  = util.Env("MONGO_DB_HOST", "")
	MongoUsername = util.Env("MONGO_USERNAME", "")
	MongoPassword = util.Env("MONGO_PASSWORD", "")
	MongoDatabase = util.Env("MONGO_DB_NAME", "")

	// redis params
	RedisURL      = ""
	RedisPassword = ""
	RedisDatabase = util.Env("REDIS_DB", "0")

	// mongo collections
	CurrencyColName      = util.Env("MONGO_CURRENCY_COL", "currency")
	CloudMintUserColName = util.Env("MONGO_CLOUDMINT_USER_COL", "cloudmint_user")
	TwitterAuthColName   = util.Env("MONGO_TWITTER_AUTH_COL", "twitter_auth")

	// others
	HMACKey             = util.Env("HMAC_KEY", "")
	FBAppId             = util.Env("FB_APP_ID", "")
	FBAppToken          = util.Env("FB_APP_TOKEN", "")
	TwitterConKey       = util.Env("TWITTER_CONSUMER_KEY", "")
	TwitterConSecret    = util.Env("TWITTER_CONSUMER_SECRET", "")
	MaxVotes            = util.Env("MAX_VOTES", "3")
	VoteSessionDuration = util.Env("VOTE_SESSION_DURATION", "1200")
)

// fetch application config
func fetchConfig() {

	names := []string{
		"MONGO_DB_HOST",
		"MONGO_DB_NAME",
		"MONGO_USERNAME",
		"MONGO_PASSWORD",
		"REDIS_URL",
		"REDIS_PWD",
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
		log.Println("failed to fetch config. ", data)
		os.Exit(1)
	}

	configs := data["values"].(map[string]interface{})
	MongoDBHosts = configs["MONGO_DB_HOST"].(string)
	MongoUsername = configs["MONGO_USERNAME"].(string)
	MongoPassword = configs["MONGO_PASSWORD"].(string)
	MongoDatabase = configs["MONGO_DB_NAME"].(string)
	RedisURL = configs["REDIS_URL"].(string)
	RedisPassword = configs["REDIS_PWD"].(string)
}

// setup middleware, logger etc
func configRouter(router *echo.Echo, testMode bool) {
	if testMode {
		// ... do test setup here
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

	return gStorageClient, gVisionClient
}

// Fatally exits if an environment variable is unset
func requiresEnv(envName string) {
	if strings.TrimSpace(util.Env(envName, "")) == "" {
		log.Fatal(envName + " environment variable is unset")
	}
}

func App(testMode, runSeed bool) (*echo.Echo, *mgo.Session) {

	// create new server and router
	router := echo.New()
	var v1 = router.Group("/v1")
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	config.HandleError(router)

	// fetch config
	if !testMode {
		requiresEnv("CONFIG_HOST")
		fetchConfig()
	}

	// setup router
	configRouter(router, testMode)

	// bucket name must be set
	requiresEnv("BUCKET_NAME")
	requiresEnv("HMAC_KEY")

	// create google service clients
	gStorageClient, gVisionClient := CreateGoogleClients()

	// add some data in global config
	config.C.Add("bucket_name", bucketName)
	config.C.Add("mongo_database", MongoDatabase)
	config.C.Add("mongo_currency_collection", CurrencyColName)
	config.C.Add("mongo_cloudmint_user_col", CloudMintUserColName)
	config.C.Add("mongo_twitter_auth_col", TwitterAuthColName)
	config.C.Add("hmac_key", HMACKey)
	config.C.Add("fb_app_token", FBAppToken)
	config.C.Add("fb_app_id", FBAppId)
	config.C.Add("twitter_con_key", TwitterConKey)
	config.C.Add("twitter_con_secret", TwitterConSecret)
	config.C.Add("max_votes", MaxVotes)
	config.C.Add("vote_session_duration", VoteSessionDuration)

	// mongo connection
	mongoSession, err := GetMongoSession(MongoDBHosts, MongoDatabase, MongoUsername, MongoPassword)
	if err != nil {
		util.Println("could not connect to mongo database -> ", err)
		os.Exit(1)
	} else {
		models.Currency.EnsureIndex(mongoSession)
		models.User.EnsureIndex(mongoSession)
	}

	// redis connection
	redisDB, _ := strconv.Atoi(RedisDatabase)
	redisPool := GetRedisPool(RedisURL, RedisPassword, redisDB)
	conn := redisPool.Get()
	defer conn.Close()
	if _, err := redisPool.Get().Do("PING"); err != nil {
		util.Println("could not connect to redis database", err)
		conn.Close()
		os.Exit(1)
	}

	// initialize controllers
	appCntrl := lib.NewAppController()
	policyCntrl := lib.NewPolicyController(mongoSession)
	mintCntrl := lib.NewMintController(mongoSession, redisPool, gStorageClient, gVisionClient)
	userCntrl := lib.NewUserController(mongoSession)
	authCntrl := lib.NewAuthController(mongoSession)

	// app management related route
	router.GET("/", extend.Handle(appCntrl.Index), UseAuthPolicy(policyCntrl)...)

	// auth route
	var authRoute = v1.Group("/auth")
	authRoute.POST("/social", extend.Handle(authCntrl.SocialAuth))
	authRoute.GET("/twitter/request_token", extend.Handle(authCntrl.GetTwitterRequestToken))
	authRoute.GET("/twitter/cb", extend.Handle(authCntrl.TwitterCallback))
	authRoute.GET("/twitter/done", extend.Handle(authCntrl.Blank))
	authRoute.GET("/me", extend.Handle(authCntrl.GetUser), UseAuthPolicy(policyCntrl)...)

	// user route
	var userRoute = v1.Group("/users")
	userRoute.GET("/currencies", extend.Handle(userCntrl.GetCurrencies), UseAuthPolicy(policyCntrl)...)

	// currency processing route
	var mintRoute = v1.Group("/mint")
	mintRoute.POST("/new", extend.Handle(mintCntrl.Process), UseAuthPolicy(policyCntrl)...)
	mintRoute.GET("/supported_currencies", extend.Handle(mintCntrl.GetSupportedCurrencies), UseAuthPolicy(policyCntrl)...)
	mintRoute.GET("/vote", extend.Handle(mintCntrl.GetVoteSession), UseAuthPolicy(policyCntrl)...)
	mintRoute.PUT("/vote", extend.Handle(mintCntrl.AddVote), UseAuthPolicy(policyCntrl)...)

	return router, mongoSession
}
