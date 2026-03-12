package services

import (
	"math"
	"strings"
	"sync"
	"time"

	"yandex2sound/backend/internal/models"
)

var Jobs sync.Map

var jobsMu sync.Mutex

type TransferService struct {
	soundCloud *SoundCloudService
}

type TransferGroup struct {
	PlaylistName string
	Tracks       []models.Track
}

func NewTransferService(soundCloud *SoundCloudService) *TransferService {
	return &TransferService{soundCloud: soundCloud}
}

func (s *TransferService) NewJob(total int) *models.TransferJob {
	job := &models.TransferJob{ID: newID(), Status: "pending", Total: total, Log: make([]models.TransferLog, 0, total)}
	Jobs.Store(job.ID, job)
	return job
}

func (s *TransferService) GetJob(jobID string) (*models.TransferJob, bool) {
	v, ok := Jobs.Load(jobID)
	if !ok {
		return nil, false
	}
	job, ok := v.(*models.TransferJob)
	if !ok {
		return nil, false
	}
	jobsMu.Lock()
	defer jobsMu.Unlock()
	copyLog := make([]models.TransferLog, len(job.Log))
	copy(copyLog, job.Log)
	clone := *job
	clone.Log = copyLog
	return &clone, true
}

func (s *TransferService) RunTransfer(jobID string, tracks []models.Track, scToken, scClientID, playlistName string) {
	setStatus(jobID, "running")
	playlistID, playlistURL, err := s.soundCloud.CreatePlaylist(scToken, playlistName)
	if err != nil {
		appendLog(jobID, models.TransferLog{TrackTitle: withErr("playlist_create", err), Status: "error"}, err != nil)
		setStatus(jobID, "error")
		return
	}

	for i, track := range tracks {
		query := strings.TrimSpace(track.Artists + " - " + track.Title)
		setCurrent(jobID, i+1, track.Title)
		trackID, _, found, err := s.withRetrySearch(scToken, query, scClientID)
		if err != nil {
			appendLog(jobID, models.TransferLog{TrackTitle: withErr("[search] "+query, err), Status: "error"}, true)
			time.Sleep(300 * time.Millisecond)
			continue
		}
		if !found {
			appendLog(jobID, models.TransferLog{TrackTitle: query, Status: "not_found"}, false)
			time.Sleep(300 * time.Millisecond)
			continue
		}
		if err := s.withRetryAdd(scToken, playlistID, trackID); err != nil {
			appendLog(jobID, models.TransferLog{TrackTitle: withErr("[add] "+query, err), Status: "error"}, true)
			time.Sleep(300 * time.Millisecond)
			continue
		}
		appendLog(jobID, models.TransferLog{TrackTitle: query, Status: "found"}, false)
		time.Sleep(300 * time.Millisecond)
	}

	jobsMu.Lock()
	if v, ok := Jobs.Load(jobID); ok {
		if job, ok := v.(*models.TransferJob); ok {
			job.ResultURL = playlistURL
			if job.Status != "error" {
				job.Status = "done"
			}
		}
	}
	jobsMu.Unlock()
}

func (s *TransferService) RunTransferGrouped(jobID string, groups []TransferGroup, scToken, scClientID string) {
	setStatus(jobID, "running")
	processed := 0
	firstResultURL := ""

	for _, group := range groups {
		playlistID, playlistURL, err := s.soundCloud.CreatePlaylist(scToken, group.PlaylistName)
		if err != nil {
			appendLog(jobID, models.TransferLog{TrackTitle: withErr("["+group.PlaylistName+"] playlist_create", err), Status: "error"}, true)
			continue
		}
		if firstResultURL == "" {
			firstResultURL = playlistURL
		}

		for _, track := range group.Tracks {
			processed++
			query := strings.TrimSpace(track.Artists + " - " + track.Title)
			displayTrack := "[" + group.PlaylistName + "] " + query
			setCurrent(jobID, processed, displayTrack)

			trackID, _, found, err := s.withRetrySearch(scToken, query, scClientID)
			if err != nil {
				appendLog(jobID, models.TransferLog{TrackTitle: withErr("[search] "+displayTrack, err), Status: "error"}, true)
				time.Sleep(300 * time.Millisecond)
				continue
			}
			if !found {
				appendLog(jobID, models.TransferLog{TrackTitle: displayTrack, Status: "not_found"}, false)
				time.Sleep(300 * time.Millisecond)
				continue
			}
			if err := s.withRetryAdd(scToken, playlistID, trackID); err != nil {
				appendLog(jobID, models.TransferLog{TrackTitle: withErr("[add] "+displayTrack, err), Status: "error"}, true)
				time.Sleep(300 * time.Millisecond)
				continue
			}
			appendLog(jobID, models.TransferLog{TrackTitle: displayTrack, Status: "found"}, false)
			time.Sleep(300 * time.Millisecond)
		}
	}

	jobsMu.Lock()
	if v, ok := Jobs.Load(jobID); ok {
		if job, ok := v.(*models.TransferJob); ok {
			job.ResultURL = firstResultURL
			if job.Status != "error" {
				job.Status = "done"
			}
		}
	}
	jobsMu.Unlock()
}

func (s *TransferService) withRetrySearch(token, query, clientID string) (string, string, bool, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		id, url, found, err := s.soundCloud.SearchTrack(token, query, clientID)
		if err == nil {
			return id, url, found, nil
		}
		lastErr = err
		time.Sleep(backoff(attempt))
	}
	return "", "", false, lastErr
}

func (s *TransferService) withRetryAdd(token, playlistID, trackID string) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		err := s.soundCloud.AddTrackToPlaylist(token, playlistID, trackID)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(backoff(attempt))
	}
	return lastErr
}

func backoff(attempt int) time.Duration {
	pow := math.Pow(2, float64(attempt))
	return time.Duration(int(pow)*250) * time.Millisecond
}

func appendLog(jobID string, entry models.TransferLog, isError bool) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	if v, ok := Jobs.Load(jobID); ok {
		if job, ok := v.(*models.TransferJob); ok {
			job.Log = append(job.Log, entry)
			switch entry.Status {
			case "found":
				job.Transferred++
			case "not_found":
				job.NotFound++
			case "error":
				job.Errors++
			}
			if isError {
				job.Status = "running"
			}
		}
	}
}

func setCurrent(jobID string, current int, track string) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	if v, ok := Jobs.Load(jobID); ok {
		if job, ok := v.(*models.TransferJob); ok {
			job.Current = current
			job.LastTrack = track
		}
	}
}

func setStatus(jobID, status string) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	if v, ok := Jobs.Load(jobID); ok {
		if job, ok := v.(*models.TransferJob); ok {
			job.Status = status
		}
	}
}

func newID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}

func withErr(title string, err error) string {
	if err == nil {
		return title
	}
	msg := strings.ReplaceAll(err.Error(), "\n", " ")
	msg = strings.TrimSpace(msg)
	if len(msg) > 180 {
		msg = msg[:180] + "..."
	}
	return title + " :: " + msg
}
