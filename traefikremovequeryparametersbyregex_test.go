package traefik_remove_query_parameters_by_regex_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Thijmen/traefik-remove-query-parameters-by-regex"
)

// region Delete
func TestDeleteQueryParam(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	expected := ""
	previous := "aa=1&bb=true"

	assertQueryModificationHelper(t, cfg, previous, expected, "/")
}

func TestDeleteQueryParamAndAllowIsNotRemoved(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	expected := "testing=1"
	previous := "aa=1&bb=true&testing=1"

	assertQueryModificationHelper(t, cfg, previous, expected, "/")
}

func TestDeleteQueryParamDoesntWorkOnProperDomain(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	cfg.ExceptURIRegex = "(qontrol)"
	expected := "aa=1&bb=true&testing=1"
	previous := "aa=1&bb=true&testing=1"

	assertQueryModificationHelper(t, cfg, previous, expected, "qontrol")
}

func TestDeleteQueryParamDoesntWorkOnProperDomainWithLongerPath(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	cfg.ExceptURIRegex = "(qontrol)"
	expected := "aa=1&bb=true&testing=1"
	previous := "aa=1&bb=true&testing=1"

	assertQueryModificationHelper(t, cfg, previous, expected, "/qontrol/test/1")
}

func TestDeleteQueryParamDoesNotWorkWithRegexWithADash(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|x-live-preview)"
	cfg.ExceptURIRegex = "(qontrol)"
	expected := "x-live-preview=1"
	previous := "test=1&x-live-preview=1"


	assertQueryModificationHelper(t, cfg, previous, expected, "")
}

func TestErrorInvalidType(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "bla"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "query-params-remover-plugin")

	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func TestErrorNoParam(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "delete"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "query-modification-plugin")

	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func createReqAndRecorder(cfg *traefik_remove_query_parameters_by_regex.Config) (http.Handler, *httptest.ResponseRecorder, *http.Request, error) {
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	handler, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "query-modification-plugin")
	if err != nil {
		return nil, nil, nil, err
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	return handler, recorder, req, err
}

func assertQueryModificationHelper(t *testing.T, cfg *traefik_remove_query_parameters_by_regex.Config, previous, expected, uriPath string) {
	t.Helper()
	handler, recorder, req, err := createReqAndRecorder(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	req.URL.RawQuery = previous
	req.URL.Path = uriPath
	handler.ServeHTTP(recorder, req)

	if req.URL.Query().Encode() != expected {
		t.Errorf("Expected %s, got %s", expected, req.URL.Query().Encode())
	}
}
