// Copyright 2023 Redpanda Data, Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.md
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0

package admin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	testRequestHandler func(w http.ResponseWriter, r *http.Request)
	testFn             func(t *testing.T) testRequestHandler
	requestTestCase    struct {
		name             string
		testFn           testFn
		shouldError      bool
		expectedErrorMsg string
		expectedResponse interface{}
	}
)

func TestStartAutomatedRecovery(t *testing.T) {
	successfullStartResponse := RecoveryStartResponse{
		Code:    200,
		Message: "Automated recovery started",
	}

	ctx := context.Background()
	runTest := func(t *testing.T, test requestTestCase) {
		baseURL := "http://non-existent-url"

		if test.testFn != nil {
			server := httptest.NewServer(http.HandlerFunc(test.testFn(t)))
			defer server.Close()
			baseURL = server.URL
		}
		client, err := NewAdminAPI([]string{baseURL}, BasicCredentials{}, nil)
		assert.NoError(t, err)

		response, err := client.StartAutomatedRecovery(ctx, ".*")

		if test.shouldError {
			assert.Error(t, err)
			// Using assert.Contains instead of assert.Equal as some error messages change depending on the environment.
			assert.Contains(t, err.Error(), test.expectedErrorMsg)
			return
		}

		assert.NoError(t, err)
		assert.Equal(t, test.expectedResponse, response)
	}

	tests := []requestTestCase{
		{
			name: "should call the correct endpoint",
			testFn: func(t *testing.T) testRequestHandler {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/v1/cloud_storage/automated_recovery", r.URL.Path)
					w.WriteHeader(http.StatusOK)
					resp, err := json.Marshal(successfullStartResponse)
					assert.NoError(t, err)
					w.Write(resp)
				}
			},
			shouldError:      false,
			expectedResponse: successfullStartResponse,
		},
		{
			name: "should have content-type application-json",
			testFn: func(t *testing.T) testRequestHandler {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					w.WriteHeader(http.StatusOK)
					resp, err := json.Marshal(successfullStartResponse)
					assert.NoError(t, err)
					w.Write(resp)
				}
			},
			shouldError:      false,
			expectedResponse: successfullStartResponse,
		},
		{
			name: "should request recovery of all topics",
			testFn: func(t *testing.T) testRequestHandler {
				return func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)

					var recoveryRequestParams RecoveryRequestParams
					err = json.Unmarshal(body, &recoveryRequestParams)
					assert.NoError(t, err)

					assert.Equal(t, ".*", recoveryRequestParams.TopicNamesPattern)

					w.WriteHeader(http.StatusOK)
					resp, err := json.Marshal(successfullStartResponse)
					assert.NoError(t, err)
					w.Write(resp)
				}
			},
			shouldError:      false,
			expectedResponse: successfullStartResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runTest(t, test)
		})
	}
}
