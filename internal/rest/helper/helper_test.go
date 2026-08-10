package helper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
		want struct {
			Name string `json:"name"`
		}
		wantErr bool
	}{
		{name: "decodes one JSON object", body: `{"name":"Ada"}`, want: struct {
			Name string `json:"name"`
		}{Name: "Ada"}},
		{name: "allows trailing whitespace", body: "{\"name\":\"Ada\"}\n  ", want: struct {
			Name string `json:"name"`
		}{Name: "Ada"}},
		{name: "rejects an empty body", body: ``, wantErr: true},
		{name: "rejects malformed JSON", body: `{"name":`, wantErr: true},
		{name: "rejects unknown fields", body: `{"name":"Ada","age":30}`, wantErr: true},
		{name: "rejects an invalid field type", body: `{"name":42}`, wantErr: true},
		{name: "rejects a second JSON value", body: `{"name":"Ada"} {"name":"Grace"}`, wantErr: true},
		{name: "rejects trailing content", body: `{"name":"Ada"} trailing`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got struct {
				Name string `json:"name"`
			}

			err := DecodeJSON(strings.NewReader(tt.body), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecodeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DecodeJSON() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteError(recorder, http.StatusBadRequest, "invalid_json", "invalid body")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var got ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := ErrorResponse{Error: ErrorDetail{Code: "invalid_json", Message: "invalid body"}}
	if got != want {
		t.Errorf("response = %+v, want %+v", got, want)
	}
}
