package handlers

import (
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
	s := session.Get(c)
	u, err := h.soundCloud.GetAuthURL(s.SoundCloudClientID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"url": u})
}

type soundCloudConfigReq struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (h *SoundCloudHandler) SetConfig(c *fiber.Ctx) error {
	var req soundCloudConfigReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	clientID := strings.TrimSpace(req.ClientID)
	clientSecret := strings.TrimSpace(req.ClientSecret)
	if clientID == "" || clientSecret == "" {
		return fiber.NewError(fiber.StatusBadRequest, "client_id and client_secret are required")
	}

	s := session.Get(c)
	s.SoundCloudClientID = clientID
	s.SoundCloudClientSecret = clientSecret
	// Existing token belongs to old app credentials and should be re-authorized.
	s.SoundCloudToken = ""
	s.SoundCloudUserID = ""
	s.SoundCloudName = ""
	return c.JSON(fiber.Map{"ok": true})
}

func (h *SoundCloudHandler) ClearConfig(c *fiber.Ctx) error {
	s := session.Get(c)
	s.SoundCloudClientID = ""
	s.SoundCloudClientSecret = ""
	s.SoundCloudToken = ""
	s.SoundCloudUserID = ""
	s.SoundCloudName = ""
	return c.JSON(fiber.Map{"ok": true})
}

func (h *SoundCloudHandler) AuthCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "code query parameter is required")
	}

	s := session.Get(c)
	token, err := h.soundCloud.ExchangeCode(code, s.SoundCloudClientID, s.SoundCloudClientSecret)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
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
	clientID := strings.TrimSpace(s.SoundCloudClientID)
	if clientID == "" {
		clientID = h.soundCloud.DefaultClientID()
	}
	hasClientSecret := strings.TrimSpace(s.SoundCloudClientSecret) != ""
	if !hasClientSecret {
		hasClientSecret = h.soundCloud.HasDefaultClientSecret()
	}

	return c.JSON(fiber.Map{
		"connected":         s.SoundCloudToken != "",
		"username":          s.SoundCloudName,
		"client_id":         clientID,
		"has_client_secret": hasClientSecret,
	})
}

func (h *SoundCloudHandler) AuthStart(c *fiber.Ctx) error {
	s := session.Get(c)
	u, err := h.soundCloud.GetAuthURL(s.SoundCloudClientID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Redirect(u, fiber.StatusTemporaryRedirect)
}
