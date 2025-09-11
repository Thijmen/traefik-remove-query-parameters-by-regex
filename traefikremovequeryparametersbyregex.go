// Package traefik_remove_query_parameters_by_regex by Thijmen Stavenuiter.
package traefik_remove_query_parameters_by_regex

import (
	"context"
	"errors"
	"net/http"
	"regexp"
)

type modificationType string

const (
	deleteExceptType modificationType = "deleteexcept"
)

// RedirectParam represents a query parameter and its redirect status code.
type RedirectParam struct {
	Param      string `json:"param"`
	StatusCode int    `json:"statusCode"`
}

// Config is the configuration for this plugin.
type Config struct {
	Type                      modificationType `json:"type"`
	AllowedValuesRegex        string           `json:"allowedValuesRegex"`
	ExceptURIRegex            string           `json:"exceptUriRegex"`
	AddOriginalHostnameHeader bool             `json:"addOriginalHostnameHeader"`
	RedirectParams            []RedirectParam  `json:"redirectParams"`
}

// CreateConfig creates a new configuration for this plugin.
func CreateConfig() *Config {
	return &Config{}
}

// QueryParameterRemover represents the basic properties of this plugin.
type QueryParameterRemover struct {
	next                       http.Handler
	name                       string
	config                     *Config
	exceptURIRegexCompiled     *regexp.Regexp
	allowedValuesRegexCompiled *regexp.Regexp
	redirectParamsMap          map[string]int
}

// New creates a new instance of this plugin.
func New(
	ctx context.Context,
	next http.Handler,
	config *Config,
	name string,
) (http.Handler, error) {
	if !config.Type.isValid() {
		return nil, errors.New("invalid modification type, expected deleteexcept")
	}

	if config.AllowedValuesRegex == "" && config.ExceptURIRegex == "" &&
		len(config.RedirectParams) == 0 {
		return nil, errors.New(
			"either AllowedValuesRegex, ExceptURIRegex, or RedirectParams must be set",
		)
	}

	for _, redirectParam := range config.RedirectParams {
		if redirectParam.StatusCode != 301 && redirectParam.StatusCode != 302 {
			return nil, errors.New("redirectParam statusCode must be 301 or 302")
		}
	}

	var exceptURIRegexCompiled *regexp.Regexp
	if config.ExceptURIRegex != "" {
		var err error
		exceptURIRegexCompiled, err = regexp.Compile(config.ExceptURIRegex)
		if err != nil {
			return nil, err
		}
	}

	var allowedValuesRegexCompiled *regexp.Regexp
	if config.AllowedValuesRegex != "" {
		var err error
		allowedValuesRegexCompiled, err = regexp.Compile(config.AllowedValuesRegex)
		if err != nil {
			return nil, err
		}
	}

	// Build map for redirect parameters with status codes for O(1) lookup
	redirectParamsMap := make(map[string]int)
	for _, redirectParam := range config.RedirectParams {
		redirectParamsMap[redirectParam.Param] = redirectParam.StatusCode
	}

	return &QueryParameterRemover{
		next:                       next,
		name:                       name,
		config:                     config,
		exceptURIRegexCompiled:     exceptURIRegexCompiled,
		allowedValuesRegexCompiled: allowedValuesRegexCompiled,
		redirectParamsMap:          redirectParamsMap,
	}, nil
}

func (q *QueryParameterRemover) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	qry := req.URL.Query()

	originalQuery := req.URL.String()

	// Check for redirect parameters first
	if len(q.redirectParamsMap) > 0 {
		for param := range qry {
			if statusCode, exists := q.redirectParamsMap[param]; exists {
				// Remove the matching parameter and redirect
				qry.Del(param)
				req.URL.RawQuery = qry.Encode()
				redirectURL := req.URL.String()
				http.Redirect(rw, req, redirectURL, statusCode)
				return
			}
		}
	}

	switch q.config.Type {
	case deleteExceptType:

		if q.config.ExceptURIRegex != "" {
			regexAllowed := q.exceptURIRegexCompiled

			isExceptMatch := regexAllowed.MatchString(req.URL.String())

			if isExceptMatch {
				break
			}
		}

		regex := regexp.MustCompile(q.config.AllowedValuesRegex)

		addOriginalHeader := false

		for param := range req.URL.Query() {
			if !regex.MatchString(param) {
				qry.Del(param)
				req.URL.RawQuery = qry.Encode()
				addOriginalHeader = true
			}
		}

		if q.config.AddOriginalHostnameHeader && addOriginalHeader {
			req.Header.Add("Plugin-Original-Uri", originalQuery)
		}
	}

	req.URL.RawQuery = qry.Encode()
	req.RequestURI = req.URL.RequestURI()

	q.next.ServeHTTP(rw, req)
}

func (mt modificationType) isValid() bool {
	switch mt {
	case deleteExceptType, "":
		return true
	}

	return false
}
