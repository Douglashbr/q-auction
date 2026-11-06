package httpapi

import "github.com/gofiber/fiber/v3"

type healthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Instance string `json:"instance"`
}

type HealthHandler struct {
	service  string
	instance string
}

func NewHealthHandler(service, instance string) *HealthHandler {
	return &HealthHandler{service: service, instance: instance}
}

func (h *HealthHandler) Register(app *fiber.App) {
	app.Get("/health", h.check)
}

// check responde com a réplica que atendeu (instance), o que permite ver o round robin do LB.
func (h *HealthHandler) check(c fiber.Ctx) error {
	return c.JSON(healthResponse{Status: "ok", Service: h.service, Instance: h.instance})
}
