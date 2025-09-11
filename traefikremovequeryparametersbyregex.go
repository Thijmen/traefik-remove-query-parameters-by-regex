// Package traefik_remove_query_parameters_by_regex by Thijmen Stavenuiter.
package traefik_remove_query_parameters_by_regex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

type modificationType string

const (
	deleteExceptType modificationType = "deleteexcept"
)

var (
	ErrInvalidModificationType = errors.New("invalid modification type, expected deleteexcept")
	ErrNoConfigurationSet      = errors.New(
		"either AllowedValuesRegex, ExceptURIRegex, or RedirectParams must be set",
	)
	ErrInvalidStatusCode = errors.New("redirectParam statusCode must be 301 or 302")
)

// RedirectParam represents a query parameter and its redirect status code.
type RedirectParam struct {
	Param      string `json:"param"`
	StatusCode int    `json:"statusCode"`
}

// Config is the configuration for this plugin.
type Config struct {
	RedirectParams            []RedirectParam  `json:"redirectParams"`
	AllowedValuesRegex        string           `json:"allowedValuesRegex"`
	ExceptURIRegex            string           `json:"exceptUriRegex"`
	Type                      modificationType `json:"type"`
	AddOriginalHostnameHeader bool             `json:"addOriginalHostnameHeader"`
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
		return nil, ErrInvalidModificationType
	}

	if config.AllowedValuesRegex == "" && config.ExceptURIRegex == "" &&
		len(config.RedirectParams) == 0 {
		return nil, ErrNoConfigurationSet
	}

	for _, redirectParam := range config.RedirectParams {
		if redirectParam.StatusCode != http.StatusMovedPermanently &&
			redirectParam.StatusCode != http.StatusFound {
			return nil, ErrInvalidStatusCode
		}
	}

	var exceptURIRegexCompiled *regexp.Regexp

	if config.ExceptURIRegex != "" {
		var err error

		exceptURIRegexCompiled, err = regexp.Compile(config.ExceptURIRegex)
		if err != nil {
			return nil, fmt.Errorf("failed to compile ExceptURIRegex: %w", err)
		}
	}

	var allowedValuesRegexCompiled *regexp.Regexp

	if config.AllowedValuesRegex != "" {
		var err error

		allowedValuesRegexCompiled, err = regexp.Compile(config.AllowedValuesRegex)
		if err != nil {
			return nil, fmt.Errorf("failed to compile AllowedValuesRegex: %w", err)
		}
	}

	// Build map for redirect parameters with status codes for O(1) lookup.
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

func (queryParameterRemoverConfig *QueryParameterRemover) ServeHTTP(
	rw http.ResponseWriter,
	req *http.Request,
) {
	qry := req.URL.Query()

	originalQuery := req.URL.String()

	// Check for redirect parameters first.
	if len(queryParameterRemoverConfig.redirectParamsMap) > 0 {
		for param := range qry {
			statusCode, exists := queryParameterRemoverConfig.redirectParamsMap[param]
			if !exists {
				continue
			}
			// Remove the matching parameter and redirect.
			qry.Del(param)
			req.URL.RawQuery = qry.Encode()
			redirectURL := req.URL.String()
			http.Redirect(rw, req, redirectURL, statusCode)

			return
		}
	}

	switch queryParameterRemoverConfig.config.Type {
	case deleteExceptType:
		if queryParameterRemoverConfig.config.ExceptURIRegex != "" {
			regexAllowed := queryParameterRemoverConfig.exceptURIRegexCompiled

			isExceptMatch := regexAllowed.MatchString(req.URL.String())

			if isExceptMatch {
				break
			}
		}

		regex := regexp.MustCompile(queryParameterRemoverConfig.config.AllowedValuesRegex)

		addOriginalHeader := false

		for param := range req.URL.Query() {
			if !regex.MatchString(param) {
				qry.Del(param)
				req.URL.RawQuery = qry.Encode()
				addOriginalHeader = true
			}
		}

		if queryParameterRemoverConfig.config.AddOriginalHostnameHeader && addOriginalHeader {
			req.Header.Add("Plugin-Original-Uri", originalQuery)
		}
	default:
		// No action needed for unknown types (validation happens in New function).
	}

	req.URL.RawQuery = qry.Encode()
	req.RequestURI = req.URL.RequestURI()

	queryParameterRemoverConfig.next.ServeHTTP(rw, req)
}

func (mt modificationType) isValid() bool {
	switch mt {
	case deleteExceptType, "":
		return true
	default:
		return false
	}
}
