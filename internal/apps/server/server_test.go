package server

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer(t *testing.T) {
	type want struct {
		code        int
		value       string
		contentType string
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "case 1: update gauge value",
			args: args{method: http.MethodPost, path: "/update/gauge/key/3.0000003"},
			want: want{code: http.StatusOK},
		},
		{
			name: "case 2: update gauge value",
			args: args{method: http.MethodPost, path: "/update/gauge/key/4.0000004"},
			want: want{code: http.StatusOK},
		},
		{
			name: "case 3: get gauge value",
			args: args{method: http.MethodGet, path: "/value/gauge/key"},
			want: want{code: http.StatusOK, value: "4.0000004", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "case 4: update counter value",
			args: args{method: http.MethodPost, path: "/update/counter/key/1"},
			want: want{code: http.StatusOK},
		},
		{
			name: "case 5: update counter value",
			args: args{method: http.MethodPost, path: "/update/counter/key/2"},
			want: want{code: http.StatusOK},
		},
		{
			name: "case 6: get counter value",
			args: args{method: http.MethodGet, path: "/value/counter/key"},
			want: want{code: http.StatusOK, value: "3", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "case 7: invalid counter value",
			args: args{method: http.MethodPost, path: "/update/counter/key/1.00001"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "case 8: invalid counter value",
			args: args{method: http.MethodPost, path: "/update/counter/key/none"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "case 9: invalid counter path",
			args: args{method: http.MethodPost, path: "/update/counter/key/"},
			want: want{code: http.StatusNotFound, value: "404 page not found\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "case 10: invalid gauge value",
			args: args{method: http.MethodPost, path: "/update/counter/key/none"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "case 11: invalid gauge path",
			args: args{method: http.MethodPost, path: "/update/counter/key/"},
			want: want{code: http.StatusNotFound, value: "404 page not found\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name: "case 12: get report",
			args: args{method: http.MethodGet, path: "/"},
			want: want{code: http.StatusOK, value: "<p>counter_key: 3\n</p><p>gauge_key: 4.0000004\n</p>", contentType: "text/html; charset=utf-8"},
		},
	}

	ts := httptest.NewServer(createRouter())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.args.method, tt.args.path)
			_ = resp.Body.Close()

			assert.Equal(t, tt.want.code, resp.StatusCode, "Unexpected status code")
			assert.Equal(t, tt.want.value, body)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
) (resp *http.Response, body string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err = ts.Client().Do(req)
	require.NoError(t, err)
	defer func(body io.Closer) { _ = body.Close() }(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.NoError(t, err)

	return resp, string(respBody)
}
