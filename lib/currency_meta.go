package lib

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/ellcrys/util"
)

var currencyMeta = map[string]map[string]interface{}{
	"ALL": map[string]interface{}{},
	"DZD": map[string]interface{}{},
	"ARS": map[string]interface{}{},
	"AUD": map[string]interface{}{},
	"BSD": map[string]interface{}{
		"lang": "en",
		"denominations": map[string]interface{}{
			"1": map[string]interface{}{
				"join_token_method": "no",
				"rx":                []string{"(?i)one", "(?i)sir", "(?i)lynden", "(?i)pindling"},
			},
		},
		"text_marks": []string{"(?i)bahamas", "(?i)central", "(?i)bank"},
		"serial": map[string]interface{}{
			"join_token_method": "no_delimiter",
			"rx":                ".*([A-Z]{1}[0-9]{6}).*",
			"rx_group":          1,
			"filters":           []string{},
		},
	},
	"BHD": map[string]interface{}{},
	"BDT": map[string]interface{}{},
	"AMD": map[string]interface{}{},
	"BBD": map[string]interface{}{},
	"BMD": map[string]interface{}{},
	"BTN": map[string]interface{}{},
	"BOB": map[string]interface{}{},
	"BWP": map[string]interface{}{},
	"BZD": map[string]interface{}{},
	"SBD": map[string]interface{}{},
	"BND": map[string]interface{}{},
	"MMK": map[string]interface{}{},
	"BIF": map[string]interface{}{},
	"KHR": map[string]interface{}{},
	"CAD": map[string]interface{}{},
	"CVE": map[string]interface{}{},
	"KYD": map[string]interface{}{},
	"LKR": map[string]interface{}{},
	"CLP": map[string]interface{}{},
	"CNY": map[string]interface{}{},
	"COP": map[string]interface{}{},
	"KMF": map[string]interface{}{},
	"CRC": map[string]interface{}{},
	"HRK": map[string]interface{}{},
	"CUP": map[string]interface{}{},
	"CZK": map[string]interface{}{},
	"DKK": map[string]interface{}{},
	"DOP": map[string]interface{}{},
	"SVC": map[string]interface{}{},
	"ETB": map[string]interface{}{},
	"ERN": map[string]interface{}{},
	"FKP": map[string]interface{}{},
	"FJD": map[string]interface{}{},
	"DJF": map[string]interface{}{},
	"GMD": map[string]interface{}{},
	"GIP": map[string]interface{}{},
	"GTQ": map[string]interface{}{},
	"GNF": map[string]interface{}{},
	"GYD": map[string]interface{}{},
	"HTG": map[string]interface{}{},
	"HNL": map[string]interface{}{},
	"HKD": map[string]interface{}{},
	"HUF": map[string]interface{}{},
	"ISK": map[string]interface{}{},
	"INR": map[string]interface{}{},
	"IDR": map[string]interface{}{},
	"IRR": map[string]interface{}{},
	"IQD": map[string]interface{}{},
	"ILS": map[string]interface{}{},
	"JMD": map[string]interface{}{},
	"JPY": map[string]interface{}{},
	"KZT": map[string]interface{}{},
	"JOD": map[string]interface{}{},
	"KES": map[string]interface{}{},
	"KPW": map[string]interface{}{},
	"KRW": map[string]interface{}{},
	"KWD": map[string]interface{}{},
	"KGS": map[string]interface{}{},
	"LAK": map[string]interface{}{},
	"LBP": map[string]interface{}{},
	"LSL": map[string]interface{}{},
	"LRD": map[string]interface{}{},
	"LYD": map[string]interface{}{},
	"LTL": map[string]interface{}{},
	"MOP": map[string]interface{}{},
	"MWK": map[string]interface{}{},
	"MYR": map[string]interface{}{},
	"MVR": map[string]interface{}{},
	"MRO": map[string]interface{}{},
	"MUR": map[string]interface{}{},
	"MXN": map[string]interface{}{},
	"MNT": map[string]interface{}{},
	"MDL": map[string]interface{}{},
	"MAD": map[string]interface{}{},
	"OMR": map[string]interface{}{},
	"NAD": map[string]interface{}{},
	"NPR": map[string]interface{}{},
	"ANG": map[string]interface{}{},
	"AWG": map[string]interface{}{},
	"VUV": map[string]interface{}{},
	"NZD": map[string]interface{}{},
	"NIO": map[string]interface{}{},
	"NGN": map[string]interface{}{
		"lang": "en",
		"denominations": map[string]interface{}{
			"5": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"five", "alhaji", "sir", "abubakar", "tafawa", "balewa"}, 2),
				"fuzzy_ignore":      []string{"(?i)[^f]i[a-z]{0,1}e"},
				"join_token_method": "no",
				"rx":                []string{"(?i)five|sir|abubakar|balewa", "(?i)tafawa|alhaji|balewa", "(?i)naira|central"},
			},
			"10": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"ten", "alvan", "ikoku"}, 2),
				"join_token_method": "no_delimiter",
				"rx2":               []string{"(?i)((?<!cen)ten)|alvan|ikoku", "(?i)ikoku|alvan|1900|1971", "(?i)central|bank|currency|naira"},
				"colors": map[string]float64{
					"184,156,135": 10.0,
					"182,172,145": 10.0,
					"248,198,148": 10.0,
					"240,206,142": 10.0,
					"230,176,130": 10.0,
				},
			},
			"20": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"twenty", "general", "murtala", "muhammed"}, 2),
				"fuzzy_ignore":      []string{"(?i)c?e.*tral"},
				"join_token_method": "no_delimiter",
				"rx":                []string{"(?i)twenty|muhammed|murtala|general|1938|1976", "(?i)central|bank|currency|naira|nigeria"},
			},
			"50": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"fifty"}, 2),
				"join_token_method": "no",
				"rx":                []string{"(?i)fifty", "(?i)central|bank|currency|naira"},
			},
			"100": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"hundred", "one", "obafemi", "awolowo", "chief"}, 2),
				"join_token_method": "no_delimiter",
				"rx":                []string{"(?i)hundred|one|1909|1987|chief|obafemi|awolowo", "(?i)chief|obafemi|awolowo", "(?i)currency|naira|central|bank"},
			},
			"200": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"two", "hundred", "sir", "alhaji", "ahmadu", "bello"}, 2),
				"fuzzy_ignore":      []string{"(?i)i[^0-9]{1,}r"},
				"join_token_method": "no_delimiter",
				"rx":                []string{"(?i)hundred|two|sir|alhaji|ahmadu|bello", "(?i)\b200\b|ahmadu|bello|1909|1966", "(?i)nigeria|currency|naira"},
			},
			"500": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"nnamdi", "azikiwe", "hundred", "five"}, 2),
				"join_token_method": "no_delimiter",
				"rx":                []string{"(?i)five|nnamdi|azikiwe|500", "(?i)hundred|nnamdi|azikiwe"},
			},
			"1000": map[string]interface{}{
				"fuzzy":             NewFuzzyModel([]string{"one", "thousand", "naira"}, 2),
				"join_token_method": "no_delimiter",
				"rx":                []string{"(?i)thousand", "(?i)one", "(?i)central|naira|nigeria"},
			},
		},
		"text_marks": []string{},
		"serial": map[string]interface{}{
			"join_token_method": "no_delimiter",
			"rx":                `([A-Z]{2}[0-9]{6,7}|[0-9]{6})`,
			"rx_group":          1,
			"filters":           []string{},
			"rx_5": map[string]interface{}{
				"rx":                `([A-Z]{2}[0-9]{7})`,
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"filters":           []string{},
			},
			"rx_10": "rx_50",
			"rx_20": map[string]interface{}{
				"rx2":               `([A-Z]{2}[0-9]{7})|([A-Z]{2}[0-9]{6})`,
				"rx2_from_right":    true,
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"match_filters":     []string{"remove-empty"},
				"filters":           []string{},
			},
			"rx_50": map[string]interface{}{
				"rx":                `([A-Z]{2}[0-9]{6,7})`,
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"remove_tokens":     []string{"[[:punct:]]"},
				"filters":           []string{},
			},
			"rx_100": map[string]interface{}{
				"rx2":               `(?:(?i:[A-Z])?([A-Z]{2}[0-9]{7})[^0-9])|(?:(?:[A-Z]{2,})?([0-9]{6})(?:[0-9]{2})?)`,
				"remove_tokens":     []string{"1987", "1909", "2014", "1914"}, // remove year
				"rx2_from_right":    true,
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"match_filters":     []string{"remove-empty"},
				"filters":           []string{},
			},
			"rx_200": map[string]interface{}{
				"rx2":               `(?:(?i:[A-Z])?([A-Z]{2}[0-9]{7})[^0-9])|(?:(?:[A-Z]{2,})?([0-9]{6})(?:[0-9]{2})?)`,
				"remove_tokens":     []string{"1909", "1966"},
				"rx2_from_right":    true,
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"match_filters":     []string{"remove-empty"},
				"filters":           []string{},
			},
			"rx_500": map[string]interface{}{
				"rx2":               `(?:(?:[A-Z]{2,})?([0-9]{6})(?i:[a-z]|[0-9]{2})?)`,
				"remove_tokens":     []string{"1904", "1996", "[[:punct:]]"},
				"rx2_from_right":    true,
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"match_filters":     []string{},
				"filters":           []string{},
			},
			"rx_1000": map[string]interface{}{
				"rx":                `(?:(?:[A-Z]{2,}|[0-9]{2,})?([0-9]{6})(?i:[a-z])?)`,
				"remove_tokens":     []string{"[[:punct:]]"},
				"rx_group":          1,
				"join_token_method": "no_delimiter",
				"match_filters":     []string{},
				"filters":           []string{},
			},
		},
	},
	"NOK": map[string]interface{}{},
	"PKR": map[string]interface{}{},
	"PAB": map[string]interface{}{},
	"PGK": map[string]interface{}{},
	"PYG": map[string]interface{}{},
	"PEN": map[string]interface{}{},
	"PHP": map[string]interface{}{},
	"QAR": map[string]interface{}{},
	"RUB": map[string]interface{}{},
	"RWF": map[string]interface{}{},
	"SHP": map[string]interface{}{},
	"STD": map[string]interface{}{},
	"SAR": map[string]interface{}{},
	"SCR": map[string]interface{}{},
	"SLL": map[string]interface{}{},
	"SGD": map[string]interface{}{},
	"VND": map[string]interface{}{},
	"SOS": map[string]interface{}{},
	"ZAR": map[string]interface{}{},
	"SSP": map[string]interface{}{},
	"SZL": map[string]interface{}{},
	"SEK": map[string]interface{}{},
	"CHF": map[string]interface{}{},
	"SYP": map[string]interface{}{},
	"THB": map[string]interface{}{},
	"TOP": map[string]interface{}{},
	"TTD": map[string]interface{}{},
	"AED": map[string]interface{}{},
	"TND": map[string]interface{}{},
	"UGX": map[string]interface{}{},
	"MKD": map[string]interface{}{},
	"EGP": map[string]interface{}{},
	"GBP": map[string]interface{}{},
	"TZS": map[string]interface{}{},
	"USD": map[string]interface{}{},
	"UYU": map[string]interface{}{},
	"UZS": map[string]interface{}{},
	"WST": map[string]interface{}{},
	"YER": map[string]interface{}{},
	"TWD": map[string]interface{}{},
	"CUC": map[string]interface{}{},
	"ZWL": map[string]interface{}{},
	"TMT": map[string]interface{}{},
	"GHS": map[string]interface{}{},
	"VEF": map[string]interface{}{},
	"SDG": map[string]interface{}{},
	"UYI": map[string]interface{}{},
	"RSD": map[string]interface{}{},
	"MZN": map[string]interface{}{},
	"AZN": map[string]interface{}{},
	"RON": map[string]interface{}{},
	"CHE": map[string]interface{}{},
	"CHW": map[string]interface{}{},
	"TRY": map[string]interface{}{},
	"XAF": map[string]interface{}{},
	"XCD": map[string]interface{}{},
	"XOF": map[string]interface{}{},
	"XPF": map[string]interface{}{},
	"XBA": map[string]interface{}{},
	"XBB": map[string]interface{}{},
	"XBC": map[string]interface{}{},
	"XBD": map[string]interface{}{},
	"XAU": map[string]interface{}{},
	"XDR": map[string]interface{}{},
	"XAG": map[string]interface{}{},
	"XPT": map[string]interface{}{},
	"XTS": map[string]interface{}{},
	"XPD": map[string]interface{}{},
	"XUA": map[string]interface{}{},
	"ZMW": map[string]interface{}{},
	"SRD": map[string]interface{}{},
	"MGA": map[string]interface{}{},
	"COU": map[string]interface{}{},
	"AFN": map[string]interface{}{},
	"TJS": map[string]interface{}{},
	"AOA": map[string]interface{}{},
	"BYR": map[string]interface{}{},
	"BGN": map[string]interface{}{},
	"CDF": map[string]interface{}{},
	"BAM": map[string]interface{}{},
	"EUR": map[string]interface{}{},
	"MXV": map[string]interface{}{},
	"UAH": map[string]interface{}{},
	"GEL": map[string]interface{}{},
	"BOV": map[string]interface{}{},
	"PLN": map[string]interface{}{},
	"BRL": map[string]interface{}{},
	"CLF": map[string]interface{}{},
	"XSU": map[string]interface{}{},
	"USN": map[string]interface{}{},
	"XXX": map[string]interface{}{},
}

