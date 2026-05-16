package mediautil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MediaInfo struct {
	Title      string
	Duration   int
	Width      int
	Height     int
	IsVertical bool
	Thumbnail  string
	Codec      string
	Bitrate    int
}

func GetMediaInfo(filePath string) (*MediaInfo, error) {
	info := &MediaInfo{}

	info.Title = extractTitle(filePath)

	if isVideoFile(filePath) {
		err := getVideoInfo(filePath, info)
		if err != nil {
			log.Printf("Warning: Failed to get video info using ffprobe: %v", err)
			return info, err
		}
	} else if isAudioFile(filePath) {
		err := getAudioInfo(filePath, info)
		if err != nil {
			log.Printf("Warning: Failed to get audio info using ffprobe: %v", err)
			return info, err
		}
	}

	return info, nil
}

func extractTitle(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	title := strings.TrimSuffix(base, ext)

	title = strings.ReplaceAll(title, "_", " ")
	title = strings.ReplaceAll(title, "-", " ")

	return strings.TrimSpace(title)
}

func isVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	videoExts := []string{".mp4", ".mkv", ".webm", ".mov", ".avi", ".flv", ".wmv", ".mpeg", ".mpg", ".m4v", ".3gp"}
	for _, e := range videoExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isAudioFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	audioExts := []string{".mp3", ".ogg", ".wav", ".flac", ".aac", ".m4a", ".wma", ".ape", ".alac"}
	for _, e := range audioExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

func isNovelFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	novelExts := []string{".txt", ".epub", ".mobi", ".pdf"}
	for _, e := range novelExts {
		if ext == e {
			return true
		}
	}
	return false
}

func getVideoInfo(filePath string, info *MediaInfo) error {
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		filePath,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffprobe error: %v - %s", err, stderr.String())
	}

	var ffprobeOutput struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Duration  string `json:"duration"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			CodecName string `json:"codec_name"`
			BitRate   string `json:"bit_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
	}

	err = json.Unmarshal(out.Bytes(), &ffprobeOutput)
	if err != nil {
		return fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	for _, stream := range ffprobeOutput.Streams {
		if stream.CodecType == "video" {
			info.Codec = stream.CodecName

			if stream.BitRate != "" {
				bitrate, _ := strconv.Atoi(stream.BitRate)
				info.Bitrate = bitrate / 1000
			}

			if stream.Duration != "" {
				dur, _ := strconv.ParseFloat(stream.Duration, 64)
				info.Duration = int(dur)
			} else if ffprobeOutput.Format.Duration != "" {
				dur, _ := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
				info.Duration = int(dur)
			}

			info.Width = stream.Width
			info.Height = stream.Height
			info.IsVertical = stream.Height > stream.Width
			break
		}
	}

	if info.Duration == 0 && ffprobeOutput.Format.Duration != "" {
		dur, _ := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
		info.Duration = int(dur)
	}

	return nil
}

