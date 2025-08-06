## 📺 {{.PlaylistTitle}} on YouTube

{{range .Videos}}
- **[{{.Title}}](https://www.youtube.com/watch?v={{.VideoID}})** - {{.PublishedAt.Format "2006-01-02"}}

{{.Description}}
{{end}}