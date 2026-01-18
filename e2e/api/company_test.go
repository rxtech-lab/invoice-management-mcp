package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CompanyTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *CompanyTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *CompanyTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *CompanyTestSuite) TestCreateCompany() {
	company := map[string]interface{}{
		"name":    "Acme Corporation",
		"address": "123 Main Street, City, Country",
		"email":   "contact@acme.com",
		"phone":   "+1-555-123-4567",
		"website": "https://acme.com",
		"tax_id":  "US123456789",
		"notes":   "Primary vendor",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/companies", company)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Acme Corporation", result["name"])
	s.Equal("123 Main Street, City, Country", result["address"])
	s.Equal("contact@acme.com", result["email"])
	s.Equal("+1-555-123-4567", result["phone"])
	s.Equal("https://acme.com", result["website"])
	s.Equal("US123456789", result["tax_id"])
	s.NotNil(result["id"])
}

func (s *CompanyTestSuite) TestCreateCompanyMissingName() {
	company := map[string]interface{}{
		"address": "Some address",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/companies", company)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *CompanyTestSuite) TestListCompanies() {
	// Create some companies
	_, err := s.setup.CreateTestCompany("Company A")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestCompany("Company B")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/companies", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["data"])
	data := result["data"].([]interface{})
	s.GreaterOrEqual(len(data), 2)
}

func (s *CompanyTestSuite) TestListCompaniesWithKeyword() {
	_, err := s.setup.CreateTestCompany("Acme Corporation")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestCompany("Beta Industries")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/companies?keyword=acme", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Equal(1, len(data))
}

func (s *CompanyTestSuite) TestGetCompany() {
	companyID, err := s.setup.CreateTestCompany("Test Company")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/companies/"+uintToString(companyID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test Company", result["name"])
}

func (s *CompanyTestSuite) TestGetCompanyNotFound() {
	resp, err := s.setup.MakeRequest("GET", "/api/companies/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *CompanyTestSuite) TestUpdateCompany() {
	companyID, err := s.setup.CreateTestCompany("Original Company")
	s.Require().NoError(err)

	update := map[string]interface{}{
		"name":    "Updated Company",
		"address": "New Address",
		"email":   "new@company.com",
	}

	resp, err := s.setup.MakeRequest("PUT", "/api/companies/"+uintToString(companyID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Company", result["name"])
	s.Equal("New Address", result["address"])
	s.Equal("new@company.com", result["email"])
}

func (s *CompanyTestSuite) TestDeleteCompany() {
	companyID, err := s.setup.CreateTestCompany("To Delete")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", "/api/companies/"+uintToString(companyID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify deletion
	resp, err = s.setup.MakeRequest("GET", "/api/companies/"+uintToString(companyID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestCompanySuite(t *testing.T) {
	suite.Run(t, new(CompanyTestSuite))
}
