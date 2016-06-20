package lib

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"strconv"
	"mime/multipart"
	"fmt"
	"path"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/util"
	"gopkg.in/mgo.v2"
	"github.com/ellcrys/openmint/extend"
	storage "google.golang.org/api/storage/v1"
)


type MintController struct {
	mongoSession 	*mgo.Session
	storageService 	*storage.Service
}

// Create storage service.
func createStorageService(storageClient *http.Client) *storage.Service {
	service, err := storage.New(storageClient)
    if err != nil {
        log.Fatalf("Unable to create storage service: %v", err)
    }
    return service
}

// Create a new controller instance
func NewMintController(mongoSession *mgo.Session, storageClient *http.Client) *MintController {
	storageService := createStorageService(storageClient)
	return &MintController{ mongoSession, storageService }
}

// Store image in google cloud storage
func (mc *MintController) SaveImage(file *multipart.FileHeader) error {
		
	// create the object
	object := &storage.Object{ Name: util.RandString(32) + path.Ext(file.Filename) }
	curFile, err := file.Open()
	defer curFile.Close()
	if err != nil {
		return errors.New("failed to open currency image. " + err.Error())
	}

	// add object to bucket
	bucketName := config.C.GetString("bucket_name")
	if res, err := mc.storageService.Objects.Insert(bucketName, object).Media(curFile).Do(); err == nil {
        fmt.Printf("Created object %v at location %v\n\n", res.Name, res.SelfLink)
        return nil
    } else {
        return errors.New("failed to create object in cloud storage. " + err.Error())
    }
}

// Process a new currency
func (mc *MintController) Process(c *extend.Context) error {

	// get currency image
	currencyImg, err := c.Echo().FormFile("currency_image")
	if err != nil {
		if err == http.ErrMissingFile {
			return config.NewHTTPError(c.Lang(), 400, "e002")
		}
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// get currency code and ensure it is valid
	curCode := c.Echo().FormValue("currency_code")
	if len(strings.TrimSpace(curCode)) == 0 {
		return config.NewHTTPError(c.Lang(), 400, "e003")
	} else if !config.IsValidCode(strings.ToUpper(curCode)) {
		return config.NewHTTPError(c.Lang(), 400, "e004")
	}

	// get currency denomination
	curDenomination := c.Echo().FormValue("currency_denom")
	if len(strings.TrimSpace(curDenomination)) == 0 {
		return config.NewHTTPError(c.Lang(), 400, "e005")
	}
	
	// convert currency denomination to integer
	curDenominationInt, err := strconv.Atoi(curDenomination)
	if err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	} else if !config.IsValidDenomination(strings.ToUpper(curCode), curDenominationInt) {
		return config.NewHTTPError(c.Lang(), 400, "e006")
	}

	// save currency image
	if err = mc.SaveImage(currencyImg); err != nil {
		util.Println(err)
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	return c.String(200, "Hello World!")
}
