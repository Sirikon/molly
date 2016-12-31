package main

import (
	"errors"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

// DaemonAction .
type DaemonAction struct {
	ProjectLogic ProjectLogic
}

// Run defines a CLI action which initializes the daemon
// which listens HTTP requests
func (da DaemonAction) Run(c *cli.Context) error {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.POST("/deploy", func(req *gin.Context) {

		project := Project{}
		var uploadedFile multipart.File

		p := Promise{}
		p.Then(func() error {
			return da.ProjectLogic.GetByName(req.PostForm("project"), &project)
		})
		p.Then(func() error {
			var receivedToken = req.PostForm("token")
			var tokenIsCorrect = da.ProjectLogic.CheckToken(project, receivedToken)
			if !tokenIsCorrect {
				return errors.New("Wrong token")
			}
			return nil
		})
		p.Then(func() error {
			file, _, err := req.Request.FormFile("artifact")
			if err != nil {
				return errors.New("Error while reading the uploaded file")
			}
			uploadedFile = file
			return nil
		})
		p.Then(func() error {
			return da.ProjectLogic.StoreArtifact(project, uploadedFile)
		})
		p.Then(func() error {
			return da.ProjectLogic.RunDeploymentScript(project)
		})
		p.Then(func() error {
			return da.ProjectLogic.RestartService(project)
		})
		p.Then(func() error {
			req.String(200, "Done\n")
			return nil
		})
		p.Catch(func(err error) {
			req.String(400, err.Error()+"\n")
		})
		p.Run()

	})

	router.Run(":8080")

	return nil
}