func getAudioInfo(filePath string, info *MediaInfo) error {
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		filePath,
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffprobe error: %v - %s", err, stderr.String())
	}

	var ffprobeOutput struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Duration  string `json:"duration"`
			CodecName string `json:"codec_name"`
			BitRate   string `json:"bit_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	err = json.Unmarshal(out.Bytes(), &ffprobeOutput)
	if err != nil {
		return fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	for _, stream := range ffprobeOutput.Streams {
		if stream.CodecType == "audio" {
			info.Codec = stream.CodecName

			if stream.BitRate != "" {
				bitrate, _ := strconv.Atoi(stream.BitRate)
				info.Bitrate = bitrate / 1000
			}

			if stream.Duration != "" {
				dur, _ := strconv.ParseFloat(stream.Duration, 64)
				info.Duration = int(dur)
			} else if ffprobeOutput.Format.Duration != "" {
				dur, _ := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
				info.Duration = int(dur)
			}
			break
		}
	}

	if info.Duration == 0 && ffprobeOutput.Format.Duration != "" {
		dur, _ := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
		info.Duration = int(dur)
	}

	return nil
}

func GenerateThumbnail(inputPath, outputPath string) error {
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %v", err)
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-ss", "00:00:01.5",
		"-vframes", "1",
		"-q:v", "2",
		"-vf", "scale=640:-1",
		"-y",
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail error: %v - %s", err, stderr.String())
	}

	log.Printf("Generated thumbnail: %s", outputPath)
	return nil
}

func GenerateThumbnailAtTime(inputPath, outputPath string, timestamp string) error {
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %v", err)
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-ss", timestamp,
		"-vframes", "1",
		"-q:v", "2",
		"-vf", "scale=640:-1",
		"-y",
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail error: %v - %s", err, stderr.String())
	}

	return nil
}

func ExtractMediaInfoSimple(filePath string) (*MediaInfo, error) {
	info := &MediaInfo{}
	info.Title = extractTitle(filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return info, fmt.Errorf("file not found")
	}

	if isImageFile(filePath) {
		info.Width = 0
		info.Height = 0
	}

	return info, nil
}

func GetMediaFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	videoExts := []string{".mp4", ".mkv", ".webm", ".mov", ".avi", ".flv", ".wmv", ".mpeg", ".mpg", ".m4v", ".3gp"}
	for _, e := range videoExts {
		if ext == e {
			return "video"
		}
	}

	audioExts := []string{".mp3", ".ogg", ".wav", ".flac", ".aac", ".m4a", ".wma", ".ape", ".alac"}
	for _, e := range audioExts {
		if ext == e {
			return "audio"
		}
	}

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	for _, e := range imageExts {
		if ext == e {
			return "image"
		}
	}

	novelExts := []string{".txt", ".epub", ".mobi", ".pdf"}
	for _, e := range novelExts {
		if ext == e {
			return "novel"
		}
	}

	return "video"
}

func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	} else if seconds < 3600 {
		mins := seconds / 60
		secs := seconds % 60
		return fmt.Sprintf("%d分%d秒", mins, secs)
	} else {
		hours := seconds / 3600
		mins := (seconds % 3600) / 60
		secs := seconds % 60
		return fmt.Sprintf("%d小时%d分%d秒", hours, mins, secs)
	}
}

func IsFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func IsFFprobeAvailable() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}

func GetFFmpegVersion() (string, error) {
	cmd := exec.Command("ffmpeg", "-version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}
	return "", nil
}

func GenerateMultipleThumbnails(inputPath, outputDir string, count int) error {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	info, err := GetMediaInfo(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get media info: %v", err)
	}

	if info.Duration < count {
		count = info.Duration
	}

	interval := info.Duration / (count + 1)

	for i := 1; i <= count; i++ {
		timestamp := time.Duration(interval*i) * time.Second
		outputPath := filepath.Join(outputDir, fmt.Sprintf("thumb_%d.jpg", i))

		err := GenerateThumbnailAtTime(inputPath, outputPath, timestamp.String())
		if err != nil {
			log.Printf("Warning: Failed to generate thumbnail %d: %v", i, err)
		}
	}

	return nil
}

func GenerateHLS(inputPath, hlsDir string, mediaID uint) error {
	mediaDir := filepath.Join(hlsDir, fmt.Sprintf("%d", mediaID))
	err := os.MkdirAll(mediaDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	playlistPath := filepath.Join(mediaDir, fmt.Sprintf("%d.m3u8", mediaID))

	if _, err := os.Stat(playlistPath); err == nil {
		log.Printf("HLS playlist already exists for media %d: %s", mediaID, playlistPath)
		return nil
	}

	log.Printf("Generating HLS (fMP4) for media %d: %s", mediaID, inputPath)

	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-ar", "44100",
		"-b:a", "128k",
		"-g", "30",
		"-sc_threshold", "0",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(mediaDir, fmt.Sprintf("%d_%%03d.mp4", mediaID)),
		"-hls_flags", "independent_segments",
		"-hls_fmp4_init_filename", filepath.Join(mediaDir, fmt.Sprintf("%d_init.mp4", mediaID)),
		"-hls_allow_cache", "1",
		"-avoid_negative_ts", "make_zero",
		"-copyts",
		"-fflags", "+genpts",
		"-preset", "ultrafast",
		"-crf", "23",
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		playlistPath,
	}

	log.Printf("FFmpeg command: ffmpeg %s", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg HLS error: %v - %s", err, stderr.String())
	}

	playlistContent, err := os.ReadFile(playlistPath)
	if err != nil {
		return fmt.Errorf("failed to read playlist: %v", err)
	}

	if !strings.HasSuffix(string(playlistContent), "#EXT-X-ENDLIST\n") && !strings.HasSuffix(string(playlistContent), "#EXT-X-ENDLIST") {
		err = os.WriteFile(playlistPath, []byte(string(playlistContent)+"\n#EXT-X-ENDLIST\n"), 0644)
		if err != nil {
			return fmt.Errorf("failed to add ENDLIST tag: %v", err)
		}
	}

	log.Printf("HLS (fMP4) generated successfully for media %d: %s", mediaID, playlistPath)
	return nil
}

func GetHLSPlaylistPath(hlsDir string, mediaID uint) string {
	mediaDir := filepath.Join(hlsDir, fmt.Sprintf("%d", mediaID))
	return filepath.Join(mediaDir, fmt.Sprintf("%d.m3u8", mediaID))
}

func GetHLSSegmentPath(hlsDir string, mediaID uint, segmentID string) string {
	mediaDir := filepath.Join(hlsDir, fmt.Sprintf("%d", mediaID))
	return filepath.Join(mediaDir, fmt.Sprintf("%d_%s.mp4", mediaID, segmentID))
}
