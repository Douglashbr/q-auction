// Package httpapi é o adapter de entrada HTTP (Fiber). Traduz requisições em chamadas
// aos casos de uso e erros em respostas no formato {"error":{"code","message"}}.
package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewApp(appName string) *fiber.App {
	return fiber.New(fiber.Config{AppName: appName, ErrorHandler: errorHandler})
}

// errorHandler é o único ponto que transforma erro em resposta HTTP.
// Erros de domínio devem ser mapeados aqui (errors.Is) para o status adequado.
func errorHandler(c fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "internal error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		message = fiberErr.Message
	}

	return c.Status(status).JSON(errorBody{Error: errorDetail{Code: statusCode(status), Message: message}})
}

// statusCode converte o status em código snake_case, ex.: 404 → "not_found".
func statusCode(status int) string {
	return strings.ReplaceAll(strings.ToLower(http.StatusText(status)), " ", "_")
}
