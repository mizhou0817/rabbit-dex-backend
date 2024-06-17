package api_client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/strips-finance/rabbit-dex-backend/api"
)

type ProfileDataSuite struct {
	APITestSuite
}

func (s *ProfileDataSuite) TestProfileDataNoAuth() {
	t := s.T()
	url := s.Client().apiUrl + "/storage/profile_data"

	t.Run("Get", func(t *testing.T) {
		resp, err := s.Client().httpClient.Get(url)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"should not be able to get profile data without auth")
		var data api.Response[any]
		defer resp.Body.Close()
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
		assert.Equal(t, false, data.Success)
		assert.NotEmpty(t, data.Error)
		assert.Empty(t, data.Result)
	})
	t.Run("Post", func(t *testing.T) {
		resp, err := s.Client().httpClient.Post(url, "application/json",
			bytes.NewReader([]byte(`{"version":1,"data":{"a":1}}`)),
		)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"should not be able to post profile data without auth")
		defer resp.Body.Close()
		var data api.Response[any]
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
		assert.Equal(t, false, data.Success)
		assert.NotEmpty(t, data.Error)
		assert.Empty(t, data.Result)
	})
}

func (s *ProfileDataSuite) TestProfileDataWithAuth() {
	t := s.T()
	_, err := s.Client().Onboarding()
	require.NoError(t, err)

	type step struct {
		name   string
		op     string
		req    string
		resp   string
		status int
	}

	url := s.client.apiUrl + "/storage/profile_data"
	steps := []step{
		{
			name:   "Initial Get request",
			op:     "get",
			resp:   `{"success":true,"error":"","result":[{"version":0,"data":{}}]}`,
			status: http.StatusOK,
		},
		{
			name:   "Subsequent Get request",
			op:     "get",
			resp:   `{"success":true,"error":"","result":[{"version":0,"data":{}}]}`,
			status: http.StatusOK,
		},
		{
			name: "Initial Post request",
			op:   "post",
			req:  `{"version":999,"data":{"a":1}}`,
			// expect response with version=1 for the first ever post request
			// no matter what version is in the request
			resp:   `{"success":true,"error":"","result":[{"version":1,"data":{"a":1}}]}`,
			status: http.StatusOK,
		},
		{
			name:   "Get after Post",
			op:     "get",
			resp:   `{"success":true,"error":"","result":[{"version":1,"data":{"a":1}}]}`,
			status: http.StatusOK,
		},
		{
			name:   "Bad version in Post request - 0",
			op:     "post",
			req:    `{"version":0,"data":{"a":1}}`,
			resp:   `{"success":false,"error":"version in db does not match version in request","result":[{"version":1,"data":{"a":1}}]}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "Bad version in Post request - 2",
			op:     "post",
			req:    `{"version":2,"data":{"a":1}}`,
			resp:   `{"success":false,"error":"version in db does not match version in request","result":[{"version":1,"data":{"a":1}}]}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "Second Post request",
			op:     "post",
			req:    `{"version":1,"data":{"a":2,"b":3}}`,
			resp:   `{"success":true,"error":"","result":[{"version":2,"data":{"a":2,"b":3}}]}`,
			status: http.StatusOK,
		},
		{
			name:   "empty data",
			op:     "post",
			req:    `{}`,
			resp:   `{"success":false,"error":"Key: 'ProfileData.Data' Error:Field validation for 'Data' failed on the 'required' tag","result":[]}`,
			status: http.StatusBadRequest,
		},
		{
			name:   "missing version",
			op:     "post",
			req:    `{"data":{"a":2,"b":3}}`,
			resp:   `{"success":false,"error":"version in db does not match version in request","result":[{"version":2,"data":{"a":2,"b":3}}]}`,
			status: http.StatusBadRequest,
		},
	}

	get := func() (status int, respJSON string) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		sig, err := s.Client().signRequest(req)
		require.NoError(t, err)
		s.Client().setHeaders(req, sig)
		resp, err := s.Client().httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	post := func(body string) (status int, respJSON string) {
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		sig, err := s.Client().signRequest(req)
		require.NoError(t, err)
		s.Client().setHeaders(req, sig)
		resp, err := s.Client().httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			var gotStatus int
			var gotBody string
			switch step.op {
			case "get":
				gotStatus, gotBody = get()
			case "post":
				gotStatus, gotBody = post(step.req)
			default:
				t.Fatalf("unknown op: %s", step.op)
			}
			assert.Equal(t, step.status, gotStatus)
			assert.Equal(t, step.resp, gotBody)
		})
	}
}

func TestProfileData(t *testing.T) {
	suite.Run(t, &ProfileDataSuite{})
}
