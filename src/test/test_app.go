package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/epsierra/phinex-blog-api/src/app"
	"github.com/epsierra/phinex-blog-api/src/database"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
)

type BlogControllerSuite struct {
	suite.Suite
	app *fiber.App
}

func TestBlogController(t *testing.T) {
	suite.Run(t, &BlogControllerSuite{})
}

func (bcSuite *BlogControllerSuite) SetupSuite() {
	// Initialize database conntention
	db, err := database.NewDatabaseConnection()
	if err != nil {
		bcSuite.FailNowf("Database Error", "%v", err.Error())
	}
	app := app.AppSetup(db)
	bcSuite.app = app

}

func (bcSuite *BlogControllerSuite) TearDownSuite() {

}

func (bcSuite *BlogControllerSuite) TestCreateBlog() {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	_, err := bcSuite.app.Test(request, -1)
	if err != nil {
		bcSuite.FailNowf("App setup error", "%v", err.Error())
	}
}
