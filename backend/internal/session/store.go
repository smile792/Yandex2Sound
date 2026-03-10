package session

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SessionData struct {
	YandexToken      string
	YandexUserID     string
	SoundCloudToken  string
	SoundCloudUserID string
	SoundCloudName   string
}

var store sync.Map

func GetOrCreateSession(c *fiber.Ctx) (string, *SessionData) {
	sid := c.Cookies("sid")
	if sid == "" {
		sid = uuid.NewString()
		c.Cookie(&fiber.Cookie{
			Name:     "sid",
			Value:    sid,
			Path:     "/",
			HTTPOnly: true,
			SameSite: "Lax",
		})
	}

	if v, ok := store.Load(sid); ok {
		if data, okData := v.(*SessionData); okData {
			return sid, data
		}
	}

	data := &SessionData{}
	store.Store(sid, data)
	return sid, data
}

func Get(c *fiber.Ctx) *SessionData {
	_, data := GetOrCreateSession(c)
	return data
}
