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
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/disintegration/imaging"
	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/util"
	"github.com/garyburd/redigo/redis"
	storage "google.golang.org/api/storage/v1"
	vision "google.golang.org/api/vision/v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const NotMultipart = "request Content-Type isn't multipart/form-data"

type addVoteBody struct {
	CurrencyId string `json:"currency_id" valid:"required"`
	Decision   int    `json:"decision"`
	VoteId     string `json:"vote_id" valid:"required"`
}

type MintController struct {
	mongoSession   *mgo.Session
	redisPool      *redis.Pool
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
func NewMintController(mongoSession *mgo.Session, redisPool *redis.Pool, storageClient *http.Client, visionClient *http.Client) *MintController {
	storageService := createStorageService(storageClient)
	visionService := createVisionService(visionClient)
	return &MintController{mongoSession, redisPool, storageService, visionService}
}

// Store image in google cloud storage.
// File can be a multipart.File or an os.File.
// File handle will be closed after save is complete.
func (self *MintController) SaveImage(file interface{}) (*storage.Object, error) {

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
	if obj, err := self.storageService.Objects.Insert(bucketName, object).Media(fileToSave).Do(); err == nil {
		return obj, nil
	} else {
		return nil, errors.New("failed to create object in cloud storage. " + err.Error())
	}
}

// Delete an image from the primary mint bucket on google storage.
func (self *MintController) DeleteImage(objName string) error {
	bucketName := config.C.GetString("bucket_name")
	if err := self.storageService.Objects.Delete(bucketName, objName).Do(); err == nil {
		return nil
	} else {
		return errors.New("failed to delete object in cloud storage. " + err.Error())
	}
}

// Analyze a currency. Determine and extract serial and denomination
func (self *MintController) AnalyzeCurrency(curCode, curDenom, imageName string) (map[string]string, error) {

	// get currency language
	lang := GetCurrencyLang(curCode)

	// process currency image. Get labels and text extracts.
	// imageName = "mumrhVdxIiEMENmGymrMStoYcSgcBXST.jpg"
	startTime := time.Now().Unix()
	gcsImageUri := fmt.Sprintf("gs://%s/%s", config.C.GetString("bucket_name"), imageName)
	imgProcRes, err := ProcessImage(lang, self.visionService, gcsImageUri)
	if err != nil {
		return nil, err
	}

	util.Println("Vision Processing Took: ", time.Now().Unix()-startTime)

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
func (self *MintController) ResizeImg(currencyImg *multipart.FileHeader, newWidth int) (*os.File, error) {

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

// @API: 				POST /v1/mint/new
// @Description: 		Accepts new currency scans and starts processing
// @Content-Type: 		application/json
//
// @Body (Multipart):
// 	currency_image 		File: 		The image of the currency
// 	currency_denom 		String:		The expected denomination on currency
// 	currency_code 		String:		The currency code (NGN, USD etc)
//
// @Response 200:
// 	id 		string: The open mint id of the currency
// 	status 	string: The open mint status
// 	name 	string: The name of the currency image
// 	link 	string: The public link to the currency image
func (self *MintController) Process(c *extend.Context) error {

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
	smallerImg, err := self.ResizeImg(currencyImg, 350)
	if err != nil {
		util.Println(err)
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	defer os.Remove(smallerImg.Name())

	startTime := time.Now().Unix()

	// save image
	smallerImgObj, err := self.SaveImage(smallerImg)
	if err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// save original currency image
	originalImageObj, err := self.SaveImage(currencyImg)
	if err != nil {
		go self.DeleteImage(smallerImgObj.Name)
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	util.Println("Image Uploaded In: ", time.Now().Unix()-startTime)

	startTime = time.Now().Unix()

	// process image asynchronously
	analysisResult, err := self.AnalyzeCurrency(curCode, curDenom, originalImageObj.Name)
	if err != nil {

		go self.DeleteImage(smallerImgObj.Name)
		go self.DeleteImage(originalImageObj.Name)

		if err.Error() == "not money" {
			return config.NewHTTPError(c.Lang(), 400, "e015")
		} else {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}
	}

	util.Println("Image Processed In: ", time.Now().Unix()-startTime)

	if analysisResult["serial"] == "" {
		go self.DeleteImage(smallerImgObj.Name)
		go self.DeleteImage(originalImageObj.Name)
		return config.NewHTTPError(c.Lang(), 400, "e016")
	}

	// find matching currency
	_, err = models.Currency.FindCurrency(self.mongoSession, curCode, analysisResult["denomination"], analysisResult["serial"])
	if err != nil && err != mgo.ErrNotFound {
		go self.DeleteImage(smallerImgObj.Name)
		go self.DeleteImage(originalImageObj.Name)
		return config.NewHTTPError(c.Lang(), 500, "e500")
	} else if err == nil {
		go self.DeleteImage(smallerImgObj.Name)
		go self.DeleteImage(originalImageObj.Name)
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

	if err = models.Currency.Create(self.mongoSession, currency); err != nil {
		go self.DeleteImage(smallerImgObj.Name)
		go self.DeleteImage(originalImageObj.Name)
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	// add currency to the vote queue
	if err = models.AddToVoteQueue(self.redisPool, currency.Id.Hex()); err != nil {
		go self.DeleteImage(smallerImgObj.Name)
		go self.DeleteImage(originalImageObj.Name)
		go models.Currency.Delete(self.mongoSession, currency.Id.Hex())
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
func (self *MintController) GetSupportedCurrencies(c *extend.Context) error {

	var supportedCurrencies = make(map[string][]string)
	for _, curCode := range GetDefinedCurrencies() {
		supportedCurrencies[curCode] = GetCurrencyDenoms(curCode)
	}

	return c.JSON(200, supportedCurrencies)
}

// @API: 		 	GET /v1/mint/vote
//
// @Description: 	Request a vote session. This endpoint will return a currency to
// 	to vote on and a vote session id. The vote session id allows vote responses to be accepted by `AddVote()`
//
// @Response 200:
// 	currency 	Object: 	The currency to vote on
// 	vote_id 	String:		The vote id
func (self *MintController) GetVoteSession(c *extend.Context) error {

	var maxRepeat = 3
	var countRepeat = 0

	for {

		// exist loop if maxRepeat threshold is reached
		if maxRepeat == countRepeat {
			break
		}

		countRepeat++

		// fetch a currency to vote on
		currencyId, err := models.GetFromVoteQueue(self.redisPool)

		// no currency found
		if err != nil && err == redis.ErrNil {
			continue
		} else if err != nil {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}

		currency, err := models.Currency.FindById(self.mongoSession, currencyId)
		if err != nil {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}

		// if current votes is equal to max vote, remove currency from queue
		if len(currency.Votes) == config.C.GetInt("max_votes") {
			if err = models.RemoveFromVoteQueue(self.redisPool, currencyId); err != nil {
				return config.NewHTTPError(c.Lang(), 500, "e500")
			}
			continue
		}

		// count the number of active voting sessions for this currency.
		// if number of active session is greater or equal to the number of votes remaining,
		// continue to next iteration
		numActiveSessions, err := models.CountActiveSessions(self.redisPool, currencyId)
		if err != nil {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}

		if numActiveSessions >= (config.C.GetInt("max_votes") - len(currency.Votes)) {
			util.Println("Max session reached")
			continue
		}

		// add new session
		voteSessionId := util.Sha1(util.RandString(32))
		err = models.AddNewSession(self.redisPool, currencyId, voteSessionId)
		if err != nil {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}

		return c.JSON(200, extend.H{
			"currency": currency,
			"vote_id":  voteSessionId,
		})
	}

	return config.NewHTTPError(c.Lang(), 400, "e018")
}

// @API: 		 	PUT /v1/mint/vote
// @Description: 	Add a vote to a currency
//
// @Body (JSON):
// 	currency_id 	String: The currency id
// 	decision		Int: The vote decision. Binary (0 or 1).
// 	vote_id 		String: The vote id received from `GetVoteSession()`
//
// @Response 200:
func (self *MintController) AddVote(c *extend.Context) error {

	authUserId := c.Get("auth_user")

	var body addVoteBody
	if c.BindJSON(&body) != nil {
		return config.NewHTTPError(c.Lang(), 400, "e001")
	}

	// validate request body
	if _, err := govalidator.ValidateStruct(body); err != nil {
		return config.ValidationError(c, err)
	}

	currency, err := models.Currency.FindById(self.mongoSession, body.CurrencyId)
	if err != nil && err == mgo.ErrNotFound {
		return config.NewHTTPError(c.Lang(), 404, "e020")
	}

	// ensure maximum number of vote hasn't been reached or (passed, if ever)
	if len(currency.Votes) >= 3 {
		return config.NewHTTPError(c.Lang(), 400, "e021")
	}

	// ensure authenticated user hasn't added a vote to this currency
	for _, vote := range currency.Votes {
		if vote.UserId.Hex() == authUserId {
			return config.NewHTTPError(c.Lang(), 400, "e023")
		}
	}

	// check if vote id is valid and is still active
	active, err := models.IsActiveSession(self.redisPool, body.CurrencyId, body.VoteId)
	if err != nil {
		return config.NewHTTPError(c.Lang(), 400, "e022")
	}

	if !active {
		return config.NewHTTPError(c.Lang(), 400, "e019")
	}

	// append votes
	currency.Votes = append(currency.Votes, models.Vote{
		Decision:  body.Decision,
		UserId:    bson.ObjectIdHex(authUserId),
		CreatedAt: time.Now().UTC(),
	})

	// update votes
	if err = models.Currency.UpdateVotes(self.mongoSession, currency.Id.Hex(), currency.Votes); err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	return c.JSON(200, currency)
}
