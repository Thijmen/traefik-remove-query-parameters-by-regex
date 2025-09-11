package traefik_remove_query_parameters_by_regex_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	traefik_remove_query_parameters_by_regex "github.com/Thijmen/traefik-remove-query-parameters-by-regex"
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

func TestItSendsTheHeader(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(testing|x-live-preview)"
	cfg.ExceptURIRegex = "(qontrol)"
	cfg.AddOriginalHostnameHeader = true
	headerValue := "http://localhost?test=1&x-live-preview=1"
	previous := "test=1&x-live-preview=1"

	assertHeaderValue(t, cfg, previous, headerValue)
}

func TestErrorInvalidType(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "bla"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := traefik_remove_query_parameters_by_regex.New(
		ctx,
		next,
		cfg,
		"query-params-remover-plugin",
	)
	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func TestErrorNoParam(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "delete"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := traefik_remove_query_parameters_by_regex.New(
		ctx,
		next,
		cfg,
		"query-modification-plugin",
	)
	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func createReqAndRecorder(
	cfg *traefik_remove_query_parameters_by_regex.Config,
) (http.Handler, *httptest.ResponseRecorder, *http.Request, error) {
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := traefik_remove_query_parameters_by_regex.New(
		ctx,
		next,
		cfg,
		"query-modification-plugin",
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create handler: %w", err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", http.NoBody)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	return handler, recorder, req, nil
}

func assertQueryModificationHelper(
	t *testing.T,
	cfg *traefik_remove_query_parameters_by_regex.Config,
	previous, expected, uriPath string,
) {
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

func assertHeaderValue(
	t *testing.T,
	cfg *traefik_remove_query_parameters_by_regex.Config,
	previous, expectedHeaderValue string,
) {
	t.Helper()

	handler, recorder, req, err := createReqAndRecorder(cfg)
	if err != nil {
		t.Fatal(err)

		return
	}

	req.URL.RawQuery = previous
	handler.ServeHTTP(recorder, req)

	header := req.Header.Get("Plugin-Original-Uri")

	if header != expectedHeaderValue {
		t.Errorf("Expected %s, got %s", expectedHeaderValue, header)
	}
}

// region Redirect Tests
func TestRedirect301WithSingleParameter(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 301},
	}
	expectedLocation := "http://localhost?other=value"
	previous := "utm_source=google&other=value"

	assertRedirectHelper(t, cfg, previous, expectedLocation, 301)
}

func TestRedirect302WithSingleParameter(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_campaign", StatusCode: 302},
	}
	expectedLocation := "http://localhost?other=value"
	previous := "utm_campaign=summer&other=value"

	assertRedirectHelper(t, cfg, previous, expectedLocation, 302)
}

func TestRedirectWithMultipleRedirectParams(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 301},
		{Param: "utm_campaign", StatusCode: 302},
		{Param: "fbclid", StatusCode: 301},
	}
	expectedLocation := "http://localhost?keep=this"
	previous := "utm_source=google&keep=this"

	assertRedirectHelper(t, cfg, previous, expectedLocation, 301)
}

func TestRedirectRemovesAllInstancesOfParameter(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "tracking", StatusCode: 302},
	}
	expectedLocation := "http://localhost?other=value"
	previous := "tracking=param1&other=value&tracking=param2"

	assertRedirectHelper(t, cfg, previous, expectedLocation, 302)
}

func TestNoRedirectWhenNoMatchingParams(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(allowed)"
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 301},
	}
	previous := "allowed=value&other=value"

	handler, recorder, req, err := createReqAndRecorder(cfg)
	if err != nil {
		t.Fatal(err)

		return
	}

	req.URL.RawQuery = previous
	handler.ServeHTTP(recorder, req)

	// Should not redirect, should process normally
	if recorder.Code == 301 || recorder.Code == 302 {
		t.Errorf("Expected no redirect, but got status %d", recorder.Code)
	}

	// Should have processed the deleteexcept logic
	if req.URL.Query().Encode() != "allowed=value" {
		t.Errorf("Expected allowed=value, got %s", req.URL.Query().Encode())
	}
}

func TestRedirectTakesPrecedenceOverDeleteExcept(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.Type = "deleteexcept"
	cfg.AllowedValuesRegex = "(allowed)"
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 301},
	}
	expectedLocation := "http://localhost?allowed=value&other=value"
	previous := "utm_source=google&allowed=value&other=value"

	assertRedirectHelper(t, cfg, previous, expectedLocation, 301)
}

func TestRedirectWithOnlyRedirectParam(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 302},
	}
	expectedLocation := "http://localhost"
	previous := "utm_source=google"

	assertRedirectHelper(t, cfg, previous, expectedLocation, 302)
}

func TestErrorInvalidRedirectStatusCode(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 200},
	}
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "redirect-plugin")

	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func TestErrorInvalidRedirectStatusCode404(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 404},
	}
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "redirect-plugin")
	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func TestRedirectOnlyConfigurationIsValid(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 301},
	}
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_remove_query_parameters_by_regex.New(ctx, next, cfg, "redirect-plugin")
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}

func TestRedirectWithDifferentStatusCodesPerParameter(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "utm_source", StatusCode: 301},
		{Param: "fbclid", StatusCode: 302},
	}

	// Test utm_source redirects with 301
	expectedLocation := "http://localhost?keep=this"
	previous := "utm_source=google&keep=this"
	assertRedirectHelper(t, cfg, previous, expectedLocation, 301)

	// Test fbclid redirects with 302
	expectedLocation = "http://localhost?keep=this"
	previous = "fbclid=abc123&keep=this"
	assertRedirectHelper(t, cfg, previous, expectedLocation, 302)
}

func TestRedirectWhileMaintainingQueryParameters(t *testing.T) {
	cfg := traefik_remove_query_parameters_by_regex.CreateConfig()
	cfg.RedirectParams = []traefik_remove_query_parameters_by_regex.RedirectParam{
		{Param: "token", StatusCode: 301},
	}

	// Test utm_source redirects with 301
	expectedLocation := "http://localhost?x-some-other-token=this"
	previous := "token=aaaAAxxx-BVS_A&x-some-other-token=this"
	assertRedirectHelper(t, cfg, previous, expectedLocation, 301)
}

func assertRedirectHelper(
	t *testing.T,
	cfg *traefik_remove_query_parameters_by_regex.Config,
	previous, expectedLocation string,
	expectedStatus int,
) {
	t.Helper()
	handler, recorder, req, err := createReqAndRecorder(cfg)
	if err != nil {
		t.Fatal(err)

		return
	}
	req.URL.RawQuery = previous
	handler.ServeHTTP(recorder, req)

	if recorder.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("Expected location %s, got %s", expectedLocation, location)
	}
}
