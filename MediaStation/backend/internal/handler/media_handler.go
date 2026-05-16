package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"mediastation/config"
	"mediastation/internal/model"
	"mediastation/internal/service"
	"mediastation/pkg/mediautil"
)

type MediaHandler struct {
	mediaService service.MediaService
	config       *config.Config
}

func NewMediaHandler(mediaService service.MediaService, cfg *config.Config) *MediaHandler {
	return &MediaHandler{mediaService: mediaService, config: cfg}
}

func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	media, err := h.mediaService.GetMedia(uint(id))
	if err != nil {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}

func (h *MediaHandler) GetMediaList(w http.ResponseWriter, r *http.Request) {
	mediaType := model.MediaType(r.URL.Query().Get("type"))
	if mediaType == "" {
		mediaType = model.MediaTypeVideo
	}

	mediaList, err := h.mediaService.GetAllMedia(mediaType)
	if err != nil {
		http.Error(w, "Failed to get media list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaList)
}

func (h *MediaHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid series ID", http.StatusBadRequest)
		return
	}

	series, err := h.mediaService.GetSeries(uint(id))
	if err != nil {
		http.Error(w, "Series not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(series)
}

func (h *MediaHandler) GetSeriesList(w http.ResponseWriter, r *http.Request) {
	mediaType := model.MediaType(r.URL.Query().Get("type"))
	if mediaType == "" {
		mediaType = model.MediaTypeVideo
	}

	seriesList, err := h.mediaService.GetAllSeries(mediaType)
	if err != nil {
		http.Error(w, "Failed to get series list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seriesList)
}

func (h *MediaHandler) GetMediaBySeries(w http.ResponseWriter, r *http.Request) {
	seriesID, err := strconv.ParseUint(r.URL.Query().Get("series_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid series ID", http.StatusBadRequest)
		return
	}

	mediaList, err := h.mediaService.GetMediaBySeries(uint(seriesID))
	if err != nil {
		http.Error(w, "Failed to get media list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaList)
}

func (h *MediaHandler) SearchMedia(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "Keyword is required", http.StatusBadRequest)
		return
	}

	mediaList, err := h.mediaService.SearchMedia(keyword)
	if err != nil {
		http.Error(w, "Failed to search media", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mediaList)
}

func (h *MediaHandler) StreamMediaByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	h.streamMediaByID(w, r, uint(id))
}

func (h *MediaHandler) StreamMedia(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	h.streamMediaByID(w, r, uint(id))
}

func (h *MediaHandler) streamMediaByID(w http.ResponseWriter, r *http.Request, id uint) {
	filePath, err := h.mediaService.StreamMedia(uint(id))
	if err != nil {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	ext := filepath.Ext(filePath)

	if isDirectStreamFile(ext) {
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Failed to open media file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Failed to get file info", http.StatusInternalServerError)
			return
		}

		contentType := getContentType(ext)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.Header().Set("Accept-Ranges", "bytes")

		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			handleRangeRequest(w, r, file, fileInfo.Size())
			return
		}

		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		return
	}

	if mediautil.IsFFmpegAvailable() {
		log.Printf("Transcoding non-supported video format: %s", ext)
		streamWithFFmpeg(w, r, filePath)
		return
	}

	http.Error(w, "Unsupported media format", http.StatusUnsupportedMediaType)
}

func streamWithFFmpeg(w http.ResponseWriter, r *http.Request, filePath string) {
	var stderr bytes.Buffer

	startTime := "0"

	seekParam := r.URL.Query().Get("seek")
	if seekParam != "" {
		startTime = seekParam
		log.Printf("Seeking to: %s seconds", startTime)
	}

	args := []string{
		"-ss", startTime,
		"-accurate_seek",
		"-i", filePath,
		"-f", "mp4",
		"-vcodec", "libx264",
		"-acodec", "aac",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof+faststart",
		"-crf", "23",
		"-preset", "ultrafast",
		"-b:a", "128k",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-g", "25",
		"-keyint_min", "25",
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-",
	}

	log.Printf("Starting FFmpeg transcoding for: %s (seek: %s)", filePath, startTime)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe: %v", err)
		http.Error(w, "Failed to start transcoding", http.StatusInternalServerError)
		return
	}

	err = cmd.Start()
	if err != nil {
		log.Printf("FFmpeg start error: %v - %s", err, stderr.String())
		http.Error(w, fmt.Sprintf("Failed to start transcoding: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range, Origin, Content-Type")

	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("Warning: ResponseWriter does not support Flusher")
	}

	buf := make([]byte, 32*1024)
	clientGone := r.Context().Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	totalBytes := 0
	lastLogTime := time.Now()
	readTimeout := time.Now().Add(120 * time.Second)

	for {
		select {
		case <-clientGone:
			log.Println("Client disconnected")
			cmd.Process.Kill()
			return
		case <-ticker.C:
			if time.Now().After(readTimeout) {
				log.Printf("Stream timeout - killing ffmpeg process")
				cmd.Process.Kill()
				return
			}
			if totalBytes > 0 {
				elapsed := time.Since(lastLogTime).Seconds()
				if elapsed > 10 {
					log.Printf("Streaming: %d bytes sent", totalBytes)
					lastLogTime = time.Now()
				}
			}
		default:
			done := make(chan bool, 1)
			var readErr error
			var readN int

			go func() {
				readN, readErr = stdout.Read(buf)
				done <- true
			}()

			select {
			case <-done:
				if readErr != nil {
					if readErr == io.EOF {
						log.Println("FFmpeg stdout EOF reached")
						return
					}
					log.Printf("Stream read error: %v", readErr)
					cmd.Process.Kill()
					return
				}
				if readN > 0 {
					if _, writeErr := w.Write(buf[:readN]); writeErr != nil {
						log.Printf("Stream write error: %v", writeErr)
						cmd.Process.Kill()
						return
					}
					totalBytes += readN
					if flusher != nil {
						flusher.Flush()
					}
				}
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}
}

func getContentType(ext string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".flv":
		return "video/x-flv"
	case ".mp3":
		return "audio/mpeg"
	case ".ogg":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	default:
		return "application/octet-stream"
	}
}

func isBrowserSupportedVideo(ext string) bool {
	supported := []string{".mp4", ".webm", ".ogg"}
	for _, e := range supported {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func isDirectStreamFile(ext string) bool {
	supported := []string{".mp4", ".webm", ".ogg", ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".txt", ".md", ".json"}
	for _, e := range supported {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func (h *MediaHandler) HLSPlaylist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	filePath, err := h.mediaService.StreamMedia(uint(id))
	if err != nil {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	hlsDir := filepath.Join(h.config.MediaDir, "hls")
	playlistPath := mediautil.GetHLSPlaylistPath(hlsDir, uint(id))

	if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
		log.Printf("HLS not found for media %d, generating...", id)
		err = mediautil.GenerateHLS(filePath, hlsDir, uint(id))
		if err != nil {
			log.Printf("Failed to generate HLS: %v", err)
			http.Error(w, "Failed to generate HLS", http.StatusInternalServerError)
			return
		}
	}

	playlistContent, err := os.ReadFile(playlistPath)
	if err != nil {
		log.Printf("Failed to read HLS playlist: %v", err)
		http.Error(w, "Failed to read HLS playlist", http.StatusInternalServerError)
		return
	}

	segmentPattern := regexp.MustCompile(fmt.Sprintf("%d_([0-9]{3})\\.mp4", id))
	updatedContent := segmentPattern.ReplaceAllStringFunc(string(playlistContent), func(match string) string {
		matches := segmentPattern.FindStringSubmatch(match)
		if len(matches) == 2 {
			return fmt.Sprintf("/hls/segment?id=%d&segment=%s", id, matches[1])
		}
		return match
	})

	initPattern := regexp.MustCompile(fmt.Sprintf("%d_init\\.mp4", id))
	updatedContent = initPattern.ReplaceAllString(updatedContent, fmt.Sprintf("/hls/segment?id=%d&segment=init", id))

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(updatedContent))

	log.Printf("Served HLS playlist for media %d", id)
}

func (h *MediaHandler) HLSSegment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	segmentID := r.URL.Query().Get("segment")
	if segmentID == "" {
		http.Error(w, "Invalid segment ID", http.StatusBadRequest)
		return
	}

	hlsDir := filepath.Join(h.config.MediaDir, "hls")
	var segmentPath string

	if segmentID == "init" {
		segmentPath = mediautil.GetHLSSegmentPath(hlsDir, uint(id), "init")
	} else {
		segmentPath = mediautil.GetHLSSegmentPath(hlsDir, uint(id), segmentID)
	}

	if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
		log.Printf("HLS segment not found: %s", segmentPath)
		http.Error(w, "HLS segment not found", http.StatusNotFound)
		return
	}

	segmentFile, err := os.Open(segmentPath)
	if err != nil {
		log.Printf("Failed to open HLS segment: %v", err)
		http.Error(w, "Failed to open HLS segment", http.StatusInternalServerError)
		return
	}
	defer segmentFile.Close()

	fileInfo, err := segmentFile.Stat()
	if err != nil {
		log.Printf("Failed to get segment info: %v", err)
		http.Error(w, "Failed to get segment info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range, Accept")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	w.WriteHeader(http.StatusOK)
	io.Copy(w, segmentFile)

	log.Printf("Served HLS segment %s (%d bytes)", segmentPath, fileInfo.Size())
}

func handleRangeRequest(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64) {
	rangeHeader := r.Header.Get("Range")
	var start, end int64
	fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

	if end == 0 || end > fileSize-1 {
		end = fileSize - 1
	}

	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	w.WriteHeader(http.StatusPartialContent)

	file.Seek(start, io.SeekStart)
	buffer := make([]byte, 1024*1024)
	n, _ := io.CopyBuffer(w, io.LimitReader(file, end-start+1), buffer)
	fmt.Printf("Sent %d bytes\n", n)
}

func (h *MediaHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s (size: %d bytes)", handler.Filename, handler.Size)

	filePath := filepath.Join(h.config.MediaDir, handler.Filename)
	err = os.MkdirAll(h.config.MediaDir, 0755)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create media directory: %v", err), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("File saved to: %s", filePath)

	autoDetectedType := mediautil.GetMediaFileType(filePath)
	log.Printf("Auto-detected media type: %s", autoDetectedType)

	var mediaInfo *mediautil.MediaInfo
	var infoErr error

	log.Printf("Checking FFprobe availability...")
	if mediautil.IsFFprobeAvailable() {
		log.Println("FFprobe is available")
		log.Printf("Extracting media info from: %s", filePath)
		mediaInfo, infoErr = mediautil.GetMediaInfo(filePath)
		if infoErr != nil {
			log.Printf("Warning: FFprobe failed: %v", infoErr)
			log.Println("Falling back to simple extraction")
			mediaInfo, _ = mediautil.ExtractMediaInfoSimple(filePath)
		} else {
			log.Printf("Media info extracted: Duration=%ds, Width=%d, Height=%d, IsVertical=%v, Codec=%s",
				mediaInfo.Duration, mediaInfo.Width, mediaInfo.Height, mediaInfo.IsVertical, mediaInfo.Codec)
		}
	} else {
		log.Println("FFprobe is NOT available")
		log.Println("Using simple info extraction only")
		mediaInfo, _ = mediautil.ExtractMediaInfoSimple(filePath)
	}

	if mediaInfo.Duration == 0 && mediaInfo.Width == 0 && mediaInfo.Height == 0 {
		log.Printf("Warning: No media metadata extracted for file: %s", filePath)
	}

	title := r.FormValue("title")
	if title == "" {
		title = mediaInfo.Title
		log.Printf("Auto-generated title from filename: %s", title)
	}

	thumbnail := r.FormValue("thumbnail")
	if thumbnail == "" && mediaInfo.Width > 0 && mediautil.IsFFmpegAvailable() {
		thumbDir := filepath.Join(h.config.MediaDir, "thumbnails")
		thumbPath := filepath.Join(thumbDir, "thumb_"+strings.ReplaceAll(handler.Filename, ".", "_")+".jpg")

		err = mediautil.GenerateThumbnail(filePath, thumbPath)
		if err == nil {
			thumbnail = "thumbnails/" + filepath.Base(thumbPath)
			log.Printf("Auto-generated thumbnail: %s", thumbnail)
		} else {
			log.Printf("Warning: Failed to generate thumbnail: %v", err)
		}
	}

	duration := mediaInfo.Duration
	if formDuration := r.FormValue("duration"); formDuration != "" && formDuration != "0" {
		duration, _ = strconv.Atoi(formDuration)
	}
	log.Printf("Duration: %d seconds", duration)

	width := mediaInfo.Width
	if formWidth := r.FormValue("width"); formWidth != "" && formWidth != "0" {
		width, _ = strconv.Atoi(formWidth)
	}

	height := mediaInfo.Height
	if formHeight := r.FormValue("height"); formHeight != "" && formHeight != "0" {
		height, _ = strconv.Atoi(formHeight)
	}

	isVertical := mediaInfo.IsVertical
	if formVertical := r.FormValue("is_vertical"); formVertical != "" {
		isVertical = formVertical == "1"
	}
	log.Printf("Dimensions: %dx%d (vertical: %v)", width, height, isVertical)

	mediaType := model.MediaType(r.FormValue("type"))
	if mediaType == "" {
		mediaType = model.MediaType(autoDetectedType)
	}

	seriesID, _ := strconv.Atoi(r.FormValue("series_id"))
	season, _ := strconv.Atoi(r.FormValue("season"))
	episode, _ := strconv.Atoi(r.FormValue("episode"))

	seriesIDUint := uint(seriesID)

	media := &model.Media{
		Title:       title,
		Description: r.FormValue("description"),
		Type:        mediaType,
		FilePath:    handler.Filename,
		Thumbnail:   thumbnail,
		Duration:    duration,
		Width:       width,
		Height:      height,
		SeriesID:    seriesIDUint,
		Season:      season,
		Episode:     episode,
		IsVertical:  isVertical,
	}

	err = h.mediaService.CreateMedia(media)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save media to database: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Media saved to database with ID: %d", media.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"id":          media.ID,
		"title":       media.Title,
		"type":        string(media.Type),
		"duration":    media.Duration,
		"width":       media.Width,
		"height":      media.Height,
		"is_vertical": media.IsVertical,
		"thumbnail":   media.Thumbnail,
		"codec":       mediaInfo.Codec,
		"bitrate":     mediaInfo.Bitrate,
	})
}

func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	err = h.mediaService.DeleteMedia(uint(id))
	if err != nil {
		http.Error(w, "Failed to delete media", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *MediaHandler) AddSeries(w http.ResponseWriter, r *http.Request) {
	var series model.MediaSeries
	err := json.NewDecoder(r.Body).Decode(&series)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.mediaService.CreateSeries(&series)
	if err != nil {
		http.Error(w, "Failed to save series to database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      series.ID,
	})
}
