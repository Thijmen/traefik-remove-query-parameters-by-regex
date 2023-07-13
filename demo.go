package traefik_remove_query_parameters_by_regex

import (
	"context"
	"errors"
	"log"
	"net/http"
	"regexp"
)

type modificationType string

const (
	deleteExceptType modificationType = "deleteexcept"
)

// Config is the configuration for this plugin.
type Config struct {
	Type               modificationType `json:"type"`
	AllowedValuesRegex string           `json:"allowedValuesRegex"`
	ExceptURIRegex     string           `json:"exceptUriRegex"`
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
}

// New creates a new instance of this plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if !config.Type.isValid() {
		return nil, errors.New("invalid modification type, expected deleteexcept")
	}

	if config.AllowedValuesRegex == "" && config.ExceptURIRegex == "" {
		return nil, errors.New("either AllowedValuesRegex or ExceptURIRegex must be set")
	}

	var exceptURIRegexCompiled *regexp.Regexp = nil
	if config.ExceptURIRegex != "" {
		var err error
		exceptURIRegexCompiled, err = regexp.Compile(config.ExceptURIRegex)
		if err != nil {
			return nil, err
		}
	}

	var allowedValuesRegexCompiled *regexp.Regexp = nil
	if config.AllowedValuesRegex != "" {
		var err error
		allowedValuesRegexCompiled, err = regexp.Compile(config.AllowedValuesRegex)
		if err != nil {
			return nil, err
		}
	}

	return &QueryParameterRemover{
		next:                       next,
		name:                       name,
		config:                     config,
		exceptURIRegexCompiled:     exceptURIRegexCompiled,
		allowedValuesRegexCompiled: allowedValuesRegexCompiled,
	}, nil
}

func (q *QueryParameterRemover) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	qry := req.URL.Query()

	switch q.config.Type {
	case deleteExceptType:

		if q.config.ExceptUriRegex != "" {
			regexAllowed := q.exceptURIRegexCompiled

			isExceptMatch := regexAllowed.MatchString(req.URL.String())

			if isExceptMatch {
				break
			}
		}

		regex := regexp.MustCompile(q.config.AllowedValuesRegex)

		for param, _ := range req.URL.Query() {
			if !regex.MatchString(param) {
				qry.Del(param)
				req.URL.RawQuery = qry.Encode()
				log.Printf("Removed parameter: %s \n", param)
			}
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