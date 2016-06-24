package fixtures

var TestCurrencyMeta = map[string]map[string]interface{}{
	"BSD": map[string]interface{}{
    	"lang": "en",
		"denominations": map[string]interface{}{
			"1": map[string]interface{}{
				"join_method": "no",
				"rx": []string{"one", "lynden", "pindling"},
			},
			"5": map[string]interface{}{
				"join_method": "no",
				"rx": []string{"one", "lynden", "pindling"},
			},
		},     
		"text_marks": []string{ "bahamas" },
		"serial": map[string]interface{}{
			"join_method": "space_delimited", 
			"rx": ".*([a-z]{1}[0-9 ]{6,})",
			"rx_group": 1, 
			"filters": []string{"remove-spaces"},
		},	
	},
}