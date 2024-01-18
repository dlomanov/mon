package server

import (
	"bytes"
	"github.com/dlomanov/mon/internal/apps/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "set gauge",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json; charset=utf-8",
				body:        `{"id":"key","type":"gauge","value":3.0000003}`,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"key","type":"gauge","value":3.0000003}`,
			},
		},
		{
			name: "update gauge",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"gauge","value":4.0000004}`,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"key","type":"gauge","value":4.0000004}`,
			},
		},
		{
			name: "get gauge",
			args: args{
				method:      http.MethodPost,
				path:        "/value/",
				contentType: "application/json",
				body:        `{"id":"key","type":"gauge"}`,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"key","type":"gauge","value":4.0000004}`,
			},
		},
		{
			name: "set counter",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":1}`,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":1}`,
			},
		},
		{
			name: "update counter",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":2}`,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":3}`,
			},
		},
		{
			name: "get counter",
			args: args{
				method:      http.MethodPost,
				path:        "/value/",
				contentType: "application/json",
				body:        `{"id":"key","type":"counter"}`,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":3}`,
			},
		},
		{
			name: "set invalid value for type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json; charset=utf-8",
				body:        `{"id":"key","type":"gauge","delta":3.0000003}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "set invalid value for type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"gauge"}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "set invalid value for type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":"2"}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "set invalid value for type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"counter","delta":2.0}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "set invalid value for type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/json",
				body:        `{"id":"key","type":"counter"}`,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "set value for type with invalid content-type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/xml",
				body:        `{"id":"key","type":"counter","delta":2}`,
			},
			want: want{
				code: http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "set value for type with invalid content-type",
			args: args{
				method:      http.MethodPost,
				path:        "/update/",
				contentType: "application/xml",
				body:        `{"id":"key","type":"gauge","delta":2.0}`,
			},
			want: want{
				code: http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "get report",
			args: args{method: http.MethodGet, path: "/"},
			want: want{code: http.StatusOK, body: "<p>counter_key: 3\n</p><p>gauge_key: 4.0000004\n</p>", contentType: "text/html; charset=utf-8"},
		},
	}

	db := mocks.NewStorage()
	r := createRouter(db)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.args)
			_ = resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode, "Unexpected status code")
			assert.Equal(t, tt.want.body, strings.TrimSuffix(body, "\n"))
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

type args struct {
	method      string
	path        string
	contentType string
	body        string
}

type want struct {
	code        int
	body        string
	contentType string
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	args args,
) (resp *http.Response, responsePayload string) {
	t.Helper()

	bodyReader := bytes.NewReader([]byte(args.body))
	req, err := http.NewRequest(args.method, ts.URL+args.path, bodyReader)
	require.NoError(t, err)
	if args.contentType != "" {
		req.Header.Set("Content-Type", args.contentType)
	}

	resp, err = ts.Client().Do(req)
	require.NoError(t, err)
	defer func(body io.Closer) { _ = body.Close() }(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.NoError(t, err)

	return resp, string(respBody)
}
