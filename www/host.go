package www

import(
	"log"
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
)

var (

	// bucket name
	bucketName = util.Env("BUCKET_NAME", "")

	// google storage credential file path
	googleStorageCredPath = util.Env("GOOGLE_STORAGE_CREDENTIALS", "")
)

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
func CreateGoogleClients() *http.Client {

	// get google storage crendentials
	data, err := ioutil.ReadFile(googleStorageCredPath)
	if err != nil {
		util.Println("Failed to read storage credential from ", "GOOGLE_STORAGE_CREDENTIALS environment")
	    log.Fatal(err)
	}
	
	scope := "https://www.googleapis.com/auth/devstorage.full_control"
	conf, err := google.JWTConfigFromJSON(data, scope)
	if err != nil {
	    log.Fatal(err)
	}

	// Initiate an http.Client
	gStorageClient := conf.Client(oauth2.NoContext)
	return gStorageClient;
}

func App(testMode, runSeed bool) (*echo.Echo) {

	// create new server and router
	router := echo.New()
	var v1 = router.Group("/v1")	
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	config.HandleError(router)

	// setup router
	configRouter(router, testMode)

	// bucket name must be set
	if strings.TrimSpace(bucketName) == "" {
		log.Fatal("BUCKET_NAME environment variable is unset")
	}

	// create google service clients
	gStorageClient := CreateGoogleClients()

	// add some data in global config
	config.C.Add("bucket_name", bucketName)

	// initialize controllers
	appCntrl 	:= lib.NewAppController()
	policyCntrl := lib.NewPolicyController(appCntrl)
	mintCntrl   := lib.NewMintController(gStorageClient)

	// app management related route
	router.GET("/", extend.Handle(appCntrl.Index), UseAuthPolicy(policyCntrl)...)

	// currency processing route
	var mintRoute = v1.Group("/mint")
	mintRoute.POST("/new", extend.Handle(mintCntrl.Process))

	return router
}