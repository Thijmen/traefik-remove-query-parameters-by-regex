package traefik_remove_query_parameters_by_regex_test

import (
	"context"
	traefik_remove_query_parameters_by_regex "github.com/Thijmen/traefik-remove-query-parameters-by-regex"
	"net/http"
	"net/http/httptest"
	"testing"
)

// region Delete
func TestDeleteQueryParam(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	expected := ""
	previous := "aa=1&bb=true"

	assertQueryModification(t, cfg, previous, expected, "/")
}

func TestDeleteQueryParamAndAllowIsNotRemoved(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	expected := "testing=1"
	previous := "aa=1&bb=true&testing=1"

	assertQueryModification(t, cfg, previous, expected, "/")
}

func TestDeleteQueryParamDoesntWorkOnProperDomain(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	cfg.ExceptUriRegex = "(qontrol)"
	expected := "aa=1&bb=true&testing=1"
	previous := "aa=1&bb=true&testing=1"

	assertQueryModification(t, cfg, previous, expected, "qontrol")
}

func TestDeleteQueryParamDoesntWorkOnProperDomainWithLongerPath(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|debugging)"
	cfg.ExceptUriRegex = "(qontrol)"
	expected := "aa=1&bb=true&testing=1"
	previous := "aa=1&bb=true&testing=1"

	assertQueryModification(t, cfg, previous, expected, "/qontrol/test/1")
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

func createReqAndRecorder(cfg *traefik_remove_query_parameters_by_regex.Config) (http.Handler, error, *httptest.ResponseRecorder, *http.Request) {
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	handler, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "query-modification-plugin")
	if err != nil {
		return nil, err, nil, nil
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	return handler, recorder, req, err
}

func assertQueryModification(t *testing.T, cfg *traefik_remove_query_parameters_by_regex.Config, previous, expected string, uriPath string) {
	handler, err, recorder, req := createReqAndRecorder(cfg)
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
