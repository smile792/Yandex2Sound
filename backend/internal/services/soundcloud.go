package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type SoundCloudService struct {
	apiBase      string
	oauthBase    string
	clientID     string
	clientSecret string
	redirectURI  string
	frontendURL  string
	httpClient   *http.Client
}

func NewSoundCloudService() *SoundCloudService {
	return &SoundCloudService{
		apiBase:      "https://api-v2.soundcloud.com",
		oauthBase:    "https://api.soundcloud.com",
		clientID:     os.Getenv("SOUNDCLOUD_CLIENT_ID"),
		clientSecret: os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
		redirectURI:  os.Getenv("SOUNDCLOUD_REDIRECT_URI"),
		frontendURL:  os.Getenv("FRONTEND_URL"),
		httpClient:   &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *SoundCloudService) GetAuthURL() string {
	q := url.Values{}
	q.Set("client_id", s.clientID)
	q.Set("redirect_uri", s.redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "non-expiring")
	return "https://soundcloud.com/connect?" + q.Encode()
}

func (s *SoundCloudService) ExchangeCode(code string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", s.clientID)
	form.Set("client_secret", s.clientSecret)
	form.Set("redirect_uri", s.redirectURI)

	resp, err := s.httpClient.Post(
		s.oauthBase+"/oauth2/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("exchange failed: %s", string(body))
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	token, _ := payload["access_token"].(string)
	if token == "" {
		return "", fmt.Errorf("no access token returned")
	}
	return token, nil
}

func (s *SoundCloudService) GetMe(token string) (string, string, error) {
	req, _ := http.NewRequest(http.MethodGet, s.oauthBase+"/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		_ = resp.Body.Close()
		fallbackReq, _ := http.NewRequest(http.MethodGet, s.oauthBase+"/me", nil)
		fallbackReq.Header.Set("Authorization", "OAuth "+token)
		resp, err = s.httpClient.Do(fallbackReq)
		if err != nil {
			return "", "", err
		}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("get me failed: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", err
	}
	id := fmt.Sprintf("%.0f", payload["id"])
	username, _ := payload["username"].(string)
	return id, username, nil
}

func (s *SoundCloudService) SearchTrack(token, query string) (string, string, bool, error) {
	parseCollection := func(body []byte) (string, string, bool, error) {
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", "", false, err
		}
		collection, _ := payload["collection"].([]any)
		if len(collection) == 0 {
			return "", "", false, nil
		}
		first, _ := collection[0].(map[string]any)
		id := fmt.Sprintf("%.0f", first["id"])
		permalink, _ := first["permalink_url"].(string)
		return id, permalink, id != "", nil
	}

	parseTracksArray := func(body []byte) (string, string, bool, error) {
		var arr []map[string]any
		if err := json.Unmarshal(body, &arr); err != nil {
			return "", "", false, err
		}
		if len(arr) == 0 {
			return "", "", false, nil
		}
		id := fmt.Sprintf("%.0f", arr[0]["id"])
		permalink, _ := arr[0]["permalink_url"].(string)
		return id, permalink, id != "", nil
	}

	// 1) Preferred: legacy /tracks search on api.soundcloud.com with auth.
	u1 := s.oauthBase + "/tracks?q=" + url.QueryEscape(query) + "&limit=1"
	req1, _ := http.NewRequest(http.MethodGet, u1, nil)
	resp1, err1 := s.doAuthed(req1, token)
	if err1 == nil {
		defer resp1.Body.Close()
		body1, _ := io.ReadAll(resp1.Body)
		if resp1.StatusCode < 400 {
			return parseTracksArray(body1)
		}
	}

	// 2) Fallback: api-v2 search with auth.
	u2 := s.apiBase + "/search/tracks?q=" + url.QueryEscape(query) + "&limit=1&client_id=" + url.QueryEscape(s.clientID)
	req2, _ := http.NewRequest(http.MethodGet, u2, nil)
	resp2, err2 := s.doAuthed(req2, token)
	if err2 == nil {
		if resp2.StatusCode == http.StatusUnauthorized || resp2.StatusCode == http.StatusForbidden {
			_ = resp2.Body.Close()
			noAuthReq, _ := http.NewRequest(http.MethodGet, u2, nil)
			resp2, err2 = s.httpClient.Do(noAuthReq)
		}
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			if resp2.StatusCode < 400 {
				return parseCollection(body2)
			}
			return "", "", false, fmt.Errorf("search failed: %s - %s", resp2.Status, strings.TrimSpace(string(body2)))
		}
	}

	if err1 != nil {
		return "", "", false, err1
	}
	if err2 != nil {
		return "", "", false, err2
	}
	return "", "", false, fmt.Errorf("search failed on all endpoints")
}

func (s *SoundCloudService) CreatePlaylist(token, name string) (string, string, error) {
	body := map[string]any{"playlist": map[string]any{"title": name, "sharing": "public", "tracks": []any{}}}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, s.oauthBase+"/playlists", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.doAuthed(req, token)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("create playlist failed: %s - %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var payload map[string]any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return "", "", err
	}
	id := fmt.Sprintf("%.0f", payload["id"])
	permalink, _ := payload["permalink_url"].(string)
	return id, permalink, nil
}

func (s *SoundCloudService) AddTrackToPlaylist(token, playlistID, trackID string) error {
	getURL := s.oauthBase + "/playlists/" + url.PathEscape(playlistID)
	getReq, _ := http.NewRequest(http.MethodGet, getURL, nil)
	getResp, err := s.doAuthed(getReq, token)
	if err != nil {
		return err
	}
	defer getResp.Body.Close()
	if getResp.StatusCode >= 400 {
		return fmt.Errorf("get playlist failed: %s", getResp.Status)
	}
	currentBody, _ := io.ReadAll(getResp.Body)
	var payload map[string]any
	if err := json.Unmarshal(currentBody, &payload); err != nil {
		return err
	}
	tracksRaw, _ := payload["tracks"].([]any)
	tracks := make([]map[string]string, 0, len(tracksRaw)+1)
	for _, t := range tracksRaw {
		m, _ := t.(map[string]any)
		id := fmt.Sprintf("%.0f", m["id"])
		if id != "" {
			tracks = append(tracks, map[string]string{"id": id})
		}
	}
	tracks = append(tracks, map[string]string{"id": trackID})

	putBody, _ := json.Marshal(map[string]any{"playlist": map[string]any{"tracks": tracks}})
	putURL := s.oauthBase + "/playlists/" + url.PathEscape(playlistID)
	putReq, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putResp, err := s.doAuthed(putReq, token)
	if err != nil {
		return err
	}
	defer putResp.Body.Close()
	if putResp.StatusCode >= 400 {
		respText, _ := io.ReadAll(putResp.Body)
		return fmt.Errorf("put playlist failed: %s - %s", putResp.Status, strings.TrimSpace(string(respText)))
	}
	return nil
}

func (s *SoundCloudService) doAuthed(req *http.Request, token string) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		return resp, nil
	}

	_ = resp.Body.Close()
	fallbackReq := req.Clone(req.Context())
	fallbackReq.Header = req.Header.Clone()
	fallbackReq.Header.Set("Authorization", "OAuth "+token)
	if req.GetBody != nil {
		if body, err := req.GetBody(); err == nil {
			fallbackReq.Body = body
		}
	}
	return s.httpClient.Do(fallbackReq)
}
