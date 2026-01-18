package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UploadTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *UploadTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *UploadTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *UploadTestSuite) TestUploadFile() {
	// Create a multipart form with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a fake file
	part, err := writer.CreateFormFile("file", "test-invoice.pdf")
	s.Require().NoError(err)

	// Write some content to simulate a PDF
	_, err = part.Write([]byte("%PDF-1.4 fake pdf content"))
	s.Require().NoError(err)

	err = writer.Close()
	s.Require().NoError(err)

	req := httptest.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Test-User-ID", s.setup.TestUserID)

	resp, err := s.setup.App.Test(req, -1)
	s.Require().NoError(err)

	// Note: The mock upload service returns a successful response
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["download_url"])
	s.NotNil(result["filename"])
	s.NotNil(result["key"])
}

func (s *UploadTestSuite) TestUploadFileNoFile() {
	// Create an empty multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err := writer.Close()
	s.Require().NoError(err)

	req := httptest.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Test-User-ID", s.setup.TestUserID)

	resp, err := s.setup.App.Test(req, -1)
	s.Require().NoError(err)

	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *UploadTestSuite) TestGetPresignedURL() {
	// Use GET with query parameters as per handler implementation
	req := httptest.NewRequest("GET", "/api/upload/presigned?filename=invoice-2024.pdf&content_type=application/pdf", nil)
	req.Header.Set("X-Test-User-ID", s.setup.TestUserID)

	resp, err := s.setup.App.Test(req, -1)
	s.Require().NoError(err)

	// Note: The mock upload service returns a successful response
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["upload_url"])
	s.NotNil(result["key"])
}

func (s *UploadTestSuite) TestGetPresignedURLMissingFilename() {
	// Use GET with query parameters - missing filename
	req := httptest.NewRequest("GET", "/api/upload/presigned?content_type=application/pdf", nil)
	req.Header.Set("X-Test-User-ID", s.setup.TestUserID)

	resp, err := s.setup.App.Test(req, -1)
	s.Require().NoError(err)

	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestUploadSuite(t *testing.T) {
	suite.Run(t, new(UploadTestSuite))
}
