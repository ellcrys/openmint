package lib 

import (
	"errors"
	"strings"
	"regexp"
	"fmt"

	vision "google.golang.org/api/vision/v1"
	"github.com/ellcrys/util"
	"github.com/ellcrys/openmint/config"
)

// Given an image, it will use google vision to detect content
// and extra text
func ProcessImage(lang string, service *vision.Service, imageUri string) (*vision.BatchAnnotateImagesResponse, error) {
	
	var res *vision.BatchAnnotateImagesResponse

	// create image 
	img := vision.Image {
		Source: &vision.ImageSource {
			GcsImageUri: imageUri,
		},
	}
	
	// create annotate requeest
	req := &vision.AnnotateImageRequest{
		
		Image: &img,

		ImageContext: &vision.ImageContext {
			LanguageHints: strings.Split(lang, ","),
		},

		// Apply features to indicate what type of image detection.
		Features: []*vision.Feature{
			{
				MaxResults: 50,
				Type: "TEXT_DETECTION",
			},
			{
				MaxResults: 10,
				Type: "LABEL_DETECTION",
			},
		},
	}

	batch := &vision.BatchAnnotateImagesRequest{
		Requests: []*vision.AnnotateImageRequest{req},
	}

	res, err := service.Images.Annotate(batch).Do()
	if err != nil {
		return res, errors.New("Unable to execute images annotate requests: " + err.Error())
	}

	return res, nil
}

// Given a currency code, check if the text marks associated with the
// currency are all found the currency tokens passed. If no text mark is 
// provided, return true
func HasTextMarks(curCode string, curTokens []string) bool {
	var textMarks = config.GetCurrencyTextMarks(curCode)
	if len(textMarks) == 0 {
		return true
	}
	for _, tm := range textMarks {
		if !util.InStringSlice(curTokens, tm) {
			return false
		}
	}
	return true
}

// Given a slice of tokens, it will join the tokens
// based on the method provided. 
// "space_delimited" will join all tokens with a single space between them.
// "no_delimiter" will join all tokens with no space between them.
func JoinToken(tokens []string, method string) string {
	switch method {
	case "space_delimited":
		return strings.Join(tokens, " ")
	case "no_delimiter":
		return strings.Join(tokens, "")
	default:
		util.Println("Warning: ", "Unsupported join method: ", method)
		return ""
	}
}

// Given a currency code, pick the token that represents
// the serial of the currency
func ExtractSerial(denomination, curCode string, curTokens []string) (string, error) {

	var serial = ""
	var data = config.GetCurrencySerialData(curCode)
	var joinMethod = "no"
	var joinedTokens = ""
	var pattern string
	var group int 
	var filters []string

	// if no denomination, use default regex pattern, group and join method
	if denomination == "" {
		
		pattern = data["rx"].(string)
		group 	= data["rx_group"].(int)
		filters = data["filters"].([]string)
		
		if data["join_method"].(string) != "" {
			joinMethod = data["join_method"].(string)
		}
		
	} else {

		// ensure a specific regex instruction exists for the
		// denomination, otherwise use default
		denomRxInst, set := data[fmt.Sprintf("rx_%s", denomination)]
		if !set {

			pattern = data["rx"].(string)
			group 	= data["rx_group"].(int)
			filters = data["filters"].([]string)

			if data["join_method"].(string) != "" {
				joinMethod = data["join_method"].(string)
			}
			
		} else {

			// string type means a reference to another
			// regex directive. Only support one-level referencing
			switch rxData := denomRxInst.(type) {

				case map[string]interface{}:

					pattern = rxData["rx"].(string)
					group   = rxData["group"].(int)
					filters = rxData["filters"].([]string)

					if rxData["join_method"].(string) != "" {
						joinMethod = rxData["join_method"].(string)
					}

				case string:
					
					var instructionName = rxData

					// referenced regex instruction must be defined
					if instructionData, set := data[instructionName]; set {

						if instruction, ok := instructionData.(map[string]interface{}); ok {
							pattern = instruction["rx"].(string)
							group   = instruction["group"].(int)
							filters = instruction["filters"].([]string)
							if instruction["join_method"].(string) != "" {
								joinMethod = instruction["join_method"].(string)
							}

						} else {
							return serial, errors.New("unsupported reference regex directive type")
						}

					} else {
						return serial, errors.New(fmt.Sprintf("regex instruction '%s' not defined", rxData))
					}

				default:
					return serial, errors.New("unsupported regex directive type")
			}
		}
	}
	
	// join tokens
	if joinMethod != "no" {
		joinedTokens = JoinToken(curTokens, joinMethod)
	}

	// find serial in the tokens
	if joinMethod == "no" {

		for _, token := range curTokens {
			match, _ := regexp.MatchString(pattern, token)
			if match {
				serial = token
				break;
			}
		}

		// pass through filter functions if required
		if data["filters"] != nil {
			serial = Filter(serial, filters)
		}
	}

	// search for serial pattern in joined tokens
	if joinedTokens != "" {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(joinedTokens); match != nil {
			serial = match[group]

			// pass through filter functions if required
			if data["filters"] != nil {
				serial = Filter(serial, filters)
			}

			return serial, nil
		}
	}

	return serial, nil
}


