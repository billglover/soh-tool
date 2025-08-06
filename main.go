package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

//go:embed template.md
var templateFile string

type Video struct {
	Title       string
	PublishedAt time.Time
	VideoID     string
	Description string
}

type TemplateData struct {
	PlaylistTitle string
	Videos        []Video
}

var days int

type YouTubeService interface {
	GetPlaylistTitle(playlistID string) (string, error)
	GetPlaylistItems(playlistID string) ([]*youtube.PlaylistItem, error)
}

type youtubeServiceImpl struct {
	service *youtube.Service
}

func (y *youtubeServiceImpl) GetPlaylistTitle(playlistID string) (string, error) {
	call := y.service.Playlists.List([]string{"snippet"}).Id(playlistID)
	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("error fetching playlist details: %w", err)
	}
	if len(response.Items) == 0 {
		return "", fmt.Errorf("playlist not found")
	}
	return response.Items[0].Snippet.Title, nil
}

func (y *youtubeServiceImpl) GetPlaylistItems(playlistID string) ([]*youtube.PlaylistItem, error) {
	call := y.service.PlaylistItems.List([]string{"snippet"}).
		PlaylistId(playlistID).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching playlist items: %w", err)
	}
	return response.Items, nil
}

type App struct {
	youtubeService YouTubeService
}

func main() {
	flag.IntVar(&days, "days", 30, "Number of days to look back for new videos")
	flag.Parse()

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "YOUTUBE_API_KEY environment variable not set")
		os.Exit(1)
	}

	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating YouTube service: %v\n", err)
		os.Exit(1)
	}

	app := &App{
		youtubeService: &youtubeServiceImpl{service: youtubeService},
	}

	if err := app.run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (app *App) run() error {
	playlistID := "PLgGXSWYM2FpNjXSFUQfFyXmFk3ENUZMGL"

	playlistTitle, err := app.youtubeService.GetPlaylistTitle(playlistID)
	if err != nil {
		return err
	}

	items, err := app.youtubeService.GetPlaylistItems(playlistID)
	if err != nil {
		return err
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)
	var videos []Video

	// Regexes to remove boilerplate phrases
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`(?i)You can participate in our live stream to ask questions or catch the replay on your preferred podcast platform.`),
	}

	for _, item := range items {
		publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			return fmt.Errorf("error parsing date: %w", err)
		}

		if publishedAt.After(cutoffDate) {
			// Clean up title and description
			title := strings.TrimSpace(item.Snippet.Title)
			description := strings.TrimSpace(item.Snippet.Description)

			for _, re := range regexes {
				description = re.ReplaceAllString(description, "")
			}

			description = strings.TrimSpace(description) // Trim again after replacement

			videos = append(videos, Video{
				Title:       title,
				PublishedAt: publishedAt,
				VideoID:     item.Snippet.ResourceId.VideoId,
				Description: description,
			})
		}
	}

	// Sort videos by published date, newest first
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].PublishedAt.After(videos[j].PublishedAt)
	})

	data := TemplateData{
		PlaylistTitle: playlistTitle,
		Videos:        videos,
	}

	tmpl, err := template.New("template").Parse(templateFile)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	if err := tmpl.Execute(os.Stdout, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}
