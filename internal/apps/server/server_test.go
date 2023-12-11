package server

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerUpdate(t *testing.T) {
	type want struct {
		code int
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
			name: "success gauge case",
			args: args{method: http.MethodPost, path: "/update/gauge/key/3.0000003"},
			want: want{code: http.StatusOK},
		},
		{
			name: "success counter case",
			args: args{method: http.MethodPost, path: "/update/counter/key/1"},
			want: want{code: http.StatusOK},
		},
		{
			name: "invalid counter value case 1",
			args: args{method: http.MethodPost, path: "/update/counter/key/1.00001"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid counter value case 2",
			args: args{method: http.MethodPost, path: "/update/counter/key/"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid counter value case 3",
			args: args{method: http.MethodPost, path: "/update/counter/key/none"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid gauge value case 1",
			args: args{method: http.MethodPost, path: "/update/counter/key/none"},
			want: want{code: http.StatusBadRequest},
		},
		{
			name: "invalid gauge value case 2",
			args: args{method: http.MethodPost, path: "/update/counter/key/"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 1",
			args: args{method: http.MethodPost, path: "/"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 2",
			args: args{method: http.MethodPost, path: "/update"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 3",
			args: args{method: http.MethodPost, path: "/update/"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 4",
			args: args{method: http.MethodPost, path: "/update/counter"},
			want: want{code: http.StatusNotFound},
		},
		{
			name: "invalid path 5",
			args: args{method: http.MethodPost, path: "/update/counter/"},
			want: want{code: http.StatusNotFound},
		},
	}

	ts := httptest.NewServer(createRouter())
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, tt.args.method, tt.args.path)
			assert.Equal(t, tt.want.code, resp.StatusCode, "Unexpected status code")
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
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
