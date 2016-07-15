package lib

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/util"
	storage "google.golang.org/api/storage/v1"
	vision "google.golang.org/api/vision/v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

// Store image in google cloud storage.
// File can be a multipart.File or an os.File.
// File handle will be closed after save is complete.
func (mc *MintController) SaveImage(file interface{}) (*storage.Object, error) {

	var fileToSave io.Reader
	var err error
	var object *storage.Object

	switch imgFile := file.(type) {

	case *multipart.FileHeader:
		object = &storage.Object{Name: util.RandString(32) + path.Ext(imgFile.Filename)}
		fileToSave, err = imgFile.Open()
		if err != nil {
			return nil, errors.New("failed to open currency image. " + err.Error())
		}

		if osFile, isOSFile := fileToSave.(*os.File); isOSFile {
			defer osFile.Close()
		}

	case *os.File:
		object = &storage.Object{Name: util.RandString(32) + ".jpg"}
		fileToSave = imgFile
		defer imgFile.Close()

	default:
		return nil, errors.New("unsupported file type")
	}

	// add object to bucket
	bucketName := config.C.GetString("bucket_name")
	if obj, err := mc.storageService.Objects.Insert(bucketName, object).Media(fileToSave).Do(); err == nil {
		return obj, nil
	} else {
		return nil, errors.New("failed to create object in cloud storage. " + err.Error())
	}
}

// Analyze a currency. Determine and extract serial and denomination
func (mc *MintController) AnalyzeCurrency(curCode, curDenom, imageName string) (map[string]string, error) {

	// get currency language
	lang := GetCurrencyLang(curCode)

	// process currency image. Get labels and text extracts.
	// imageName = "mumrhVdxIiEMENmGymrMStoYcSgcBXST.jpg"
	gcsImageUri := fmt.Sprintf("gs://%s/%s", config.C.GetString("bucket_name"), imageName)
	imgProcRes, err := ProcessImage(lang, mc.visionService, gcsImageUri)
	if err != nil {
		return nil, err
	}

	// ensure we got a response
	if len(imgProcRes.Responses) == 0 {
		return nil, errors.New("No annotation response found")
	}

	imgProp, _ := imgProcRes.Responses[0].ImagePropertiesAnnotation.MarshalJSON()
	util.Println(string(imgProp))

	// extract tokens from text annotation
	var tokens = AnalyzeText(imgProcRes.Responses[0].TextAnnotations)

	// analyze the tokens extracted from the currency image
	result, err := AnalyzeCurrencyData(curCode, curDenom, tokens, imgProcRes.Responses[0].LabelAnnotations, imgProcRes.Responses[0].ImagePropertiesAnnotation)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Resize image
func (mc *MintController) ResizeImg(currencyImg *multipart.FileHeader, newWidth int) (*os.File, error) {

	file, _ := currencyImg.Open()
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		util.Println(err)
		return nil, errors.New("failed to create image from uploaded file")
	}

	// rotate and resize image
	// img := imaging.Rotate90(img)
	if newWidth > 0 {
		img = imaging.Resize(img, newWidth, 0, imaging.Lanczos)
	}

	// create temp file
	tempFile, err := NewTempFile(os.TempDir(), "openming_img"+util.RandString(32), ".jpg")
	if err != nil {
		util.Println(err)
		return nil, errors.New("failed to create temp file")
	}

	if err = imaging.Save(img, tempFile.Name()); err != nil {
		util.Println(err)
		return nil, errors.New("failed to save new resized/rotated image")
	}

	return tempFile, nil
}

// @API: 			POST /v1/mint/new
// @Description: 	Accepts new currency scans and starts processing
// @Content-Type: 	application/json
// @Body Params: 	currency_image {file}, currency_denom {string}, currency_code {string}
// @Response 200:
// 	id 		{string}: The open mint id of the currency
// 	status 	{string}: The open mint status
// 	name 	{string}: The name of the currency image
// 	link 	{string}: The public link to the currency image
func (mc *MintController) Process(c *extend.Context) error {

	authUserId := c.Get("auth_user")

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

	// resize image for display in applications
	smallerImg, err := mc.ResizeImg(currencyImg, 350)
	if err != nil {
		util.Println(err)
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	defer os.Remove(smallerImg.Name())

	// save image
	smallerImgObj, err := mc.SaveImage(smallerImg)
	if err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// save original currency image
	originalImageObj, err := mc.SaveImage(currencyImg)
	if err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// process image asynchronously
	analysisResult, err := mc.AnalyzeCurrency(curCode, curDenom, originalImageObj.Name)
	if err != nil {
		if err.Error() == "not money" {
			return config.NewHTTPError(c.Lang(), 400, "e015")
		} else {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}
	}

	if analysisResult["serial"] == "" {
		return config.NewHTTPError(c.Lang(), 400, "e016")
	}

	// find matching currency
	_, err = models.Currency.FindCurrency(mc.mongoSession, curCode, analysisResult["denomination"], analysisResult["serial"])
	if err != nil && err != mgo.ErrNotFound {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	} else if err == nil {
		return config.NewHTTPError(c.Lang(), 400, "e017")
	}

	// create currency entry
	currency := &models.CurrencyModel{
		Id:               models.NewId(),
		UserId:           bson.ObjectIdHex(authUserId),
		ImageURL:         fmt.Sprintf("http://storage.googleapis.com/%s/%s", smallerImgObj.Bucket, smallerImgObj.Name),
		OriginalImageURL: fmt.Sprintf("http://storage.googleapis.com/%s/%s", originalImageObj.Bucket, originalImageObj.Name),
		CurrencyCode:     curCode,
		Denomination:     analysisResult["denomination"],
		Serial:           analysisResult["serial"],
		Status:           "awaiting_votes",
	}

	if err = models.Currency.Create(mc.mongoSession, currency); err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	return c.JSON(201, extend.H{
		"id":                 currency.Id.Hex(),
		"image_url":          fmt.Sprintf("http://storage.googleapis.com/%s/%s", smallerImgObj.Bucket, smallerImgObj.Name),
		"original_image_url": fmt.Sprintf("http://storage.googleapis.com/%s/%s", originalImageObj.Bucket, originalImageObj.Name),
		"currency_code":      curCode,
		"cur_denomination":   curDenom,
		"status":             currency.Status,
		"serial":             analysisResult["serial"],
		"denomination":       analysisResult["denomination"],
	})
}

// @API: 				GET /v1/mint/supported_currencies
// @Description: 		Get a map of supported currencies and thier denominations
// @Response 200:
// 	USD {Array[String]}: A list of denominations
func (mc *MintController) GetSupportedCurrencies(c *extend.Context) error {

	var supportedCurrencies = make(map[string][]string)
	for _, curCode := range GetDefinedCurrencies() {
		supportedCurrencies[curCode] = GetCurrencyDenoms(curCode)
	}

	return c.JSON(200, supportedCurrencies)
}
