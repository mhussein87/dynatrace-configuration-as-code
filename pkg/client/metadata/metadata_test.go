//go:build unit

/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metadata

import (
	"github.com/dynatrace/dynatrace-configuration-as-code/v2/pkg/rest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDynatraceClassicEnvironment(t *testing.T) {

	tests := []struct {
		name                 string
		serverResponse       string
		serverResponseStatus int
		want                 string
		wantErr              bool
	}{
		{
			name:                 "server responds with code != 200",
			serverResponseStatus: http.StatusNotFound,
			want:                 "",
			wantErr:              true,
		},
		{
			name:                 "server response with invalid data",
			serverResponseStatus: http.StatusOK,
			serverResponse:       "}",
			want:                 "",
			wantErr:              true,
		},
		{
			name:                 "server response with valid data",
			serverResponseStatus: http.StatusOK,
			serverResponse:       `{"domain" : "http://classic.env.com"}`,
			want:                 "http://classic.env.com",
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path == ClassicEnvironmentDomainPath {
					rw.WriteHeader(tt.serverResponseStatus)
					_, _ = rw.Write([]byte(tt.serverResponse))
				} else {
					rw.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			got, err := GetDynatraceClassicURL(context.TODO(), rest.NewRestClient(&http.Client{}, nil, rest.CreateRateLimitStrategy()), server.URL)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)

		})
	}
}

func TestGetDynatraceClassicEnvironmentWorksWithTrailingSlash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == ClassicEnvironmentDomainPath {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte(`{"domain" : "http://classic.env.com"}`))
		} else {
			rw.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	got, err := GetDynatraceClassicURL(context.TODO(), rest.NewRestClient(&http.Client{}, nil, rest.CreateRateLimitStrategy()), server.URL+"/")
	assert.Equal(t, "http://classic.env.com", got)
	assert.NoError(t, err)
}

func TestGetDynatraceClassicEnvironmentFallsBackToDeprecatedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == ClassicEnvironmentDomainPath {
			rw.WriteHeader(http.StatusNotFound)
			_, _ = rw.Write([]byte("<html>Some useless response</html>"))
		} else if req.URL.Path == DeprecatedClassicEnvDomainPath {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte(`{"endpoint" : "http://fallback.classic.env.com"}`))
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	got, err := GetDynatraceClassicURL(context.TODO(), rest.NewRestClient(&http.Client{}, nil, rest.CreateRateLimitStrategy()), server.URL+"/")
	assert.Equal(t, "http://fallback.classic.env.com", got)
	assert.NoError(t, err)
}
