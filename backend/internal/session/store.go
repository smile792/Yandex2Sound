package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

const (
	sessionCookieName = "y2s_session"
	sessionLocalKey   = "y2s_session_data"
	defaultMaxAge     = 60 * 60 * 24 * 30
)

type SessionData struct {
	YandexToken            string `json:"yandex_token,omitempty"`
	YandexUserID           string `json:"yandex_user_id,omitempty"`
	SoundCloudToken        string `json:"soundcloud_token,omitempty"`
	SoundCloudUserID       string `json:"soundcloud_user_id,omitempty"`
	SoundCloudName         string `json:"soundcloud_name,omitempty"`
	SoundCloudClientID     string `json:"soundcloud_client_id,omitempty"`
	SoundCloudClientSecret string `json:"soundcloud_client_secret,omitempty"`
}

var (
	keyOnce sync.Once
	key     [32]byte
)

func GetOrCreateSession(c *fiber.Ctx) (string, *SessionData) {
	if v := c.Locals(sessionLocalKey); v != nil {
		if data, ok := v.(*SessionData); ok {
			return sessionCookieName, data
		}
	}

	data := &SessionData{}
	raw := strings.TrimSpace(c.Cookies(sessionCookieName))
	if raw != "" {
		if decoded, err := decode(raw); err == nil {
			data = decoded
		}
	}

	c.Locals(sessionLocalKey, data)
	return sessionCookieName, data
}

func Get(c *fiber.Ctx) *SessionData {
	_, data := GetOrCreateSession(c)
	return data
}

func Save(c *fiber.Ctx) {
	v := c.Locals(sessionLocalKey)
	data, ok := v.(*SessionData)
	if !ok || data == nil {
		return
	}

	encoded, err := encode(data)
	if err != nil {
		return
	}

	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   cookieSecure(c),
		MaxAge:   defaultMaxAge,
	})
}

func encode(data *SessionData) (string, error) {
	plain, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(getKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	payload := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decode(raw string) (*SessionData, error) {
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(getKey())
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return nil, fiber.ErrUnauthorized
	}
	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	var data SessionData
	if err := json.Unmarshal(plain, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func getKey() []byte {
	keyOnce.Do(func() {
		secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
		if secret == "" {
			// Auto-generate a process-local key when SESSION_SECRET is not set.
			random := make([]byte, 32)
			if _, err := rand.Read(random); err == nil {
				sum := sha256.Sum256(random)
				key = sum
				return
			}
		}
		sum := sha256.Sum256([]byte(secret))
		key = sum
	})
	return key[:]
}

func cookieSecure(c *fiber.Ctx) bool {
	if strings.EqualFold(c.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	if strings.EqualFold(c.Protocol(), "https") {
		return true
	}
	frontendURL := strings.ToLower(strings.TrimSpace(os.Getenv("FRONTEND_URL")))
	return strings.HasPrefix(frontendURL, "https://")
}
