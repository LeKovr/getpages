package pkgtest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/LeKovr/getpages/pkgtest"
)

func TestHandler(t *testing.T) {
	tests := map[string]struct {
		path       string
		want       []byte
		wantStatus int
	}{
		"normal":    {path: pkgtest.PageTimeout, want: pkgtest.TestBody, wantStatus: http.StatusOK},
		"not found": {path: pkgtest.PageNotFound, want: []byte("404\n"), wantStatus: http.StatusNotFound},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/"+tc.path, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(pkgtest.Handler)
			handler.ServeHTTP(rr, req)
			if status := rr.Code; status != tc.wantStatus {
				t.Errorf("status code not eq: got %v want %v",
					status, tc.wantStatus)
			}
			got := rr.Body.Bytes()
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("response body not eq: got %v want %v",
					rr.Body.String(), string(tc.want))
			}
		})
	}
}

func TestFillSource(t *testing.T) {
	filePath := pkgtest.FillSource(t, "")
	defer func() {
		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()
	_, err := os.Stat(filePath)
	if err != nil {
		t.Error(err)
	}
}
