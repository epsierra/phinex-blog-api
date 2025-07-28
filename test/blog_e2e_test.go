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

type BlogControllerSuite struct {
	suite.Suite
	app       *fiber.App
	db        *gorm.DB
	testUser  *models.User
	authToken string
}

func TestBlogController(t *testing.T) {
	suite.Run(t, &BlogControllerSuite{})
}

func (bcSuite *BlogControllerSuite) SetupSuite() {
	// Initialize database connection
	db, err := database.NewDatabaseConnection()
	if err != nil {
		bcSuite.FailNowf("Database Error", "%v", err.Error())
	}
	bcSuite.db = db
	app := app.AppSetup(db)
	bcSuite.app = app

	// Create a test user
	testUser := models.User{
		UserId:    utils.GenerateID(),
		Email:     "test@example.com",
		Password:  "password123", // In a real app, hash this
		FullName:  "Test User",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	bcSuite.db.Create(&testUser)
	bcSuite.testUser = &testUser

	// Assign a role to the test user
	testRole := models.Role{
		RoleId:    utils.GenerateID(),
		RoleName:  models.RoleNameAuthenticated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	bcSuite.db.FirstOrCreate(&testRole, models.Role{RoleName: models.RoleNameAuthenticated})

	userRole := models.UserRole{
		UserRoleId: utils.GenerateID(),
		UserId:     testUser.UserId,
		RoleId:     testRole.RoleId,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CreatedBy:  "test",
		UpdatedBy:  "test",
	}
	bcSuite.db.Create(&userRole)

	// Get an auth token for the test user
	authService := auth.NewAuthService(db)
	tokenResponse, err := authService.GetTokenByEmail(testUser.Email)
	if err != nil {
		bcSuite.FailNowf("Failed to get auth token", "%v", err.Error())
	}
	bcSuite.authToken = tokenResponse.Token
}

func (bcSuite *BlogControllerSuite) TearDownSuite() {
	// Clean up test data
	if bcSuite.db != nil {
		// Delete blogs created by the test user
		bcSuite.db.Where("user_id = ?", bcSuite.testUser.UserId).Delete(&models.Blog{})
		// Delete the test user's roles
		bcSuite.db.Where("user_id = ?", bcSuite.testUser.UserId).Delete(&models.UserRole{})
		// Delete the test user
		bcSuite.db.Delete(bcSuite.testUser)
		// Delete the test role if it was created by the test setup
		bcSuite.db.Where("role_name = ?", models.RoleNameAuthenticated).Delete(&models.Role{})
	}
}

func (bcSuite *BlogControllerSuite) TestCreateBlog() {
	assert := bcSuite.Assert()

	// Define the request body for creating a blog
	blogPayload := map[string]interface{}{
		"title": "Test Blog Title",
		"text":  "This is the content of the test blog.",
		"url":   "http://example.com/test-blog",
	}
	jsonPayload, _ := json.Marshal(blogPayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPost, "/blogs", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken) // Use the generated token

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusCreated, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)

	// Assert the response message
	assert.Equal("Blog created successfully", responseBody["message"])

	// Optionally, verify the blog was created in the database
	var createdBlog models.Blog
	err = bcSuite.db.Where("title = ?", blogPayload["title"]).First(&createdBlog).Error
	assert.NoError(err)
	assert.Equal(blogPayload["text"], createdBlog.Text)
	assert.Equal(bcSuite.testUser.UserId, createdBlog.UserId)
}

func (bcSuite *BlogControllerSuite) TestFindAllBlogs() {
	assert := bcSuite.Assert()

	// Seed some blogs for the test user
	seededBlogs := []models.Blog{
		{
			BlogId:    utils.GenerateID(),
			UserId:    bcSuite.testUser.UserId,
			Title:     "Blog 1",
			Text:      "Content 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
		{
			BlogId:    utils.GenerateID(),
			UserId:    bcSuite.testUser.UserId,
			Title:     "Blog 2",
			Text:      "Content 2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
	}
	bcSuite.db.Create(&seededBlogs)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/blogs", nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 2) // Assuming 2 blogs were seeded
	assert.Equal(int64(2), paginatedResponse.Metadata.TotalItems)
	assert.Equal(int64(1), paginatedResponse.Metadata.CurrentPage)
	assert.Equal(int64(10), paginatedResponse.Metadata.ItemsPerPage)

	// Clean up seeded blogs
	bcSuite.db.Delete(&seededBlogs)
}

func (bcSuite *BlogControllerSuite) TestFindBlogById() {
	assert := bcSuite.Assert()

	// Seed a blog for the test user
	seededBlog := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Blog to Find",
		Text:      "Content of blog to find",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&seededBlog)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/blogs/%s", seededBlog.BlogId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var blogResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&blogResponse)
	assert.NoError(err)

	// Assert the blog details
	assert.Equal(seededBlog.Title, blogResponse["title"])
	assert.Equal(seededBlog.Text, blogResponse["text"])

	// Clean up seeded blog
	bcSuite.db.Delete(&seededBlog)
}

func (bcSuite *BlogControllerSuite) TestUpdateBlog() {
	assert := bcSuite.Assert()

	// Seed a blog to update
	blogToUpdate := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Original Title",
		Text:      "Original Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&blogToUpdate)

	// Define the update payload
	updatePayload := map[string]interface{}{
		"title": "Updated Title",
		"text":  "Updated Content",
	}
	jsonPayload, _ := json.Marshal(updatePayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/blogs/%s", blogToUpdate.BlogId), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusAccepted, resp.StatusCode) // 202 Accepted

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("Blog updated successfully", responseBody["message"])

	// Verify the blog was updated in the database
	var updatedBlog models.Blog
	err = bcSuite.db.Where("blog_id = ?", blogToUpdate.BlogId).First(&updatedBlog).Error
	assert.NoError(err)
	assert.Equal(updatePayload["title"], updatedBlog.Title)
	assert.Equal(updatePayload["text"], updatedBlog.Text)

	// Clean up seeded blog
	bcSuite.db.Delete(&blogToUpdate)
}

func (bcSuite *BlogControllerSuite) TestDeleteBlog() {
	assert := bcSuite.Assert()

	// Seed a blog to delete
	blogToDelete := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Blog to Delete",
		Text:      "Content of blog to delete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&blogToDelete)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/blogs/%s", blogToDelete.BlogId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("Blog deleted successfully", responseBody["message"])

	// Verify the blog was deleted from the database
	var deletedBlog models.Blog
	err = bcSuite.db.Where("blog_id = ?", blogToDelete.BlogId).First(&deletedBlog).Error
	assert.Error(err) // Should return an error (record not found)
	assert.Equal(gorm.ErrRecordNotFound, err)
}

func (bcSuite *BlogControllerSuite) TestLikeBlog() {
	assert := bcSuite.Assert()

	// Seed a blog to like
	blogToLike := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Blog to Like",
		Text:      "Content of blog to like",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&blogToLike)

	// Create a new HTTP request to like the blog
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/blogs/%s/likes", blogToLike.BlogId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal(true, responseBody["liked"])

	// Verify the like was recorded in the database
	var like models.Like
	err = bcSuite.db.Where("ref_id = ? AND user_id = ?", blogToLike.BlogId, bcSuite.testUser.UserId).First(&like).Error
	assert.NoError(err)

	// Now, unlike the blog
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/blogs/%s/likes", blogToLike.BlogId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	resp, err = bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	assert.Equal(http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal(false, responseBody["liked"])

	// Verify the like was removed from the database
	err = bcSuite.db.Where("ref_id = ? AND user_id = ?", blogToLike.BlogId, bcSuite.testUser.UserId).First(&like).Error
	assert.Error(err) // Should return an error (record not found)
	assert.Equal(gorm.ErrRecordNotFound, err)

	// Clean up seeded blog
	bcSuite.db.Delete(&blogToLike)
}

func (bcSuite *BlogControllerSuite) TestAddComment() {
	assert := bcSuite.Assert()

	// Seed a blog to comment on
	blogToComment := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Blog to Comment",
		Text:      "Content of blog to comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&blogToComment)

	// Define the comment payload
	commentPayload := map[string]interface{}{
		"text": "This is a test comment.",
	}
	jsonPayload, _ := json.Marshal(commentPayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/blogs/%s/comments", blogToComment.BlogId), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusCreated, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("Comment added successfully", responseBody["message"])

	// Verify the comment was added to the database
	var comment models.Comment
	err = bcSuite.db.Where("ref_id = ? AND user_id = ?", blogToComment.BlogId, bcSuite.testUser.UserId).First(&comment).Error
	assert.NoError(err)
	assert.Equal(commentPayload["text"], comment.Text)

	// Clean up seeded blog and comment
	bcSuite.db.Delete(&blogToComment)
	bcSuite.db.Delete(&comment)
}

func (bcSuite *BlogControllerSuite) TestFindComments() {
	assert := bcSuite.Assert()

	// Seed a blog
	blogWithComments := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Blog with Comments",
		Text:      "Content of blog with comments",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&blogWithComments)

	// Seed some comments for the blog
	seededComments := []models.Comment{
		{
			CommentId: utils.GenerateID(),
			RefId:     blogWithComments.BlogId,
			UserId:    bcSuite.testUser.UserId,
			Text:      "First comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
		{
			CommentId: utils.GenerateID(),
			RefId:     blogWithComments.BlogId,
			UserId:    bcSuite.testUser.UserId,
			Text:      "Second comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
	}
	bcSuite.db.Create(&seededComments)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/blogs/%s/comments", blogWithComments.BlogId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 2)
	assert.Equal(int64(2), paginatedResponse.Metadata.TotalItems)

	// Clean up seeded data
	bcSuite.db.Delete(&blogWithComments)
	bcSuite.db.Delete(&seededComments)
}

func (bcSuite *BlogControllerSuite) TestUpdateComment() {
	assert := bcSuite.Assert()

	// Seed a comment to update
	commentToUpdate := models.Comment{
		CommentId: utils.GenerateID(),
		RefId:     "some-blog-id", // This can be a dummy ID as we are testing comment update directly
		UserId:    bcSuite.testUser.UserId,
		Text:      "Original comment text",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&commentToUpdate)

	// Define the update payload
	updatePayload := map[string]interface{}{
		"text": "Updated comment text",
	}
	jsonPayload, _ := json.Marshal(updatePayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s", commentToUpdate.CommentId), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusAccepted, resp.StatusCode) // 202 Accepted

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("Comment updated successfully", responseBody["message"])

	// Verify the comment was updated in the database
	var updatedComment models.Comment
	err = bcSuite.db.Where("comment_id = ?", commentToUpdate.CommentId).First(&updatedComment).Error
	assert.NoError(err)
	assert.Equal(updatePayload["text"], updatedComment.Text)

	// Clean up seeded comment
	bcSuite.db.Delete(&commentToUpdate)
}

func (bcSuite *BlogControllerSuite) TestDeleteComment() {
	assert := bcSuite.Assert()

	// Seed a comment to delete
	commentToDelete := models.Comment{
		CommentId: utils.GenerateID(),
		RefId:     "some-blog-id", // This can be a dummy ID
		UserId:    bcSuite.testUser.UserId,
		Text:      "Comment to delete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&commentToDelete)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/comments/%s", commentToDelete.CommentId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("Comment deleted successfully", responseBody["message"])

	// Verify the comment was deleted from the database
	var deletedComment models.Comment
	err = bcSuite.db.Where("comment_id = ?", commentToDelete.CommentId).First(&deletedComment).Error
	assert.Error(err) // Should return an error (record not found)
	assert.Equal(gorm.ErrRecordNotFound, err)
}

func (bcSuite *BlogControllerSuite) TestLikeComment() {
	assert := bcSuite.Assert()

	// Seed a comment to like
	commentToLike := models.Comment{
		CommentId: utils.GenerateID(),
		RefId:     "some-blog-id", // Dummy ID
		UserId:    bcSuite.testUser.UserId,
		Text:      "Comment to like",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&commentToLike)

	// Create a new HTTP request to like the comment
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s/likes", commentToLike.CommentId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal(true, responseBody["liked"])

	// Verify the like was recorded in the database
	var like models.Like
	err = bcSuite.db.Where("ref_id = ? AND user_id = ?", commentToLike.CommentId, bcSuite.testUser.UserId).First(&like).Error
	assert.NoError(err)

	// Now, unlike the comment
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/comments/%s/likes", commentToLike.CommentId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	resp, err = bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	assert.Equal(http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal(false, responseBody["liked"])

	// Verify the like was removed from the database
	err = bcSuite.db.Where("ref_id = ? AND user_id = ?", commentToLike.CommentId, bcSuite.testUser.UserId).First(&like).Error
	assert.Error(err) // Should return an error (record not found)
	assert.Equal(gorm.ErrRecordNotFound, err)

	// Clean up seeded comment
	bcSuite.db.Delete(&commentToLike)
}

func (bcSuite *BlogControllerSuite) TestAddReply() {
	assert := bcSuite.Assert()

	// Seed a comment to reply to
	commentToReply := models.Comment{
		CommentId: utils.GenerateID(),
		RefId:     "some-blog-id", // Dummy ID
		UserId:    bcSuite.testUser.UserId,
		Text:      "Comment to reply to",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&commentToReply)

	// Define the reply payload
	replyPayload := map[string]interface{}{
		"text": "This is a test reply.",
	}
	jsonPayload, _ := json.Marshal(replyPayload)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/comments/%s/replies", commentToReply.CommentId), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusCreated, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)
	assert.Equal("Reply added successfully", responseBody["message"])

	// Verify the reply was added to the database (replies are also comments with refId as parent commentId)
	var reply models.Comment
	err = bcSuite.db.Where("ref_id = ? AND user_id = ?", commentToReply.CommentId, bcSuite.testUser.UserId).First(&reply).Error
	assert.NoError(err)
	assert.Equal(replyPayload["text"], reply.Text)

	// Clean up seeded comment and reply
	bcSuite.db.Delete(&commentToReply)
	bcSuite.db.Delete(&reply)
}

func (bcSuite *BlogControllerSuite) TestFindReplies() {
	assert := bcSuite.Assert()

	// Seed a comment
	parentComment := models.Comment{
		CommentId: utils.GenerateID(),
		RefId:     "some-blog-id", // Dummy ID
		UserId:    bcSuite.testUser.UserId,
		Text:      "Parent comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&parentComment)

	// Seed some replies for the comment
	seededReplies := []models.Comment{
		{
			CommentId: utils.GenerateID(),
			RefId:     parentComment.CommentId,
			UserId:    bcSuite.testUser.UserId,
			Text:      "First reply",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
		{
			CommentId: utils.GenerateID(),
			RefId:     parentComment.CommentId,
			UserId:    bcSuite.testUser.UserId,
			Text:      "Second reply",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
	}
	bcSuite.db.Create(&seededReplies)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/comments/%s/replies", parentComment.CommentId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 2)
	assert.Equal(int64(2), paginatedResponse.Metadata.TotalItems)

	// Clean up seeded data
	bcSuite.db.Delete(&parentComment)
	bcSuite.db.Delete(&seededReplies)
}

func (bcSuite *BlogControllerSuite) TestFindUserBlogs() {
	assert := bcSuite.Assert()

	// Seed some blogs for the test user
	seededBlogs := []models.Blog{
		{
			BlogId:    utils.GenerateID(),
			UserId:    bcSuite.testUser.UserId,
			Title:     "User Blog 1",
			Text:      "Content of user blog 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
		{
			BlogId:    utils.GenerateID(),
			UserId:    bcSuite.testUser.UserId,
			Title:     "User Blog 2",
			Text:      "Content of user blog 2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
	}
	bcSuite.db.Create(&seededBlogs)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/blogs", bcSuite.testUser.UserId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 2)
	assert.Equal(int64(2), paginatedResponse.Metadata.TotalItems)

	// Clean up seeded blogs
	bcSuite.db.Delete(&seededBlogs)
}

func (bcSuite *BlogControllerSuite) TestFindSessionBlogs() {
	assert := bcSuite.Assert()

	// Seed some blogs for the test user
	seededBlogs := []models.Blog{
		{
			BlogId:    utils.GenerateID(),
			UserId:    bcSuite.testUser.UserId,
			Title:     "Session Blog 1",
			Text:      "Content of session blog 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
		{
			BlogId:    utils.GenerateID(),
			UserId:    bcSuite.testUser.UserId,
			Title:     "Session Blog 2",
			Text:      "Content of session blog 2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: bcSuite.testUser.FullName,
			UpdatedBy: bcSuite.testUser.FullName,
		},
	}
	bcSuite.db.Create(&seededBlogs)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/sessions/blogs", nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var paginatedResponse models.PaginatedResponse
	err = json.NewDecoder(resp.Body).Decode(&paginatedResponse)
	assert.NoError(err)

	// Assert the data and metadata
	assert.Len(paginatedResponse.Data.([]interface{}), 2)
	assert.Equal(int64(2), paginatedResponse.Metadata.TotalItems)

	// Clean up seeded blogs
	bcSuite.db.Delete(&seededBlogs)
}

func (bcSuite *BlogControllerSuite) TestFindLikesAndFollowers() {
	assert := bcSuite.Assert()

	// Seed a blog
	blog := models.Blog{
		BlogId:    utils.GenerateID(),
		UserId:    bcSuite.testUser.UserId,
		Title:     "Blog for Likes and Followers",
		Text:      "Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&blog)

	// Seed a like for the blog
	like := models.Like{
		LikeId:    utils.GenerateID(),
		RefId:     blog.BlogId,
		UserId:    bcSuite.testUser.UserId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: bcSuite.testUser.FullName,
		UpdatedBy: bcSuite.testUser.FullName,
	}
	bcSuite.db.Create(&like)

	// Seed another user and a follow relationship
	otherUser := models.User{
		UserId:    "other-user-id",
		Email:     "other@example.com",
		Password:  "password",
		FullName:  "Other User",
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: "test",
		UpdatedBy: "test",
	}
	bcSuite.db.Create(&otherUser)

	follow := models.Follow{
		FollowId:    "follow-1",
		FollowerId:  otherUser.UserId,
		FollowingId: bcSuite.testUser.UserId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   "test",
		UpdatedBy:   "test",
	}
	bcSuite.db.Create(&follow)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/blogs/%s/follows/likes", blog.BlogId), nil)
	req.Header.Set("Authorization", "Bearer "+bcSuite.authToken)

	// Perform the request
	resp, err := bcSuite.app.Test(req, -1)
	assert.NoError(err)
	defer resp.Body.Close()

	// Assert the response status code
	assert.Equal(http.StatusOK, resp.StatusCode)

	// Decode the response body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(err)

	// Assert the content
	assert.Contains(responseBody, "likes")
	assert.Contains(responseBody, "followers")

	likes := responseBody["likes"].([]interface{})
	followers := responseBody["followers"].([]interface{})

	assert.Len(likes, 1)
	assert.Len(followers, 1)

	// Clean up seeded data
	bcSuite.db.Delete(&blog)
	bcSuite.db.Delete(&like)
	bcSuite.db.Delete(&otherUser)
	bcSuite.db.Delete(&follow)
}