// Given a currency code and a slice of tokens, determine
// the denomination of the currency represented by the currency code.
func DetermineDenomination(curCode string, curTokens []string) string {

	var result string

	// denomination data
	denominationData := config.GetDenominationData(curCode)
	for denom, data := range denominationData {

		// join tokens?
		var joinedTokens = ""
		var joinMethod = data.(map[string]interface{})["join_method"].(string)
		if joinMethod != "" {
			joinedTokens = JoinToken(curTokens, joinMethod)
		}

		// search for denomination patterns in joined tokens
		if joinedTokens != "" {
			var matchCount = 0
			var patterns = data.(map[string]interface{})["rx"].([]string)
			for _, pattern := range patterns {
				match, _ := regexp.MatchString(pattern, joinedTokens)
				if match {
					matchCount++
				}
				if matchCount == len(patterns) {
					result = denom
					break
				}
			}
		}

		// search for denonimationn patterns in token slice
		if joinMethod == "no" {
			var matchCount = 0
			var patterns = data.(map[string]interface{})["rx"].([]string)
			for _, pattern := range patterns {
				for _, token := range curTokens {
					match, _ := regexp.MatchString(pattern, token)
					if match {
						matchCount++
						break
					}
				}
				if matchCount == len(patterns) {
					result = denom
					break
				}
			}
		}
	}

	return result
}

// Given a collection of labels that describe a currency image
// and a collection tokens extracted from the image, it tries to 
// determine if the image is a valid currency while taking the
// expected denomination and other textual landmarks.
func ProcessMoney(curCode string, curTokens []string, labels []*vision.EntityAnnotation) (map[string]string, error) {

	var minScore = 0.5
	var labelsFound = []string{}
	var result = make(map[string]string)

	for _, label := range labels {
		var expectedDescription = []string{"currency", "money", "cash"}
		if util.InStringSlice(expectedDescription, label.Description) {
			if label.Score >= minScore {
				labelsFound = append(labelsFound, label.Description)
			}
		}
	}

	// currency & money labels must be found
	if !util.InStringSlice(labelsFound, "currency") && !util.InStringSlice(labelsFound, "money") {
		return result, errors.New("not money")
	}

	// currency token must contain text marks
	if !HasTextMarks(curCode, curTokens) {
		return result, errors.New("text mark check failed")
	}

	// determine denomination
	suggestedDenomination := DetermineDenomination(curCode, curTokens)
	if suggestedDenomination != "" {
		result["denomination"] = suggestedDenomination
	}

	// extract serial number
	serial, err := ExtractSerial(suggestedDenomination, curCode, curTokens)
	if result["serial"] == "" || err != nil {
		return result, errors.New("failed to extract serial. " + err.Error())
	}

	result["serial"] = serial
	return result, nil
}

// Given text extracted from an currency image, clean it up
// and generate tokens
func ProcessText(texts []*vision.EntityAnnotation) []string {

	var tokens = []string{}
	for _, text := range texts {

		// clean up text, create tokens
		var t = text.Description
		t = strings.Replace(t, "\n", " ", -1)
		for _, tok := range strings.Split(t, " ") {
			tok = strings.TrimSpace(tok)
			if len(tok) != 0 {
				tokens = append(tokens, tok)
			}
		}
	}

	return tokens
}