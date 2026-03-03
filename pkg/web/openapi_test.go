package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// handleOpenAPISpec
// ---------------------------------------------------------------------------

func TestHandleOpenAPISpec_ReturnsJSON(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q; want %q", ct, "application/json")
	}
}

func TestHandleOpenAPISpec_ValidJSON(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("OpenAPI spec is not valid JSON: %v", err)
	}
}

func TestHandleOpenAPISpec_HasRequiredFields(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Check required top-level fields per OpenAPI 3.0 specification
	for _, field := range []string{"openapi", "info", "paths"} {
		if _, ok := spec[field]; !ok {
			t.Errorf("OpenAPI spec missing required field %q", field)
		}
	}

	// Check openapi version
	version, ok := spec["openapi"].(string)
	if !ok {
		t.Fatal("openapi field is not a string")
	}
	if !strings.HasPrefix(version, "3.0") {
		t.Errorf("openapi version = %q; want 3.0.x", version)
	}
}

func TestHandleOpenAPISpec_InfoSection(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec struct {
		Info struct {
			Title   string `json:"title"`
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if spec.Info.Title == "" {
		t.Error("info.title should not be empty")
	}
	if spec.Info.Version == "" {
		t.Error("info.version should not be empty")
	}
}

func TestHandleOpenAPISpec_PathsExist(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec struct {
		Paths map[string]interface{} `json:"paths"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify key API paths are documented
	expectedPaths := []string{
		"/api/v1/health",
		"/api/v1/auth/login",
		"/api/v1/tasks",
		"/api/v1/chat/sessions",
		"/api/v1/skills",
	}
	for _, path := range expectedPaths {
		if _, ok := spec.Paths[path]; !ok {
			t.Errorf("OpenAPI spec missing expected path %q", path)
		}
	}
}

func TestHandleOpenAPISpec_SecurityScheme(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec struct {
		Components struct {
			SecuritySchemes map[string]struct {
				Type   string `json:"type"`
				Scheme string `json:"scheme"`
			} `json:"securitySchemes"`
		} `json:"components"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	bearer, ok := spec.Components.SecuritySchemes["bearerAuth"]
	if !ok {
		t.Fatal("securitySchemes missing bearerAuth")
	}
	if bearer.Type != "http" {
		t.Errorf("bearerAuth type = %q; want %q", bearer.Type, "http")
	}
	if bearer.Scheme != "bearer" {
		t.Errorf("bearerAuth scheme = %q; want %q", bearer.Scheme, "bearer")
	}
}

// ---------------------------------------------------------------------------
// handleAPIDocsUI
// ---------------------------------------------------------------------------

func TestHandleAPIDocsUI_ReturnsHTML(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	s.handleAPIDocsUI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q; want text/html", ct)
	}
}

func TestHandleAPIDocsUI_ContainsSwaggerUI(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	s.handleAPIDocsUI(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "swagger-ui") {
		t.Error("Swagger UI HTML should contain 'swagger-ui'")
	}
	if !strings.Contains(body, "openapi.json") {
		t.Error("Swagger UI HTML should reference openapi.json")
	}
}

func TestHandleAPIDocsUI_ContainsTitle(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	s.handleAPIDocsUI(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "<title>") {
		t.Error("Swagger UI HTML should contain a <title> tag")
	}
}

// ---------------------------------------------------------------------------
// OpenAPI spec schema definitions
// ---------------------------------------------------------------------------

func TestHandleOpenAPISpec_SchemasExist(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec struct {
		Components struct {
			Schemas map[string]interface{} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	expectedSchemas := []string{"Task", "ChatMessage", "Error"}
	for _, schema := range expectedSchemas {
		if _, ok := spec.Components.Schemas[schema]; !ok {
			t.Errorf("OpenAPI spec missing expected schema %q", schema)
		}
	}
}

// ---------------------------------------------------------------------------
// Health endpoint is documented as public (no security)
// ---------------------------------------------------------------------------

func TestHandleOpenAPISpec_HealthEndpointNoSecurity(t *testing.T) {
	s := newTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	s.handleOpenAPISpec(rr, req)

	var spec struct {
		Paths map[string]struct {
			Get struct {
				Security []map[string]interface{} `json:"security"`
			} `json:"get"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	healthPath, ok := spec.Paths["/api/v1/health"]
	if !ok {
		t.Fatal("health path not found in spec")
	}

	// Health endpoint should have security: [] (empty array = no auth required)
	if len(healthPath.Get.Security) != 0 {
		t.Errorf("health endpoint security should be empty (public), got %v", healthPath.Get.Security)
	}
}
