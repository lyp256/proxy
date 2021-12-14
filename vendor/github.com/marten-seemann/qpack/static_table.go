package qpack

var staticTableEntries = [...]HeaderField{
	{Name: ":authority"},
	{Name: ":path", Value: "/"},
	{Name: "age", Value: "0"},
	{Name: "content-disposition"},
	{Name: "content-length", Value: "0"},
	{Name: "cookie"},
	{Name: "date"},
	{Name: "etag"},
	{Name: "if-modified-since"},
	{Name: "if-none-match"},
	{Name: "last-modified"},
	{Name: "link"},
	{Name: "location"},
	{Name: "referer"},
	{Name: "set-cookie"},
	{Name: ":method", Value: "CONNECT"},
	{Name: ":method", Value: "DELETE"},
	{Name: ":method", Value: "GET"},
	{Name: ":method", Value: "HEAD"},
	{Name: ":method", Value: "OPTIONS"},
	{Name: ":method", Value: "POST"},
	{Name: ":method", Value: "PUT"},
	{Name: ":scheme", Value: "http"},
	{Name: ":scheme", Value: "https"},
	{Name: ":status", Value: "103"},
	{Name: ":status", Value: "200"},
	{Name: ":status", Value: "304"},
	{Name: ":status", Value: "404"},
	{Name: ":status", Value: "503"},
	{Name: "accept", Value: "*/*"},
	{Name: "accept", Value: "application/dns-message"},
	{Name: "accept-encoding", Value: "gzip, deflate, br"},
	{Name: "accept-ranges", Value: "bytes"},
	{Name: "access-control-allow-headers", Value: "cache-control"},
	{Name: "access-control-allow-headers", Value: "content-type"},
	{Name: "access-control-allow-origin", Value: "*"},
	{Name: "cache-control", Value: "max-age=0"},
	{Name: "cache-control", Value: "max-age=2592000"},
	{Name: "cache-control", Value: "max-age=604800"},
	{Name: "cache-control", Value: "no-cache"},
	{Name: "cache-control", Value: "no-store"},
	{Name: "cache-control", Value: "public, max-age=31536000"},
	{Name: "content-encoding", Value: "br"},
	{Name: "content-encoding", Value: "gzip"},
	{Name: "content-type", Value: "application/dns-message"},
	{Name: "content-type", Value: "application/javascript"},
	{Name: "content-type", Value: "application/json"},
	{Name: "content-type", Value: "application/x-www-form-urlencoded"},
	{Name: "content-type", Value: "image/gif"},
	{Name: "content-type", Value: "image/jpeg"},
	{Name: "content-type", Value: "image/png"},
	{Name: "content-type", Value: "text/css"},
	{Name: "content-type", Value: "text/html; charset=utf-8"},
	{Name: "content-type", Value: "text/plain"},
	{Name: "content-type", Value: "text/plain;charset=utf-8"},
	{Name: "range", Value: "bytes=0-"},
	{Name: "strict-transport-security", Value: "max-age=31536000"},
	{Name: "strict-transport-security", Value: "max-age=31536000; includesubdomains"},
	{Name: "strict-transport-security", Value: "max-age=31536000; includesubdomains; preload"},
	{Name: "vary", Value: "accept-encoding"},
	{Name: "vary", Value: "origin"},
	{Name: "x-content-type-options", Value: "nosniff"},
	{Name: "x-xss-protection", Value: "1; mode=block"},
	{Name: ":status", Value: "100"},
	{Name: ":status", Value: "204"},
	{Name: ":status", Value: "206"},
	{Name: ":status", Value: "302"},
	{Name: ":status", Value: "400"},
	{Name: ":status", Value: "403"},
	{Name: ":status", Value: "421"},
	{Name: ":status", Value: "425"},
	{Name: ":status", Value: "500"},
	{Name: "accept-language"},
	{Name: "access-control-allow-credentials", Value: "FALSE"},
	{Name: "access-control-allow-credentials", Value: "TRUE"},
	{Name: "access-control-allow-headers", Value: "*"},
	{Name: "access-control-allow-methods", Value: "get"},
	{Name: "access-control-allow-methods", Value: "get, post, options"},
	{Name: "access-control-allow-methods", Value: "options"},
	{Name: "access-control-expose-headers", Value: "content-length"},
	{Name: "access-control-request-headers", Value: "content-type"},
	{Name: "access-control-request-method", Value: "get"},
	{Name: "access-control-request-method", Value: "post"},
	{Name: "alt-svc", Value: "clear"},
	{Name: "authorization"},
	{Name: "content-security-policy", Value: "script-src 'none'; object-src 'none'; base-uri 'none'"},
	{Name: "early-data", Value: "1"},
	{Name: "expect-ct"},
	{Name: "forwarded"},
	{Name: "if-range"},
	{Name: "origin"},
	{Name: "purpose", Value: "prefetch"},
	{Name: "server"},
	{Name: "timing-allow-origin", Value: "*"},
	{Name: "upgrade-insecure-requests", Value: "1"},
	{Name: "user-agent"},
	{Name: "x-forwarded-for"},
	{Name: "x-frame-options", Value: "deny"},
	{Name: "x-frame-options", Value: "sameorigin"},
}