func SetCurrencyMeta(newMeta map[string]map[string]interface{}) {
	currencyMeta = newMeta
}

// Get a list of all currency code with defined meta data
func GetDefinedCurrencies() []string {
	var definedCurrencies []string
	for cur, meta := range currencyMeta {
		if len(meta) != 0 {
			definedCurrencies = append(definedCurrencies, cur)
		}
	}
	return definedCurrencies
}

// Check if a currency code is valid
func IsValidCode(code string) bool {
	for curName, _ := range currencyMeta {
		if curName == code {
			return true
		}
	}
	return false
}

// Check if currency denomination is
// valid for the currency code supplied.
func IsValidDenomination(curCode string, denom int) bool {

	if !IsValidCode(curCode) {
		panic("Invalid currency code")
	}

	// check denomination
	var curMeta = currencyMeta[curCode]
	for dn, _ := range curMeta["denominations"].(map[string]interface{}) {
		if dn == fmt.Sprintf("%d", denom) {
			return true
		}
	}

	return false
}

// Get the language of a currency
func GetCurrencyLang(curCode string) string {
	if curMeta, set := currencyMeta[curCode]; set {
		return curMeta["lang"].(string)
	}

	return ""
}

// Get the denominations associated with a currency
func GetCurrencyDenoms(curCode string) []string {
	if curMeta, set := currencyMeta[curCode]; set {
		var denoms = util.GetMapKeys(curMeta["denominations"].(map[string]interface{}))
		var ints []int
		for _, str := range denoms {
			i, _ := strconv.Atoi(str)
			ints = append(ints, i)
		}
		sort.Ints(ints)
		for i, anInt := range ints {
			denoms[i] = fmt.Sprintf("%d", anInt)
		}
		return denoms
	}
	return nil
}

// Get the text marks associated with a currency
func GetCurrencyTextMarks(curCode string) []string {
	if curMeta, set := currencyMeta[curCode]; set {
		return curMeta["text_marks"].([]string)
	}
	return nil
}

// Get the seria regex associated with a currency
func GetCurrencySerialData(curCode string) map[string]interface{} {
	if curMeta, set := currencyMeta[curCode]; set {
		return curMeta["serial"].(map[string]interface{})
	}
	return nil
}

// Get denomination data of a currency
func GetDenominationData(curCode string) map[string]interface{} {
	if curMeta, set := currencyMeta[curCode]; set {
		return curMeta["denominations"].(map[string]interface{})
	}
	return nil
}
