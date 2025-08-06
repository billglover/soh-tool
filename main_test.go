package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/youtube/v3"
)

// mockYouTubeService is a mock implementation of the YouTubeService interface.

type mockYouTubeService struct {
	playlistTitle string
	playlistItems []*youtube.PlaylistItem
	err           error
}

func (m *mockYouTubeService) GetPlaylistTitle(playlistID string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.playlistTitle, nil
}

func (m *mockYouTubeService) GetPlaylistItems(playlistID string) ([]*youtube.PlaylistItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.playlistItems, nil
}

func TestRun(t *testing.T) {
	tests := []struct {
		name                 string
		days                 int
		mockService          *mockYouTubeService
		expectedToContain    string
		expectedToNotContain string
	}{
		{
			name: "Single recent video",
			days: 30,
			mockService: &mockYouTubeService{
				playlistTitle: "Test Playlist",
				playlistItems: []*youtube.PlaylistItem{
					{
						Snippet: &youtube.PlaylistItemSnippet{
							Title:       "Test Video 1",
							Description: "This is a test video.",
							PublishedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
							ResourceId: &youtube.ResourceId{
								VideoId: "123",
							},
						},
					},
				},
			},
			expectedToContain: "- **[Test Video 1](https://www.youtube.com/watch?v=123)**",
		},
		{
			name: "Date filtering",
			days: 30,
			mockService: &mockYouTubeService{
				playlistTitle: "Test Playlist",
				playlistItems: []*youtube.PlaylistItem{
					{
						Snippet: &youtube.PlaylistItemSnippet{
							Title:       "Recent Video",
							PublishedAt: time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
							ResourceId:  &youtube.ResourceId{VideoId: "recent"},
						},
					},
					{
						Snippet: &youtube.PlaylistItemSnippet{
							Title:       "Old Video",
							PublishedAt: time.Now().Add(-45 * 24 * time.Hour).Format(time.RFC3339),
							ResourceId:  &youtube.ResourceId{VideoId: "old"},
						},
					},
				},
			},
			expectedToContain:    "Recent Video",
			expectedToNotContain: "Old Video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days = tt.days

			app := &App{
				youtubeService: tt.mockService,
			}

			// Redirect stdout to a buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			if err := app.run(); err != nil {
				t.Fatalf("run() failed: %v", err)
			}

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.expectedToContain) {
				t.Errorf("Expected output to contain %q, but it didn't. Got: %s", tt.expectedToContain, output)
			}

			if tt.expectedToNotContain != "" && strings.Contains(output, tt.expectedToNotContain) {
				t.Errorf("Expected output to not contain %q, but it did. Got: %s", tt.expectedToNotContain, output)
			}
		})
	}
}