// Only needed for tests.
// use go:linkname to retrieve the static table.
//nolint:deadcode,unused
func getStaticTable() []HeaderField {
	return staticTableEntries[:]
}

type indexAndValues struct {
	idx    uint8
	values map[string]uint8
}

// A map of the header names from the static table to their index in the table.
// This is used by the encoder to quickly find if a header is in the static table
// and what value should be used to encode it.
// There's a second level of mapping for the headers that have some predefined
// values in the static table.
var encoderMap = map[string]indexAndValues{
	":authority":          {0, nil},
	":path":               {1, map[string]uint8{"/": 1}},
	"age":                 {2, map[string]uint8{"0": 2}},
	"content-disposition": {3, nil},
	"content-length":      {4, map[string]uint8{"0": 4}},
	"cookie":              {5, nil},
	"date":                {6, nil},
	"etag":                {7, nil},
	"if-modified-since":   {8, nil},
	"if-none-match":       {9, nil},
	"last-modified":       {10, nil},
	"link":                {11, nil},
	"location":            {12, nil},
	"referer":             {13, nil},
	"set-cookie":          {14, nil},
	":method": {15, map[string]uint8{
		"CONNECT": 15,
		"DELETE":  16,
		"GET":     17,
		"HEAD":    18,
		"OPTIONS": 19,
		"POST":    20,
		"PUT":     21,
	}},
	":scheme": {22, map[string]uint8{
		"http":  22,
		"https": 23,
	}},
	":status": {24, map[string]uint8{
		"103": 24,
		"200": 25,
		"304": 26,
		"404": 27,
		"503": 28,
		"100": 63,
		"204": 64,
		"206": 65,
		"302": 66,
		"400": 67,
		"403": 68,
		"421": 69,
		"425": 70,
		"500": 71,
	}},
	"accept": {29, map[string]uint8{
		"*/*":                     29,
		"application/dns-message": 30,
	}},
	"accept-encoding": {31, map[string]uint8{"gzip, deflate, br": 31}},
	"accept-ranges":   {32, map[string]uint8{"bytes": 32}},
	"access-control-allow-headers": {33, map[string]uint8{
		"cache-control": 33,
		"content-type":  34,
		"*":             75,
	}},
	"access-control-allow-origin": {35, map[string]uint8{"*": 35}},
	"cache-control": {36, map[string]uint8{
		"max-age=0":                36,
		"max-age=2592000":          37,
		"max-age=604800":           38,
		"no-cache":                 39,
		"no-store":                 40,
		"public, max-age=31536000": 41,
	}},
	"content-encoding": {42, map[string]uint8{
		"br":   42,
		"gzip": 43,
	}},
	"content-type": {44, map[string]uint8{
		"application/dns-message":           44,
		"application/javascript":            45,
		"application/json":                  46,
		"application/x-www-form-urlencoded": 47,
		"image/gif":                         48,
		"image/jpeg":                        49,
		"image/png":                         50,
		"text/css":                          51,
		"text/html; charset=utf-8":          52,
		"text/plain":                        53,
		"text/plain;charset=utf-8":          54,
	}},
	"range": {55, map[string]uint8{"bytes=0-": 55}},
	"strict-transport-security": {56, map[string]uint8{
		"max-age=31536000":                             56,
		"max-age=31536000; includesubdomains":          57,
		"max-age=31536000; includesubdomains; preload": 58,
	}},
	"vary": {59, map[string]uint8{
		"accept-encoding": 59,
		"origin":          60,
	}},
	"x-content-type-options": {61, map[string]uint8{"nosniff": 61}},
	"x-xss-protection":       {62, map[string]uint8{"1; mode=block": 62}},
	// ":status" is duplicated and takes index 63 to 71
	"accept-language": {72, nil},
	"access-control-allow-credentials": {73, map[string]uint8{
		"FALSE": 73,
		"TRUE":  74,
	}},
	// "access-control-allow-headers" is duplicated and takes index 75
	"access-control-allow-methods": {76, map[string]uint8{
		"get":                76,
		"get, post, options": 77,
		"options":            78,
	}},
	"access-control-expose-headers":  {79, map[string]uint8{"content-length": 79}},
	"access-control-request-headers": {80, map[string]uint8{"content-type": 80}},
	"access-control-request-method": {81, map[string]uint8{
		"get":  81,
		"post": 82,
	}},
	"alt-svc":       {83, map[string]uint8{"clear": 83}},
	"authorization": {84, nil},
	"content-security-policy": {85, map[string]uint8{
		"script-src 'none'; object-src 'none'; base-uri 'none'": 85,
	}},
	"early-data":                {86, map[string]uint8{"1": 86}},
	"expect-ct":                 {87, nil},
	"forwarded":                 {88, nil},
	"if-range":                  {89, nil},
	"origin":                    {90, nil},
	"purpose":                   {91, map[string]uint8{"prefetch": 91}},
	"server":                    {92, nil},
	"timing-allow-origin":       {93, map[string]uint8{"*": 93}},
	"upgrade-insecure-requests": {94, map[string]uint8{"1": 94}},
	"user-agent":                {95, nil},
	"x-forwarded-for":           {96, nil},
	"x-frame-options": {97, map[string]uint8{
		"deny":       97,
		"sameorigin": 98,
	}},
}
