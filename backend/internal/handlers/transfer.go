package handlers

import (
	"bufio"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"

	"yandex2sound/backend/internal/models"
	"yandex2sound/backend/internal/services"
	"yandex2sound/backend/internal/session"
)

type TransferHandler struct {
	yandex   *services.YandexService
	transfer *services.TransferService
}

func NewTransferHandler(yandex *services.YandexService, transfer *services.TransferService) *TransferHandler {
	return &TransferHandler{yandex: yandex, transfer: transfer}
}

type transferReq struct {
	PlaylistIDs           []string `json:"playlist_ids"`
	PlaylistName          string   `json:"playlist_name"`
	PreserveOriginalNames bool     `json:"preserve_original_names"`
}

func (h *TransferHandler) StartTransfer(c *fiber.Ctx) error {
	s := session.Get(c)
	if s.YandexToken == "" || s.YandexUserID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "yandex is not connected")
	}
	if s.SoundCloudToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "soundcloud is not connected")
	}
	var req transferReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if len(req.PlaylistIDs) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "playlist_ids must not be empty")
	}
	if req.PlaylistName == "" {
		req.PlaylistName = "From Yandex"
	}

	if req.PreserveOriginalNames {
		playlists, err := h.yandex.GetPlaylists(s.YandexToken, s.YandexUserID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		titleByID := map[string]string{"liked": "Liked tracks"}
		for _, p := range playlists {
			titleByID[p.ID] = p.Title
		}

		groups := make([]services.TransferGroup, 0, len(req.PlaylistIDs))
		total := 0
		for _, id := range req.PlaylistIDs {
			var tracks []models.Track
			var err error
			if id == "liked" {
				tracks, err = h.yandex.GetLikedTracks(s.YandexToken, s.YandexUserID)
			} else {
				tracks, err = h.yandex.GetPlaylistTracks(s.YandexToken, s.YandexUserID, id)
			}
			if err != nil {
				return fiber.NewError(fiber.StatusBadGateway, err.Error())
			}
			total += len(tracks)
			name := titleByID[id]
			if name == "" {
				name = "Yandex " + id
			}
			groups = append(groups, services.TransferGroup{
				PlaylistName: name,
				Tracks:       tracks,
			})
		}
		job := h.transfer.NewJob(total)
		go h.transfer.RunTransferGrouped(job.ID, groups, s.SoundCloudToken, s.SoundCloudClientID)
		return c.JSON(fiber.Map{"job_id": job.ID})
	}

	trackMap := make(map[string]models.Track)
	for _, id := range req.PlaylistIDs {
		var tracks []models.Track
		var err error
		if id == "liked" {
			tracks, err = h.yandex.GetLikedTracks(s.YandexToken, s.YandexUserID)
		} else {
			tracks, err = h.yandex.GetPlaylistTracks(s.YandexToken, s.YandexUserID, id)
		}
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		for _, track := range tracks {
			if track.ID == "" {
				track.ID = track.Artists + "::" + track.Title
			}
			trackMap[track.ID] = track
		}
	}

	allTracks := make([]models.Track, 0, len(trackMap))
	for _, t := range trackMap {
		allTracks = append(allTracks, t)
	}

	job := h.transfer.NewJob(len(allTracks))
	go h.transfer.RunTransfer(job.ID, allTracks, s.SoundCloudToken, s.SoundCloudClientID, req.PlaylistName)
	return c.JSON(fiber.Map{"job_id": job.ID})
}

func (h *TransferHandler) ProgressSSE(c *fiber.Ctx) error {
	jobID := c.Params("job_id")
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			job, ok := h.transfer.GetJob(jobID)
			if !ok {
				_, _ = w.WriteString("event: error\n")
				_, _ = w.WriteString("data: {\"message\":\"job not found\"}\n\n")
				_ = w.Flush()
				return
			}
			payload, _ := json.Marshal(job)
			_, _ = w.WriteString("data: ")
			_, _ = w.Write(payload)
			_, _ = w.WriteString("\n\n")
			_ = w.Flush()

			if job.Status == "done" || job.Status == "error" {
				return
			}
			<-ticker.C
		}
	})
	return nil
}
