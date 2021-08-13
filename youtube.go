package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/nemphi/sento"
)

func (a *agata) playYoutube(bot *sento.Bot, info sento.HandleInfo, kill chan struct{}) (io.ReadCloser, error) {
	url, err := url.Parse(info.MessageContent)

	videoUrl := "https://www.youtube.com/watch?v="
	if err == nil && (strings.Contains(url.Host, "youtube.com") || strings.Contains(url.Host, "youtu.be")) {
		id, err := extractVideoID(info.MessageContent)
		if err != nil {
			return nil, err
		}
		videoUrl += id

	} else {
		ytRes, err := a.youtubeSearch(info.MessageContent)
		if err != nil {
			return nil, err
		}
		if len(ytRes.Items) < 1 {
			bot.Send(info, "Bobo tonto no hay video pa ti")
			return nil, constError("no video")
		}
		videoUrl += ytRes.Items[0].ID.VideoID
	}
	ytdlCmd := exec.Command("youtube-dl", videoUrl, "--no-playlist", "-q", "-o", "-")
	ytdlReader, err := ytdlCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = ytdlCmd.Start()
	if err != nil {
		return nil, err
	}
	go func(ytdlCmd *exec.Cmd) {
		<-kill
		ytdlCmd.Process.Kill()
	}(ytdlCmd)
	return ytdlReader, nil
}

// Agata makes a Youtube search and returns the first result
func (a *agata) youtubeSearch(query string) (*YoutubeResponse, error) {
	res, err := http.Get("https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=1&q=" + url.QueryEscape(query) + "&key=" + a.youtubeKey)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	ytRes := &YoutubeResponse{}
	err = json.NewDecoder(res.Body).Decode(ytRes)
	if err != nil {
		return nil, err
	}
	return ytRes, nil
}

var videoRegexpList = []*regexp.Regexp{
	regexp.MustCompile(`(?:v|embed|shorts|watch\?v)(?:=|/)([^"&?/=%]{11})`),
	regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
	regexp.MustCompile(`([^"&?/=%]{11})`),
}

type constError string

func (e constError) Error() string {
	return string(e)
}

const (
	ErrInvalidCharactersInVideoID = constError("invalid characters in video id")
	ErrVideoIDMinLength           = constError("the video id must be at least 10 characters long")
)

func extractVideoID(videoID string) (string, error) {
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		for _, re := range videoRegexpList {
			if isMatch := re.MatchString(videoID); isMatch {
				subs := re.FindStringSubmatch(videoID)
				videoID = subs[1]
			}
		}
	}

	if strings.ContainsAny(videoID, "?&/<%=") {
		return "", ErrInvalidCharactersInVideoID
	}
	if len(videoID) < 10 {
		return "", ErrVideoIDMinLength
	}

	return videoID, nil
}

type YoutubeResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken"`
	RegionCode    string `json:"regionCode"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []struct {
		Kind string `json:"kind"`
		Etag string `json:"etag"`
		ID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			PublishedAt time.Time `json:"publishedAt"`
			ChannelID   string    `json:"channelId"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
			} `json:"thumbnails"`
			ChannelTitle         string    `json:"channelTitle"`
			LiveBroadcastContent string    `json:"liveBroadcastContent"`
			PublishTime          time.Time `json:"publishTime"`
		} `json:"snippet"`
	} `json:"items"`
}
