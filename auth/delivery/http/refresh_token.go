package http

import (
	"errors"
	"net/http"

	"github.com/adhtanjung/go-starter/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RefreshTokenHandler struct {
	UUsecase domain.UserUsecase
}

func NewRefreshTokenHandler(e *echo.Group, u domain.UserUsecase) (err error) {
	handler := &RefreshTokenHandler{
		UUsecase: u,
	}

	e.GET("", handler.GenerateRefreshToken)

	return
}

func (handler *RefreshTokenHandler) GenerateRefreshToken(c echo.Context) (err error) {
	userID := c.Get("user_id")
	if userID == nil {
		return errors.New("cannot parse user id")
	}
	ctx := c.Request().Context()

	toUUIDType, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}
	refreshToken, accessToken, err := handler.UUsecase.GetUsingRefreshToken(ctx, toUUIDType)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": accessToken, "refresh_token": refreshToken})

}
