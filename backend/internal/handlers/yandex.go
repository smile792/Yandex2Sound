package handlers

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"yandex2sound/backend/internal/models"
	"yandex2sound/backend/internal/services"
	"yandex2sound/backend/internal/session"
)

type YandexHandler struct {
	yandex *services.YandexService
}

func NewYandexHandler(yandex *services.YandexService) *YandexHandler {
	return &YandexHandler{yandex: yandex}
}

type validateReq struct {
	Token string `json:"token"`
}

func (h *YandexHandler) ValidateToken(c *fiber.Ctx) error {
	var req validateReq
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token is required")
	}
	token := normalizeYandexToken(req.Token)
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "empty yandex token after normalization")
	}
	userID, err := h.yandex.ValidateToken(token)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid yandex token: "+err.Error())
	}
	s := session.Get(c)
	s.YandexToken = token
	s.YandexUserID = userID
	return c.JSON(fiber.Map{"ok": true, "user_id": userID})
}

func (h *YandexHandler) AuthURL(c *fiber.Ctx) error {
	if !h.yandex.HasOAuthConfig() {
		return fiber.NewError(fiber.StatusBadRequest, "yandex oauth is not configured")
	}
	return c.JSON(fiber.Map{"url": h.yandex.GetAuthURL()})
}

func (h *YandexHandler) AuthCallback(c *fiber.Ctx) error {
	if !h.yandex.HasOAuthConfig() {
		return fiber.NewError(fiber.StatusBadRequest, "yandex oauth is not configured")
	}
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "code query parameter is required")
	}
	token, err := h.yandex.ExchangeCode(code)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to exchange yandex oauth code: "+err.Error())
	}
	token = normalizeYandexToken(token)
	userID, err := h.yandex.ValidateToken(token)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "received invalid yandex token from oauth: "+err.Error())
	}

	s := session.Get(c)
	s.YandexToken = token
	s.YandexUserID = userID
	return c.Redirect(h.yandex.FrontendURL()+"/?ym=connected", fiber.StatusTemporaryRedirect)
}

func normalizeYandexToken(raw string) string {
	token := strings.TrimSpace(raw)
	if token == "" {
		return ""
	}

	// If user pasted full cookie header, extract Session_id value.
	re := regexp.MustCompile(`(?i)(?:^|;\s*)Session_id=([^;]+)`)
	if m := re.FindStringSubmatch(token); len(m) == 2 {
		token = m[1]
	}

	token = strings.TrimPrefix(token, "OAuth ")
	token = strings.TrimPrefix(token, "oauth ")
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")
	token = strings.TrimPrefix(token, "Session_id=")
	token = strings.TrimPrefix(token, "session_id=")
	token = strings.TrimSpace(token)

	if decoded, err := url.QueryUnescape(token); err == nil {
		token = decoded
	}
	return strings.TrimSpace(token)
}

func (h *YandexHandler) GetPlaylists(c *fiber.Ctx) error {
	s := session.Get(c)
	if s.YandexToken == "" || s.YandexUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "yandex is not connected")
	}
	playlists, err := h.yandex.GetPlaylists(s.YandexToken, s.YandexUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	items := make([]models.Playlist, 0, len(playlists)+1)
	likedTracks, err := h.yandex.GetLikedTracks(s.YandexToken, s.YandexUserID)
	if err == nil {
		items = append(items, models.Playlist{
			ID:         "liked",
			Title:      "Liked tracks ❤",
			TrackCount: len(likedTracks),
			CoverURL:   "",
		})
	}
	items = append(items, playlists...)
	return c.JSON(fiber.Map{"items": items})
}

func (h *YandexHandler) GetPlaylistTracks(c *fiber.Ctx) error {
	s := session.Get(c)
	if s.YandexToken == "" || s.YandexUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "yandex is not connected")
	}
	id := c.Params("id")
	if id == "liked" {
		tracks, err := h.yandex.GetLikedTracks(s.YandexToken, s.YandexUserID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		return c.JSON(fiber.Map{"items": tracks})
	}
	tracks, err := h.yandex.GetPlaylistTracks(s.YandexToken, s.YandexUserID, id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(fiber.Map{"items": tracks})
}

func (h *YandexHandler) Clear(c *fiber.Ctx) error {
	s := session.Get(c)
	s.YandexToken = ""
	s.YandexUserID = ""
	return c.JSON(fiber.Map{"ok": true})
}

func (h *YandexHandler) Status(c *fiber.Ctx) error {
	s := session.Get(c)
	return c.JSON(fiber.Map{
		"connected": s.YandexToken != "" && s.YandexUserID != "",
		"user_id":   s.YandexUserID,
	})
}
