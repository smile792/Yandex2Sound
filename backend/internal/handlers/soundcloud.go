package handlers

import (
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"

	"yandex2sound/backend/internal/services"
	"yandex2sound/backend/internal/session"
)

type SoundCloudHandler struct {
	soundCloud *services.SoundCloudService
}

func NewSoundCloudHandler(soundCloud *services.SoundCloudService) *SoundCloudHandler {
	return &SoundCloudHandler{soundCloud: soundCloud}
}

func (h *SoundCloudHandler) AuthURL(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"url": h.soundCloud.GetAuthURL()})
}

func (h *SoundCloudHandler) AuthCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "code query parameter is required")
	}
	token, err := h.soundCloud.ExchangeCode(code)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	s := session.Get(c)
	s.SoundCloudToken = token
	id, username, err := h.soundCloud.GetMe(token)
	if err == nil {
		s.SoundCloudUserID = id
		s.SoundCloudName = username
	} else if strings.Contains(err.Error(), "403") {
		// Some SoundCloud apps can exchange code for token, but /me is forbidden.
		// Keep token to allow follow-up API checks from app flow.
		s.SoundCloudUserID = ""
		s.SoundCloudName = "unknown"
	} else {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	redirectTo := os.Getenv("FRONTEND_URL")
	if redirectTo == "" {
		redirectTo = "http://localhost:5173"
	}
	return c.Redirect(redirectTo+"/?sc=connected", fiber.StatusTemporaryRedirect)
}

func (h *SoundCloudHandler) Status(c *fiber.Ctx) error {
	s := session.Get(c)
	return c.JSON(fiber.Map{"connected": s.SoundCloudToken != "", "username": s.SoundCloudName})
}

func (h *SoundCloudHandler) AuthStart(c *fiber.Ctx) error {
	u := h.soundCloud.GetAuthURL()
	return c.Redirect(u, fiber.StatusTemporaryRedirect)
}

func callbackURL(base string) string {
	u, _ := url.Parse(base)
	if u == nil {
		return ""
	}
	return u.String()
}
