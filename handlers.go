package main

import (
	"codeconverter/CodeGenerator"
	"codeconverter/Config"
	"codeconverter/MessageQ"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

func MakeModel(c echo.Context) error {
	var project CodeGenerator.Project
	err := project.BindProject(c.Request())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	err = project.GenerateModel()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Attach result python file
	err = c.Attachment("./model.py", "model.py")
	if err != nil {
		return err
	}

	err = os.Remove("./model.py")
	if err != nil {
		return err
	}

	return nil
}

func Fit(c echo.Context) error {
	var project CodeGenerator.Project

	err := project.BindProject(c.Request())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Zip saved model
	targetBase := fmt.Sprintf("./%s/", project.UserId)
	files, err := CodeGenerator.GetFileLists(targetBase + "Model")
	if err != nil {
		return err
	}

	err = CodeGenerator.Zip(targetBase + "Model.zip", files)
	if err != nil {
		return err
	}

	// Request to GPU server
	body := project.GetTrainBody()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	cfg, err := Config.GetConfig()
	if err != nil {
		return err
	}

	conn, err := MessageQ.CreateConnection(cfg.Account, cfg.Pw, cfg.Host, cfg.VHost)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Publish(jsonBody)
	if err != nil {
		return err
	}

	return nil
}

func GetSavedModel(c echo.Context) error {
	userId := c.Request().Header.Get("id")
	target := fmt.Sprintf("./%s/", userId)
	return c.File(target + "Model.zip")
}