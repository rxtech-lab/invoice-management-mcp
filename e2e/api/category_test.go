package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CategoryTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *CategoryTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *CategoryTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *CategoryTestSuite) TestCreateCategory() {
	category := map[string]interface{}{
		"name":        "Office Supplies",
		"description": "Supplies for the office",
		"color":       "#3498db",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/categories", category)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Office Supplies", result["name"])
	s.Equal("Supplies for the office", result["description"])
	s.Equal("#3498db", result["color"])
	s.NotNil(result["id"])
}

func (s *CategoryTestSuite) TestCreateCategoryMissingName() {
	category := map[string]interface{}{
		"description": "Test description",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/categories", category)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *CategoryTestSuite) TestListCategories() {
	// Create some categories
	_, err := s.setup.CreateTestCategory("Category 1")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestCategory("Category 2")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/categories", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["data"])
	data := result["data"].([]interface{})
	s.GreaterOrEqual(len(data), 2)
}

func (s *CategoryTestSuite) TestListCategoriesWithKeyword() {
	_, err := s.setup.CreateTestCategory("Office Supplies")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestCategory("Travel Expenses")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/categories?keyword=office", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Equal(1, len(data))
}

func (s *CategoryTestSuite) TestGetCategory() {
	categoryID, err := s.setup.CreateTestCategory("Test Category")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/categories/"+uintToString(categoryID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test Category", result["name"])
}

func (s *CategoryTestSuite) TestGetCategoryNotFound() {
	resp, err := s.setup.MakeRequest("GET", "/api/categories/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *CategoryTestSuite) TestUpdateCategory() {
	categoryID, err := s.setup.CreateTestCategory("Original Name")
	s.Require().NoError(err)

	update := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
	}

	resp, err := s.setup.MakeRequest("PUT", "/api/categories/"+uintToString(categoryID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Name", result["name"])
	s.Equal("Updated description", result["description"])
}

func (s *CategoryTestSuite) TestDeleteCategory() {
	categoryID, err := s.setup.CreateTestCategory("To Delete")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", "/api/categories/"+uintToString(categoryID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify deletion
	resp, err = s.setup.MakeRequest("GET", "/api/categories/"+uintToString(categoryID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestCategorySuite(t *testing.T) {
	suite.Run(t, new(CategoryTestSuite))
}

// Helper function
func uintToString(n uint) string {
	return fmt.Sprintf("%d", n)
}
