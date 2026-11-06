package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	app := NewApp("test")
	NewHealthHandler("user", "replica-1").Register(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	want := healthResponse{Status: "ok", Service: "user", Instance: "replica-1"}
	if body != want {
		t.Fatalf("body = %+v, want %+v", body, want)
	}
}

func TestUnknownRouteReturnsErrorContract(t *testing.T) {
	app := NewApp("test")

	resp, err := app.Test(httptest.NewRequest("GET", "/nope", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}

	var body errorBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "not_found" {
		t.Fatalf("error.code = %q, want not_found", body.Error.Code)
	}
}
