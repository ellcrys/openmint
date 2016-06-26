package lib 

import (
	"errors"
	"strings"
	"regexp"
	"fmt"

	vision "google.golang.org/api/vision/v1"
	"github.com/ellcrys/util"
	"github.com/dlclark/regexp2"
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
	
	var textMarks = GetCurrencyTextMarks(curCode)
	if len(textMarks) == 0 {
		return true
	}

	var found = 0
	for _, tm := range textMarks {
		for _, token := range curTokens {
			if match, _ := regexp.MatchString(tm, token); match {
				found++
				break
			}
		}
	}

	return found == len(textMarks)
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
	var data = GetCurrencySerialData(curCode)
	var joinMethod = "no"
	var joinedTokens = ""
	var pattern string
	var group int 
	var filters []string
	var matchFilters []string
	var backtrackEnabled = false
	var rightToLeft = false
	var tokensToRemove []string
	
	// if no denomination, use default regex pattern, group and join method
	if denomination == "" {
		
		// enable backtrack regex engine if `rx2` is set
		if rx, set := data["rx2"]; set {
			pattern = rx.(string)
			backtrackEnabled = true

			if data["rx2_from_right"] != nil && data["rx2_from_right"].(bool) {
				rightToLeft = true
			}

		} else {
			pattern = data["rx"].(string)
			backtrackEnabled = false
		}

		group 	= data["rx_group"].(int)
		filters = data["filters"].([]string)
		
		if data["join_token_method"].(string) != "" {
			joinMethod = data["join_token_method"].(string)
		}

		if _, set := data["match_filters"]; set {
			matchFilters = data["match_filters"].([]string)
		}

		if data["remove_tokens"] != nil {
			tokensToRemove = data["remove_tokens"].([]string)
		}
		
	} else {

		// ensure a specific regex instruction exists for the
		// denomination, otherwise use default
		denomRxInst, set := data[fmt.Sprintf("rx_%s", denomination)]
		if !set {
			// enable backtrack regex engine if `rx2` is set
			if rx2, set := data["rx2"]; set {
				pattern = rx2.(string)
				backtrackEnabled = true

				if data["rx2_from_right"] != nil && data["rx2_from_right"].(bool) {
					rightToLeft = true
				}

			} else {
				pattern = data["rx"].(string)
				backtrackEnabled = false
			}

			group 	= data["rx_group"].(int)
			filters = data["filters"].([]string)

			if data["join_token_method"].(string) != "" {
				joinMethod = data["join_token_method"].(string)
			}
			
			if _, set := data["match_filters"]; set {
				matchFilters = data["match_filters"].([]string)
			}

			if data["remove_tokens"] != nil {
				tokensToRemove = data["remove_tokens"].([]string)
			}

		} else {

			// string type means a reference to another
			// regex directive. Only support one-level referencing
			switch rxData := denomRxInst.(type) {

				case map[string]interface{}:

					if rx2, set := rxData["rx2"]; set {

						pattern = rx2.(string)
						backtrackEnabled = true

						if rxData["rx2_from_right"] != nil && rxData["rx2_from_right"].(bool) {
							rightToLeft = true
						}

					} else {
						pattern = rxData["rx"].(string)
						backtrackEnabled = false
					}

					group   = rxData["rx_group"].(int)
					filters = rxData["filters"].([]string)

					if rxData["join_token_method"].(string) != "" {
						joinMethod = rxData["join_token_method"].(string)
					}

					if _, set := rxData["match_filters"]; set {
						matchFilters = rxData["match_filters"].([]string)
					}

					if rxData["remove_tokens"] != nil {
						tokensToRemove = rxData["remove_tokens"].([]string)
					}

				case string:
					
					var instructionName = rxData

					// referenced regex instruction must be defined
					if instructionData, set := data[instructionName]; set {

						if instruction, ok := instructionData.(map[string]interface{}); ok {
							
							// enable backtrack regex engine if `rx2` is set
							if rx2, set := instruction["rx2"]; set {

								pattern = rx2.(string)
								backtrackEnabled = true

								if instruction["rx2_from_right"] != nil && instruction["rx2_from_right"].(bool) {
									rightToLeft = true
								}

							} else {
								pattern = instruction["rx"].(string)
								backtrackEnabled = false
							}

							group   = instruction["rx_group"].(int)
							filters = instruction["filters"].([]string)

							if instruction["join_token_method"].(string) != "" {
								joinMethod = instruction["join_token_method"].(string)
							}

							if _, set := instruction["match_filters"]; set {
								matchFilters = instruction["match_filters"].([]string)
							}

							if instruction["remove_tokens"] != nil {
								tokensToRemove = instruction["remove_tokens"].([]string)
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

			// if token match any of the token pattern in the 'tokens to remove' slice,
			// ignore it
			if util.StringSliceMatchString(tokensToRemove, token) != "" {
				continue
			}

			if !backtrackEnabled {
				if match, _ := regexp.MatchString(pattern, token); match {
					serial = token
					break;
				}

			} else {

				var reOpt regexp2.RegexOptions
				if rightToLeft {
					reOpt = reOpt | regexp2.RightToLeft
				}

				re := regexp2.MustCompile(pattern, reOpt)

				if rightToLeft {
					re.RightToLeft()
				}

				match, _ := re.MatchString(token)
				if match {
					serial = token
					break;
				}
			}
		}

		// pass through filter functions if required
		serial = Filter(serial, filters)
		return serial, nil
	}

	// search for serial pattern in joined tokens
	if joinedTokens != "" {
		
		var match []string

		for _, tokenPattern := range tokensToRemove {
			reg := regexp.MustCompile(tokenPattern)
			joinedTokens = reg.ReplaceAllString(joinedTokens, "")
		}

		util.Println(joinedTokens)
		util.Println("Backtrack Enabled: ", backtrackEnabled)

		if !backtrackEnabled {

			re := regexp.MustCompile(pattern)
			if match = re.FindStringSubmatch(joinedTokens); match != nil {
				
				// pass through match filters if required
				match = MatchFilter(match, matchFilters)
				serial = match[group]

				// pass through filter functions if required
				serial = Filter(serial, filters)

				return serial, nil
			}

		} else {

			var reOpt regexp2.RegexOptions
			if rightToLeft {
				reOpt = reOpt | regexp2.RightToLeft
			}

			re := regexp2.MustCompile(pattern, reOpt)

			if m, _ := re.FindStringMatch(joinedTokens); m != nil {

				for i := 0; i < m.GroupCount(); i++ {
					grp := m.GroupByNumber(i);
					match = append(match, grp.String())
				}

				// pass through match filters if required
				match = MatchFilter(match, matchFilters)
				serial = match[group]

				// pass through filter functions if required
				if data["filters"] != nil {
					serial = Filter(serial, filters)
				}

				return serial, nil
			}
		}
	}

	return serial, nil
}

// Given a slice of tokens, it will use a trained fuzzy model
// to suggest alternate words for every token and includes
// suggested tokens in the slice of tokens
func FuzzySuggestTokens(fuzzyModel *FuzzyModel, tokens []string, tokensToIgnore []string) []string {

	var newTokens []string
	
	for _, token := range tokens {

		// if token matches a pattern in the fuzzy ignore list, ignore token
		if util.StringSliceMatchString(tokensToIgnore, token) != "" {
			util.Println("Ignore (Fuzzy): ", token)
			continue
		}

		suggestedTokens := fuzzyModel.GetSuggestions(token)
		newTokens = append(newTokens, token)
		if len(suggestedTokens) > 0 {
			for _, suggestedToken := range suggestedTokens {
				newTokens = append(newTokens, suggestedToken)
			}
		}

		util.Println("Word: ", token, " Suggestion: ", suggestedTokens)
	}
	return newTokens
}

// Given a currency code and a slice of tokens, determine
// the denomination of the currency represented by the currency code.
func DetermineDenomination(curCode string, curTokens []string) string {

	var result string

	for _, denom := range GetCurrencyDenoms(curCode) {

		var tokens = curTokens

		// denomination data
		denominationData := GetDenominationData(curCode)
		data := denominationData[denom]		

		// fuzzy token suggestion
		if data.(map[string]interface{})["fuzzy"] != nil {

			// get slice of tokens to skip in fuzzy suggestion
			ignoreTokens := []string{}
			if data.(map[string]interface{})["fuzzy_ignore"] != nil {
				ignoreTokens = data.(map[string]interface{})["fuzzy_ignore"].([]string)
			}

			tokens = FuzzySuggestTokens(data.(map[string]interface{})["fuzzy"].(*FuzzyModel), tokens, ignoreTokens)
		}

		util.Println("Checking Denom: ", denom)
		for _, t := range tokens {
			util.Println("[",t,"]")
		}
		util.Println("*************************************")

		// join tokens?
		var joinedTokens = ""
		var joinMethod = data.(map[string]interface{})["join_token_method"].(string)
		if joinMethod != "" && joinMethod != "no" {
			joinedTokens = JoinToken(tokens, joinMethod)
		}

		// search for denomination patterns in joined tokens
		if joinedTokens != "" {

			var matchCount = 0
			var backtrackEnabled = false;
			var patterns []string

			// Use backtrack enabled regex engine if `rx2` is used.
			// Otherwise, use native go regex with no backtrack support.
			if (data.(map[string]interface{})["rx"] != nil) {
				patterns = data.(map[string]interface{})["rx"].([]string)
			} else {
				patterns = data.(map[string]interface{})["rx2"].([]string)
				backtrackEnabled = true
			}

			for _, pattern := range patterns {

				if !backtrackEnabled {

					match, _ := regexp.MatchString(pattern, joinedTokens)
					if match {
						matchCount++
					}

				} else {

					re := regexp2.MustCompile(pattern, 0)
					match, _ := re.MatchString(joinedTokens)
					if match {
						matchCount++
					}
				}
           		
				if matchCount == len(patterns) {
					return denom
				}
			}
		}

		// search for denonimationn patterns in token slice
		if joinMethod == "no" {
			
			var matchCount = 0
			var backtrackEnabled = false;
			var patterns []string

			// Use backtrack enabled regex engine if `rx2` is used.
			// Otherwise, use native go regex with no backtrack support.
			if (data.(map[string]interface{})["rx"] != nil) {
				patterns = data.(map[string]interface{})["rx"].([]string)

			} else {
				patterns = data.(map[string]interface{})["rx2"].([]string)
				backtrackEnabled = true
			}

			// find a token match for every pattern
			for _, pattern := range patterns {

				for _, token := range tokens {

					if !backtrackEnabled {

						match, _ := regexp.MatchString(pattern, token)
						if match {
							matchCount++
							break
						}

					} else {

						re := regexp2.MustCompile(pattern, 0)
						match, _ := re.MatchString(token)
						if match {
							matchCount++
							break
						}
					}
				}

				if matchCount == len(patterns) {
					return denom
				}
			}
		}
	}

	return result
}

// Given a collection of labels that describe a currency image
// and a collection tokens extracted from the image, it tries to 
// determine if the image is a valid currency while taking an
// optional denomination and other textual landmarks.
// 
// If denomination is provided, the function will not attempt to
// detect denomination. It will simply match againts the specified 
// denomination data. 
func ProcessMoney(curCode, curDenom string, curTokens []string, labels []*vision.EntityAnnotation) (map[string]string, error) {

	var minScore = 0.5
	var labelsFound = []string{}
	var result = make(map[string]string)

	for _, label := range labels {
		var expectedDescription = []string{"currency", "money"}
		if util.InStringSlice(expectedDescription, label.Description) {
			if label.Score >= minScore {
				labelsFound = append(labelsFound, label.Description)
			}
		}
	}

	// either currency or money label must be found
	if !util.InStringSlice(labelsFound, "currency") && !util.InStringSlice(labelsFound, "money") {
		return result, errors.New("not money")
	}

	// currency token must contain text marks
	if !HasTextMarks(curCode, curTokens) {
		return result, errors.New("text mark check failed")
	}	

	if curDenom == "" {

		// determine denomination
		curDenom = DetermineDenomination(curCode, curTokens)
		if curDenom != "" {
			result["denomination"] = curDenom
		}

	} else {
		result["denomination"] = curDenom
	}

	// extract serial number
	serial, err := ExtractSerial(curDenom, curCode, curTokens)
	if err != nil {
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