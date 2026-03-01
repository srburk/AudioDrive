package model_test

import (
	"testing"

	"audiodrive/internal/model"
)

func TestURL_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{name: "valid http", rawURL: "http://example.com", wantErr: false},
		{name: "valid https", rawURL: "https://example.com/path?q=1", wantErr: false},
		{name: "empty string", rawURL: "", wantErr: true},
		{name: "relative path", rawURL: "/just/a/path", wantErr: true},
		{name: "ftp scheme", rawURL: "ftp://example.com", wantErr: true},
		{name: "no scheme", rawURL: "example.com", wantErr: true},
		{name: "scheme only", rawURL: "https://", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := model.URL{RawURL: tc.rawURL}
			err := u.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("expected error for %q, got nil", tc.rawURL)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error for %q, got %v", tc.rawURL, err)
			}
		})
	}
}
