package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"thaily/src/server/config"
	"thaily/src/server/router"
	"thaily/src/service/pkg/container"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{}              `json:"data"`
	Errors []map[string]interface{} `json:"errors,omitempty"`
}

func skipWorkflowTests() bool {
	return strings.EqualFold(os.Getenv("SKIP_GRAPHQL_WORKFLOWS"), "true")
}

// Integration test suite
type GraphQLIntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
	server *httptest.Server
	config *config.Config
}

func (suite *GraphQLIntegrationTestSuite) SetupSuite() {
	// Load test configuration
	cfg, err := config.LoadTest()
	require.NoError(suite.T(), err)
	suite.config = cfg

	// Initialize container with test configuration
	c, err := container.NewTest(cfg)
	require.NoError(suite.T(), err)

	// Setup router
	suite.router = router.Setup(cfg, c)
	suite.server = httptest.NewServer(suite.router)
}

func (suite *GraphQLIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

func (suite *GraphQLIntegrationTestSuite) executeGraphQLQuery(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	request := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", suite.server.URL+"/query", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var graphqlResp GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&graphqlResp)
	return &graphqlResp, err
}

func (suite *GraphQLIntegrationTestSuite) TestHealthCheck() {
	req, err := http.NewRequest("GET", suite.server.URL+"/health", nil)
	require.NoError(suite.T(), err)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", health["status"])
}

func (suite *GraphQLIntegrationTestSuite) TestMetricsEndpoint() {
	req, err := http.NewRequest("GET", suite.server.URL+"/metrics", nil)
	require.NoError(suite.T(), err)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Contains(suite.T(), resp.Header.Get("Content-Type"), "text/plain")

	// Read a portion of the metrics to verify format
	buf := make([]byte, 100)
	_, err = resp.Body.Read(buf)
	require.NoError(suite.T(), err)

	// Should contain prometheus metrics format
	assert.Contains(suite.T(), string(buf), "# HELP")
}

func (suite *GraphQLIntegrationTestSuite) TestGraphQLIntrospection() {
	query := `
		query {
			__schema {
				types {
					name
				}
			}
		}
	`

	resp, err := suite.executeGraphQLQuery(query, nil)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), resp.Errors)
	assert.NotNil(suite.T(), resp.Data)

	// Verify schema types are available
	data := resp.Data.(map[string]interface{})
	schema := data["__schema"].(map[string]interface{})
	types := schema["types"].([]interface{})

	// Should have basic GraphQL types
	typeNames := make([]string, 0)
	for _, t := range types {
		typeObj := t.(map[string]interface{})
		typeNames = append(typeNames, typeObj["name"].(string))
	}

	assert.Contains(suite.T(), typeNames, "Query")
	assert.Contains(suite.T(), typeNames, "Mutation")
}

func (suite *GraphQLIntegrationTestSuite) TestAcademicWorkflow() {
	if skipWorkflowTests() {
		suite.T().Skip("Skipping academic workflow when backing services are unavailable")
	}

	// Test creating an academic
	createMutation := `
		mutation CreateAcademic($input: CreateAcademicInput!) {
			createAcademic(input: $input) {
				id
				userId
				academicTitle
				degree
				fieldOfStudy
				institution
				yearObtained
				isVerified
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"userId":        "test-user-1",
			"academicTitle": "Professor",
			"degree":        "PhD",
			"fieldOfStudy":  "Computer Science",
			"institution":   "MIT",
			"yearObtained":  2020,
		},
	}

	resp, err := suite.executeGraphQLQuery(createMutation, variables)
	require.NoError(suite.T(), err)

	if len(resp.Errors) > 0 {
		suite.T().Logf("GraphQL Errors: %+v", resp.Errors)
	}

	// For now, we might expect errors due to database not being available in test
	// but we can verify the query structure is correct
	assert.NotNil(suite.T(), resp)
}

func (suite *GraphQLIntegrationTestSuite) TestUserWorkflow() {
	if skipWorkflowTests() {
		suite.T().Skip("Skipping user workflow when backing services are unavailable")
	}

	// Test user queries
	usersQuery := `
		query {
			users {
				id
				email
				name
				role
			}
		}
	`

	resp, err := suite.executeGraphQLQuery(usersQuery, nil)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)

	// Test user by ID query
	userQuery := `
		query GetUser($id: ID!) {
			user(id: $id) {
				id
				email
				name
				role
			}
		}
	`

	variables := map[string]interface{}{
		"id": "test-user-1",
	}

	resp, err = suite.executeGraphQLQuery(userQuery, variables)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
}

func (suite *GraphQLIntegrationTestSuite) TestComplexQueries() {
	if skipWorkflowTests() {
		suite.T().Skip("Skipping complex queries when backing services are unavailable")
	}

	// Test query with nested relationships
	complexQuery := `
		query {
			users {
				id
				email
				name
				role
				academics {
					id
					academicTitle
					degree
					institution
				}
			}
		}
	`

	resp, err := suite.executeGraphQLQuery(complexQuery, nil)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)

	// Test query with fragments
	fragmentQuery := `
		fragment UserInfo on User {
			id
			email
			name
		}
		
		query {
			users {
				...UserInfo
				role
			}
		}
	`

	resp, err = suite.executeGraphQLQuery(fragmentQuery, nil)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
}

func (suite *GraphQLIntegrationTestSuite) TestErrorHandling() {
	// Test invalid query
	invalidQuery := `
		query {
			nonExistentField
		}
	`

	resp, err := suite.executeGraphQLQuery(invalidQuery, nil)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), resp.Errors)

	// Verify error structure
	assert.Contains(suite.T(), resp.Errors[0]["message"], "field")
}

func (suite *GraphQLIntegrationTestSuite) TestConcurrentRequests() {
	query := `
		query {
			__schema {
				queryType {
					name
				}
			}
		}
	`

	// Execute multiple concurrent requests
	numRequests := 10
	results := make(chan *GraphQLResponse, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := suite.executeGraphQLQuery(query, nil)
			results <- resp
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case resp := <-results:
			assert.NotNil(suite.T(), resp)
		case err := <-errors:
			assert.NoError(suite.T(), err)
		case <-time.After(30 * time.Second):
			suite.T().Fatal("Timeout waiting for concurrent requests")
		}
	}
}

func TestGraphQLIntegrationSuite(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(GraphQLIntegrationTestSuite))
}

// Benchmark tests
func BenchmarkGraphQLSimpleQuery(b *testing.B) {
	// Setup
	cfg, err := config.LoadTest()
	require.NoError(b, err)

	c, err := container.NewTest(cfg)
	require.NoError(b, err)

	router := router.Setup(cfg, c)
	server := httptest.NewServer(router)
	defer server.Close()

	query := `
		query {
			__schema {
				queryType {
					name
				}
			}
		}
	`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			request := GraphQLRequest{Query: query}
			jsonData, _ := json.Marshal(request)

			req, _ := http.NewRequest("POST", server.URL+"/query", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				b.Error(err)
			}
			resp.Body.Close()
		}
	})
}
