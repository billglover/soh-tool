# YouTube Playlist Reporter

A command-line interface (CLI) that queries a YouTube playlist for content released within a specific number of days and returns it as a markdown formatted list.

## Installation

Pre-built binaries for macOS, Linux, and Windows are available on the [GitHub Releases page](https://github.com/billglover/soh-tool/releases). Download the appropriate binary for your operating system, unzip it, and place it in a directory that is in your system's `PATH`.

## Usage

To use the CLI, you first need to obtain a YouTube Data API key from the [Google Cloud Console](https://console.cloud.google.com/).

Once you have your API key, set it as an environment variable:

```bash
export YOUTUBE_API_KEY="your_api_key_here"
```

Then, you can run the CLI:

```bash
soh-tool [flags]
```

### Flags

- `-days int`: Specifies the number of days to look back for new videos. (default: `30`)

### Example

To get a report of all videos published in the last 60 days:

```bash
soh-tool -days=60
```

This will output a markdown-formatted list of videos, including their title, publication date, a link to the video, and a description.
