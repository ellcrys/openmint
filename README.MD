## Openmint

Openmint server application accepts request containing information such as a currency image, currency code and optionally,
the denomination. It analyzes it image, extracts the serial and denomination (if not provided). The `lib/currency_meta.go`
file is the place where currency is described. If a currency is not included in this file, 
it is therefore not recognized. 

Openmint uses Google Vision API to determine if an image is a currency or money and also 
to retrieve text inscribed on the image. It then tries to suggests the denomination
based on the data provided for a currency. I may return false positives denomination suggestions
when the uniqueness of denomination is low between two or more different denomination of a 
same currency. For better denomination suggestion and subsequently more accurate serial 
extraction, the request can contain the denomination of the currency. 

Below is an example request using Nodejs request package:

```js
var fs = require("fs");
var request = require("request");

var options = { method: 'POST',
  url: 'http://example.com/v1/mint/new',
  headers: 
   { 'postman-token': 'b506f89f-4ac1-7b45-34d0-7132e6dd3963',
     'cache-control': 'no-cache',
     'content-type': 'multipart/form-data; boundary=---011000010111000001101001' },
  formData: 
   { currency_image: 
      { value: 'fs.createReadStream("[object Object]")',
        options: { filename: { '0': {} }, contentType: null } },
     currency_code: 'USD' } };

request(options, function (error, response, body) {
  if (error) throw new Error(error);

  console.log(body);
});

```

### Curreny Meta

Every supported currency is described in `lib/currency_meta.go`. The `currencyMeta`
map holds instructions on how to parse and extract the data we need.
For this doc, lets assume `$currency` represents a currency. We can now describe all the available 
instructions as follows:

# $currency.lang
This is the language the currency text is written in. e.g en, fr, ru. This
helps to greatly improve the text extraction operation. I can also
produce really bad results if the wrong language is used.

# $currency.denominations
This is a map describing all the denominations of the currency. It
takes the currency codes as the key and a map of instruction as value.

Example:

```go
map[string]interface{}{
	"10": map[string]interface{}{
		"fuzzy": NewFuzzyModel([]string{ "ten", "alvan", "ikoku" }, 2),
		"join_token_method": "no_delimiter",
		"rx": []string{ "ten","ikoku", "something" },
	},
}
```

##### $currency.denominations.{DENOM}.join_token_method
This instruction tells the engine how to handle tokens. Passing `no` will
leave the tokens in their inital state as a list of tokens extracted from the curreny. `space_delimited` 
will join the tokens into a single string with space between each token. `no_delimiter` will join the tokens together
with no delimeter between them. These options give you the freedom to search for specific terms in a one
whole text.

##### $currency.denominations.{DENOM}.rx
A list of regular expression pattern for picking out patterns that uniquely
identify the denomination. Before a denomination is selected as a match, 
all patterns in the slice must match. The regex engine used is the default
go engine. To use an alternative engine that is based on .NET regexp, use
the `$currency.denominations.{DENOM}.rx2` option instead.

##### $currency.denominations.{DENOM}.rx2 - (optional)
Similar to `$currency.denominations.{DENOM}.rx`. It supports
backtracking and as such supports lookahead operations. It uses [regexp2](https://github.com/dlclark/regexp2). 

##### $currency.denominations.{DENOM}.fuzzy - (optional):
Pass an instance to a `NewFuzzyModel` object. The `NewFuzzyModel` provides fuzzy 
searching capabilities. This will allow us fix tokens with erroneous or incomplete spelling.

##### $currency.denominations.{DENOM}.fuzzy_ignore - (optional):
Provide a slice of tokens to be ignored when running the extracted
tokens through the fuzzy model. This is useful for when you need to
prevent certain words from returning unwanted fuzzy matches. 

##### $currency.denominations.text_marks - (optional): 
This should optionally contain a slice of tokens that must be
present in every denomination of a single currency.

##### $currency.denominations.{DENOM}.serial 
This is a key/value map of instructions that helps in the
extraction of a currency's serial number. It can contain instructions 
for every denominations. By convention, the instruction for a 
denomination is named in the format `rx_{YOUR_DENOMINATION}`. For example, the 
instruction for a 100 USD bill will be named `rx_100`. 

A denomination can also use the instructions of another
denominations by simply referencing the name of the instruction it wants to use. 
If a denomination does not have have a specific instruction, the default 
is used. Any instruction option that is not within a named denomination instructions
key/value map is considered a default instruction option. 

Here are the specific options in the instructions map:

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.join_token_method
This directive tells the engine how it should handle tokens. Passing `no` will
leave the tokens in their inital state as a list of tokens extracted from the curreny. `space_delimited` 
will join the tokens into a single string with space between each token. `no_delimiter` will join the tokens together
with no delimeter between them. These options give you the freedom to search for specific terms in one
whole text.

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.rx 
A regular expression pattern for selecting the serial number from 
the tokens extracted from the currency. Uses golang's native regex engine.

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.rx2
Similar to `$currency.denominations.{DENOM}.serial.{rx_DENOM}.rx`. It supports
backtracking and as such supports lookahead operations. It uses [regexp2](https://github.com/dlclark/regexp2). 

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.rx2_from_right
Sets the `RIGHT_TO_LEFT` option of the regex engine. This will only 
work if `$currency.denominations.{DENOM}.serial.{rx_DENOM}.rx2` is being used. 

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.rx_group
This is the group index to use as the resulting serial when a match is found.

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.filters
This is a list of names that represents filter to pass the resulting
serial through. Useful for performing operations on an extracted serial.

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.match_filters
This is a list of names that represents filter to pass the regex
match slice through. Useful for performing operations result of the `rx` or `rx2`
patterns.

##### $currency.denominations.{DENOM}.serial.{rx_DENOM}.remove_tokens
This should hold a slice of strings to remove from the list of tokens
is `join_token_method` is `no` or the text if `join_token_method` is either
`space_delimited` or `no_delimiter`. Regex patterns can also be included. 
Only native regex engine is currently supported. 

