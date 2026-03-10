package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"yandex2sound/backend/internal/models"
)

type YandexService struct {
	baseURL      string
	oauthBaseURL string
	clientID     string
	clientSecret string
	redirectURI  string
	frontendURL  string
	client       *http.Client
}

func NewYandexService() *YandexService {
	return &YandexService{
		baseURL:      "https://api.music.yandex.net",
		oauthBaseURL: "https://oauth.yandex.ru",
		clientID:     os.Getenv("YANDEX_CLIENT_ID"),
		clientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
		redirectURI:  os.Getenv("YANDEX_REDIRECT_URI"),
		frontendURL:  os.Getenv("FRONTEND_URL"),
		client:       &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *YandexService) HasOAuthConfig() bool {
	return s.clientID != "" && s.clientSecret != "" && s.redirectURI != ""
}

func (s *YandexService) GetAuthURL() string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", s.clientID)
	q.Set("redirect_uri", s.redirectURI)
	q.Set("force_confirm", "yes")
	return s.oauthBaseURL + "/authorize?" + q.Encode()
}

func (s *YandexService) ExchangeCode(code string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", s.clientID)
	form.Set("client_secret", s.clientSecret)
	form.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequest(http.MethodPost, s.oauthBaseURL+"/token", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("yandex oauth exchange failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	token, _ := payload["access_token"].(string)
	if token == "" {
		return "", fmt.Errorf("access_token not found in oauth response")
	}
	return token, nil
}

func (s *YandexService) FrontendURL() string {
	if s.frontendURL != "" {
		return s.frontendURL
	}
	return "http://localhost:5173"
}

func (s *YandexService) doRequest(ctx context.Context, token, path string) ([]byte, error) {
	resp, body, err := s.doRequestWithHeaders(ctx, token, path, true)
	if err != nil {
		return nil, err
	}
	// Some session tokens from browser cookies are rejected as OAuth header but
	// accepted as Session_id cookie.
	if resp.StatusCode == http.StatusUnauthorized {
		resp2, body2, err2 := s.doRequestWithHeaders(ctx, token, path, false)
		if err2 == nil {
			resp = resp2
			body = body2
		}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yandex api error: %s", resp.Status)
	}
	return body, nil
}

func (s *YandexService) doRequestWithHeaders(ctx context.Context, token, path string, useOAuth bool) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-Yandex-Music-Client", "WindowsPhone/3.20")
	if useOAuth {
		req.Header.Set("Authorization", "OAuth "+token)
	} else {
		req.Header.Set("Cookie", "Session_id="+token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

func (s *YandexService) ValidateToken(token string) (string, error) {
	parseID := func(body []byte) (string, error) {
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", err
		}
		// Try strict known locations first.
		result, _ := payload["result"].(map[string]any)
		account, _ := result["account"].(map[string]any)
		candidates := []string{
			asString(account["uid"]),
			asString(account["id"]),
			asString(result["uid"]),
			asString(result["id"]),
			asString(account["login"]),
			asString(result["login"]),
		}
		for _, v := range candidates {
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), nil
			}
		}
		// Fallback: recursively scan whole payload for uid/id/login.
		if found := findFirstID(payload); found != "" {
			return found, nil
		}
		return "", nil
	}

	body, err := s.doRequest(context.Background(), token, "/account/status")
	if err == nil {
		id, parseErr := parseID(body)
		if parseErr != nil {
			return "", parseErr
		}
		if id != "" {
			return id, nil
		}
	}

	// Some tokens return useful user identity via /users/me.
	bodyMe, errMe := s.doRequest(context.Background(), token, "/users/me")
	if errMe == nil {
		id, parseErr := parseID(bodyMe)
		if parseErr != nil {
			return "", parseErr
		}
		if id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not parse user id from yandex responses")
}

func findFirstID(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, key := range []string{"uid", "id", "login"} {
			if raw, ok := t[key]; ok {
				if s := strings.TrimSpace(asString(raw)); s != "" {
					return s
				}
			}
		}
		for _, child := range t {
			if found := findFirstID(child); found != "" {
				return found
			}
		}
	case []any:
		for _, child := range t {
			if found := findFirstID(child); found != "" {
				return found
			}
		}
	}
	return ""
}

func (s *YandexService) GetPlaylists(token, userID string) ([]models.Playlist, error) {
	path := fmt.Sprintf("/users/%s/playlists/list", url.PathEscape(userID))
	body, err := s.doRequest(context.Background(), token, path)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	result, _ := payload["result"].([]any)
	playlists := make([]models.Playlist, 0, len(result))
	for _, raw := range result {
		obj, _ := raw.(map[string]any)
		coverObj, _ := obj["cover"].(map[string]any)
		coverURL := asString(coverObj["uri"])
		coverURL = strings.ReplaceAll(coverURL, "%%", "200x200")
		if strings.HasPrefix(coverURL, "avatars.yandex.net") {
			coverURL = "https://" + coverURL
		}
		playlists = append(playlists, models.Playlist{
			ID:         asString(obj["kind"]),
			Title:      asString(obj["title"]),
			TrackCount: asInt(obj["trackCount"]),
			CoverURL:   coverURL,
		})
	}
	return playlists, nil
}

func (s *YandexService) GetLikedTracks(token, userID string) ([]models.Track, error) {
	likesPath := fmt.Sprintf("/users/%s/likes/tracks", url.PathEscape(userID))
	body, err := s.doRequest(context.Background(), token, likesPath)
	if err != nil {
		return nil, err
	}
	var likes map[string]any
	if err := json.Unmarshal(body, &likes); err != nil {
		return nil, err
	}
	result, _ := likes["result"].(map[string]any)
	library, _ := result["library"].(map[string]any)
	tracks, _ := library["tracks"].([]any)
	ids := make([]string, 0, len(tracks))
	for _, tr := range tracks {
		m, _ := tr.(map[string]any)
		ids = append(ids, asString(m["id"]))
	}
	if len(ids) == 0 {
		return []models.Track{}, nil
	}
	trackBody, err := s.doRequest(context.Background(), token, "/tracks?track-ids="+url.QueryEscape(strings.Join(ids, ",")))
	if err != nil {
		return nil, err
	}
	var trackPayload map[string]any
	if err := json.Unmarshal(trackBody, &trackPayload); err != nil {
		return nil, err
	}
	items, _ := trackPayload["result"].([]any)
	return mapTracks(items), nil
}

func (s *YandexService) GetPlaylistTracks(token, userID, playlistKind string) ([]models.Track, error) {
	path := fmt.Sprintf("/users/%s/playlists/%s", url.PathEscape(userID), url.PathEscape(playlistKind))
	body, err := s.doRequest(context.Background(), token, path)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	result, _ := payload["result"].(map[string]any)
	tracks, _ := result["tracks"].([]any)
	items := make([]any, 0, len(tracks))
	for _, wrapped := range tracks {
		obj, _ := wrapped.(map[string]any)
		if tr, ok := obj["track"]; ok {
			items = append(items, tr)
		}
	}
	return mapTracks(items), nil
}

func mapTracks(items []any) []models.Track {
	out := make([]models.Track, 0, len(items))
	for _, raw := range items {
		obj, _ := raw.(map[string]any)
		artistsArr, _ := obj["artists"].([]any)
		artistNames := make([]string, 0, len(artistsArr))
		for _, a := range artistsArr {
			am, _ := a.(map[string]any)
			name := asString(am["name"])
			if name != "" {
				artistNames = append(artistNames, name)
			}
		}
		albums, _ := obj["albums"].([]any)
		album := ""
		cover := ""
		if len(albums) > 0 {
			alb, _ := albums[0].(map[string]any)
			album = asString(alb["title"])
			cover = asString(alb["coverUri"])
			cover = strings.ReplaceAll(cover, "%%", "200x200")
			if strings.HasPrefix(cover, "avatars.yandex.net") {
				cover = "https://" + cover
			}
		}
		out = append(out, models.Track{
			ID:         asString(obj["id"]),
			Title:      asString(obj["title"]),
			Artists:    strings.Join(artistNames, ", "),
			Album:      album,
			DurationMs: asInt(obj["durationMs"]),
			CoverURL:   cover,
		})
	}
	return out
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	default:
		return ""
	}
}

func asInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		return 0
	}
}
