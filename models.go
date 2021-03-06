package main

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type Remainder struct {
	To        string    `bson:"to" json:"to"`
	Title     string    `bson:"title" json:"title"`
	Message   string    `bson:"message" json:"message"`
	Type      string    `bson:"type" json:"type"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type (
	User struct {
		Username string `bson:"username" json:"username" validate:"required"`
		Password string `bson:"password" json:"password" validate:"required"`
	}

	/*Search struct {
		Filter string `bson:"filter" json:"filter" validate:"required"`
	}*/

	CustomValidator struct {
		validator *validator.Validate
	}

	Recipient struct {
		Email          string
		Role           string
		SlugIdentifier string
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
