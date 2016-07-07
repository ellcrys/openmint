package lib

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"strings"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/util"
	storage "google.golang.org/api/storage/v1"
	vision "google.golang.org/api/vision/v1"
	"gopkg.in/mgo.v2"
)

const NotMultipart = "request Content-Type isn't multipart/form-data"

type MintController struct {
	mongoSession   *mgo.Session
	storageService *storage.Service
	visionService  *vision.Service
}

// Create storage service.
func createStorageService(client *http.Client) *storage.Service {
	service, err := storage.New(client)
	if err != nil {
		log.Fatalf("Unable to create storage service: %v", err)
	}
	return service
}

// Create vision service
func createVisionService(client *http.Client) *vision.Service {
	service, err := vision.New(client)
	if err != nil {
		log.Fatalf("Unable to create vision service: %v", err)
	}
	return service
}

// Create a new controller instance
func NewMintController(mongoSession *mgo.Session, storageClient *http.Client, visionClient *http.Client) *MintController {
	storageService := createStorageService(storageClient)
	visionService := createVisionService(visionClient)
	return &MintController{mongoSession, storageService, visionService}
}

// Store image in google cloud storage
func (mc *MintController) SaveImage(file *multipart.FileHeader) (string, error) {

	// create the object
	object := &storage.Object{Name: util.RandString(32) + path.Ext(file.Filename)}
	curFile, err := file.Open()
	defer curFile.Close()
	if err != nil {
		return "", errors.New("failed to open currency image. " + err.Error())
	}

	// add object to bucket
	bucketName := config.C.GetString("bucket_name")
	if res, err := mc.storageService.Objects.Insert(bucketName, object).Media(curFile).Do(); err == nil {
		return res.Name, nil
	} else {
		return "", errors.New("failed to create object in cloud storage. " + err.Error())
	}
}

// Process a currency
func (mc *MintController) ProcessCurrency(curCode, curDenom, imageName, currencyId string) {

	// get currency language
	lang := GetCurrencyLang(curCode)

	// process currency image. Get labels and text extracts.
	gcsImageUri := fmt.Sprintf("gs://%s/%s", config.C.GetString("bucket_name"), imageName)
	imgProcRes, err := ProcessImage(lang, mc.visionService, gcsImageUri)
	if err != nil {
		if err = models.Currency.UpdateStatus(mc.mongoSession, currencyId, "failed"); err != nil {
			util.Println("Failed to update currency status")
		}
		return
	}

	// ensure we got a response
	if len(imgProcRes.Responses) == 0 {
		util.Println("Got not annotation response for currency#" + currencyId)
		return
	}

	// extract tokens from text annotation
	var tokens = ProcessText(imgProcRes.Responses[0].TextAnnotations)

	result, err := ProcessMoney(curCode, curDenom, tokens, imgProcRes.Responses[0].LabelAnnotations)
	util.Println(result, err)
	if err != nil {
		if err = models.Currency.UpdateStatus(mc.mongoSession, currencyId, "failed"); err != nil {
			util.Println("Failed to update currency status")
		}
		return
	}
}

// Process a new currency
func (mc *MintController) Process(c *extend.Context) error {

	// TODO: get user id from session
	var userId = models.NewId()
	var imageName string

	// get currency image
	currencyImg, err := c.Echo().FormFile("currency_image")
	if err != nil {
		util.Println(err)
		if err == http.ErrMissingFile || err.Error() == NotMultipart {
			return config.NewHTTPError(c.Lang(), 400, "e002")
		}
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// get currency code and ensure it is valid
	curCode := c.Echo().FormValue("currency_code")
	if len(strings.TrimSpace(curCode)) == 0 {
		return config.NewHTTPError(c.Lang(), 400, "e003")
	}

	// currency code must be known
	if !IsValidCode(strings.ToUpper(curCode)) {
		return config.NewHTTPError(c.Lang(), 400, "e004")
	}

	// currency code must have meta definition
	if !util.InStringSlice(GetDefinedCurrencies(), strings.ToUpper(curCode)) {
		return config.NewHTTPError(c.Lang(), 400, "e009")
	}

	// currency denomination (optional)
	curDenom := c.Echo().FormValue("currency_denom")
	if curDenom != "" && !util.InStringSlice(GetCurrencyDenoms(curCode), curDenom) {
		return config.NewHTTPError(c.Lang(), 400, "e005")
	}

	// save currency image
	if imageName, err = mc.SaveImage(currencyImg); err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// create currency entry
	currency := &models.CurrencyModel{
		Id:      models.NewId(),
		UserId:  userId,
		ImageID: imageName,
		Code:    curCode,
	}

	if err = models.Currency.Create(mc.mongoSession, currency); err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// process image asynchronously
	go mc.ProcessCurrency(curCode, curDenom, imageName, currency.Id.Hex())

	return c.JSON(201, extend.H{
		"status": "processing",
		"id":     currency.Id.Hex(),
	})
}
