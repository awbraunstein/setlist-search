package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type data struct {
	A string `json:"a"`
	B string `json:"b"`
}

func TestMergeJsonBody(t *testing.T) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		var d data
		if err := MergeJSONBody(c, &d); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
		}
		return c.JSON(http.StatusOK, d)
	})
	tests := []struct {
		name string
		url  string
		body string
		want string
		err  bool
	}{
		{
			name: "no param, no body",
			url:  "/",
			body: "",
			want: `{"a":"","b":""}`,
		}, {
			name: "a param, no body",
			url:  "/?a=foo",
			body: "",
			want: `{"a":"foo","b":""}`,
		}, {
			name: "a param, a in body",
			url:  "/?a=foo",
			body: `{"a":"bar"}`,
			want: `{"a":"foo","b":""}`,
		}, {
			name: "b param, a in body",
			url:  "/?b=foo",
			body: `{"a":"bar"}`,
			want: `{"a":"bar","b":"foo"}`,
		}, {
			name: "bad body",
			url:  "/?b=foo",
			body: `[]`,
			err:  true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if tc.err && rec.Code != http.StatusBadRequest {
				t.Fatalf("Expected bad BadRequest, but got %d", rec.Code)
			}
			if !tc.err {
				if rec.Code != http.StatusOK {
					t.Fatalf("Expected OK status, but got %d", rec.Code)
				}
				if got := strings.TrimSpace(rec.Body.String()); got != tc.want {
					t.Fatalf("Expected %s, but got: %s", tc.want, got)
				}
			}
		})
	}
}
