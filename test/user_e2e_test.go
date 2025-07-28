package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epsierra/phinex-blog-api/src/app"
	"github.com/epsierra/phinex-blog-api/src/auth"
	"github.com/epsierra/phinex-blog-api/src/database"
	"github.com/epsierra/phinex-blog-api/src/models"
	"github.com/epsierra/phinex-blog-api/src/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func init() {
	log.Println("Running Test")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

type UserControllerSuite struct {
	suite.Suite
	app       *fiber.App
	db        *gorm.DB
	testUser  *models.User
	authToken string
}

func TestUserController(t *testing.T) {
	suite.Run(t, &UserControllerSuite{})
}

func (ucSuite *UserControllerSuite) SetupSuite() {
	// Initialize database connection
	db, err := database.NewDatabaseConnection()
	if err != nil {
		ucSuite.FailNowf("Database Error", "%v", err.Error())
	}
	ucSuite.db = db
	app := app.AppSetup(db)
	ucSuite.app = app

	// Create a test user
	testUser := models.User{
		UserId:    utils.GenerateID(),
		Email:     "test_user@example.com",
		Password:  "password123", // In a real app, hash this
		FullName:  "Test User 2",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.Create(&testUser)
	ucSuite.testUser = &testUser

	// Assign a role to the test user
	testRole := models.Role{
		RoleId:    utils.GenerateID(),
		RoleName:  models.RoleNameAuthenticated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.FirstOrCreate(&testRole, models.Role{RoleName: models.RoleNameAuthenticated})

	userRole := models.UserRole{
		UserRoleId: utils.GenerateID(),
		UserId:     testUser.UserId,
		RoleId:     testRole.RoleId,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CreatedBy:  "test",
		UpdatedBy:  "test",
	}
	ucSuite.db.Create(&userRole)

	// Get an auth token for the test user
	authService := auth.NewAuthService(db)
	tokenResponse, err := authService.GetTokenByEmail(testUser.Email)
	if err != nil {
		ucSuite.FailNowf("Failed to get auth token", "%v", err.Error())
	}
	ucSuite.authToken = tokenResponse.Token
}

func (ucSuite *UserControllerSuite) TearDownSuite() {
	// Clean up test data
	if ucSuite.db != nil {
		// Delete the test user's roles
		ucSuite.db.Where("user_id = ?", ucSuite.testUser.UserId).Delete(&models.UserRole{})
		// Delete the test user
		ucSuite.db.Delete(ucSuite.testUser)
		// Delete the test role if it was created by the test setup
		ucSuite.db.Where("role_name = ?", models.RoleNameAuthenticated).Delete(&models.Role{})
	}
}

func (ucSuite *UserControllerSuite) TestCreateUser() {
	assert := ucSuite.Assert()

	// Define the request body for creating a user
	userPayload := map[string]interface{}{
		"email":    "newuser@example.com",
		"password": "newpassword123",
		"fullName": "New User",
	}
	jsonPayload, _ := json.Marshal(userPayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusCreated, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)

	// Assert the response message and user details
	assert.Equal("User created successfully", responseBody["message"])
	userData := responseBody["user"].(map[string]interface{})
	assert.Equal(userPayload["email"], userData["email"])
	assert.Equal(userPayload["fullName"], userData["fullName"])

	// Clean up the created user
	var createdUser models.User
	ucSuite.db.Where("email = ?", userPayload["email"]).First(&createdUser)
	ucSuite.db.Delete(&createdUser)
}

func (ucSuite *UserControllerSuite) TestFindAllUsers() {
	assert := ucSuite.Assert()

	// Seed some users
	seededUsers := []models.User{
		{
			UserId:    utils.GenerateID(),
			Email:     "seeded1@example.com",
			Password:  "password",
			FullName:  "Seeded User 1",
			Verified:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "test",
			UpdatedBy: "test",
		},
		{
			UserId:    utils.GenerateID(),
			Email:     "seeded2@example.com",
			Password:  "password",
			FullName:  "Seeded User 2",
			Verified:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "test",
			UpdatedBy: "test",
		},
	}
	ucSuite.db.Create(&seededUsers)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	// We expect at least 3 users: the test user and the two seeded users
	assert.GreaterOrEqual(len(paginatedResponse.Data.([]interface{})), 3)
	assert.GreaterOrEqual(paginatedResponse.Metadata.TotalItems, int64(3))

	// Clean up seeded users
	ucSuite.db.Delete(&seededUsers)
}

func (ucSuite *UserControllerSuite) TestFindUserById() {
	assert := ucSuite.Assert()

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", ucSuite.testUser.UserId), nil)
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var userResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	assert.NoError(err)

	// Assert the user details
	assert.Equal(ucSuite.testUser.Email, userResponse["email"])
	assert.Equal(ucSuite.testUser.FullName, userResponse["fullName"])
}

func (ucSuite *UserControllerSuite) TestUpdateUser() {
	assert := ucSuite.Assert()

	// Define the update payload
	updatePayload := map[string]interface{}{
		"fullName": "Updated Test User",
		"bio":      "This is an updated bio.",
	}
	jsonPayload, _ := json.Marshal(updatePayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", ucSuite.testUser.UserId), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("User updated successfully", responseBody["message"])

	// Verify the user was updated in the database
	var updatedUser models.User
	ucSuite.db.Where("user_id = ?", ucSuite.testUser.UserId).First(&updatedUser)
	assert.Equal(updatePayload["fullName"], updatedUser.FullName)
	assert.Equal(updatePayload["bio"], updatedUser.Bio)
}

func (ucSuite *UserControllerSuite) TestDeleteUser() {
	assert := ucSuite.Assert()

	// Create a temporary user to delete
	userToDelete := models.User{
		UserId:    utils.GenerateID(),
		Email:     "delete@example.com",
		Password:  "password",
		FullName:  "User To Delete",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.Create(&userToDelete)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", userToDelete.UserId), nil)
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("User deleted successfully", responseBody["message"])

	// Verify the user was deleted from the database
	var deletedUser models.User
	err = ucSuite.db.Where("user_id = ?", userToDelete.UserId).First(&deletedUser).Error
	assert.Error(err) // Should return an error (record not found)
	assert.Equal(gorm.ErrRecordNotFound, err)
}

func (ucSuite *UserControllerSuite) TestFollowUnfollowUser() {
	assert := ucSuite.Assert()

	// Create a user to follow/unfollow
	userToInteract := models.User{
		UserId:    utils.GenerateID(),
		Email:     "interact@example.com",
		Password:  "password",
		FullName:  "User To Interact",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.Create(&userToInteract)

	// --- Test Follow ---
	followPayload := map[string]interface{}{
		"followerId":  ucSuite.testUser.UserId,
		"followingId": userToInteract.UserId,
	}
	jsonPayload, _ := json.Marshal(followPayload)

	req := httptest.NewRequest(http.MethodPost, "/users/follows", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	assert.Equal(http.StatusOK, resp.StatusCode)
	var followResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&followResponse)
	assert.NoError(err)
	assert.Equal("User followed successfully", followResponse["message"])

	// Verify follow in DB
	var follow models.Follow
	err = ucSuite.db.Where("follower_id = ? AND following_id = ?", ucSuite.testUser.UserId, userToInteract.UserId).First(&follow).Error
	assert.NoError(err)

	// --- Test Unfollow ---
	unfollowPayload := map[string]interface{}{
		"followerId":  ucSuite.testUser.UserId,
		"followingId": userToInteract.UserId,
	}
	jsonPayload, _ = json.Marshal(unfollowPayload)

	req = httptest.NewRequest(http.MethodPost, "/users/follows", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	resp, err = ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	assert.Equal(http.StatusOK, resp.StatusCode)
	var unfollowResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&unfollowResponse)
	assert.NoError(err)
	assert.Equal("User unfollowed successfully", unfollowResponse["message"])

	// Verify unfollow in DB
	err = ucSuite.db.Where("follower_id = ? AND following_id = ?", ucSuite.testUser.UserId, userToInteract.UserId).First(&follow).Error
	assert.Error(err) // Should be not found
	assert.Equal(gorm.ErrRecordNotFound, err)

	// Clean up
	ucSuite.db.Delete(&userToInteract)
}

func (ucSuite *UserControllerSuite) TestFindUsersNotFollowing() {
	assert := ucSuite.Assert()

	// Create some users
	users := []models.User{
		{
			UserId:    utils.GenerateID(),
			Email:     "notfollowing1@example.com",
			Password:  "password",
			FullName:  "Not Following User 1",
			Verified:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "test",
			UpdatedBy: "test",
		},
		{
			UserId:    utils.GenerateID(),
			Email:     "notfollowing2@example.com",
			Password:  "password",
			FullName:  "Not Following User 2",
			Verified:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "test",
			UpdatedBy: "test",
		},
	}
	ucSuite.db.Create(&users)

	// Make testUser follow one of them
	follow := models.Follow{
		FollowId:    utils.GenerateID(),
		FollowerId:  ucSuite.testUser.UserId,
		FollowingId: users[0].UserId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   "test",
		UpdatedBy:   "test",
	}
	ucSuite.db.Create(&follow)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/unfollowings", ucSuite.testUser.UserId), nil)
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert that only 'not-following-user-2' is returned (or more if other users exist in DB)
	found := false
	for _, item := range paginatedResponse.Data.([]interface{}) {
		user := item.(map[string]interface{})
		if user["userId"] == users[1].UserId {
			found = true
			break
		}
	}
	assert.True(found)

	// Clean up
	ucSuite.db.Delete(&users)
	ucSuite.db.Delete(&follow)
}

func (ucSuite *UserControllerSuite) TestFindUserFollowers() {
	assert := ucSuite.Assert()

	// Create a user who will follow testUser
	followerUser := models.User{
		UserId:    utils.GenerateID(),
		Email:     "follower@example.com",
		Password:  "password",
		FullName:  "Follower User",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.Create(&followerUser)

	// Create a follow relationship
	follow := models.Follow{
		FollowId:    utils.GenerateID(),
		FollowerId:  followerUser.UserId,
		FollowingId: ucSuite.testUser.UserId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   "test",
		UpdatedBy:   "test",
	}
	ucSuite.db.Create(&follow)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/followers", ucSuite.testUser.UserId), nil)
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 1)
	assert.Equal(int64(1), paginatedResponse.Metadata.TotalItems)
	assert.Equal(followerUser.Email, paginatedResponse.Data.([]interface{})[0].(map[string]interface{})["email"])

	// Clean up
	ucSuite.db.Delete(&followerUser)
	ucSuite.db.Delete(&follow)
}

func (ucSuite *UserControllerSuite) TestFindUserFollowings() {
	assert := ucSuite.Assert()

	// Create a user that testUser will follow
	followingUser := models.User{
		UserId:    utils.GenerateID(),
		Email:     "following@example.com",
		Password:  "password",
		FullName:  "Following User",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.Create(&followingUser)

	// Create a follow relationship
	follow := models.Follow{
		FollowId:    utils.GenerateID(),
		FollowerId:  ucSuite.testUser.UserId,
		FollowingId: followingUser.UserId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(), CreatedBy: "test",
		UpdatedBy: "test",
	}
	ucSuite.db.Create(&follow)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/followings", ucSuite.testUser.UserId), nil)
	req.Header.Set("Authorization", "Bearer "+ucSuite.authToken)

	// Perform the request
	resp, err := ucSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 1)
	assert.Equal(int64(1), paginatedResponse.Metadata.TotalItems)
	assert.Equal(followingUser.Email, paginatedResponse.Data.([]interface{})[0].(map[string]interface{})["email"])

	// Clean up
	ucSuite.db.Delete(&followingUser)
	ucSuite.db.Delete(&follow)
}
