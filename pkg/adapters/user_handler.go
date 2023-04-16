// 特定のフレームワーク（Echo）を使用して、usecaseを実行できるようにする
package adapters

import (
	"go-app/pkg/domain"
	"go-app/pkg/usecase"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
)

type UserHandler struct {
	UserRepo domain.UserRepository
}

func (h *UserHandler) Register(c echo.Context) error {
	user := new(domain.User)
	if err := c.Bind(user); err != nil {
		return err
	}

	err := usecase.StoreUser(user, h.UserRepo)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "新規登録に成功しました")
}

func (h *UserHandler) Login(c echo.Context) error {
	req := new(domain.LoginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	err := godotenv.Load()
	if err != nil {
		return err
	}

	token, err := usecase.AuthenticateUser(req.Email, req.Password, h.UserRepo, os.Getenv("jwtSecretKey"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, token)
}
