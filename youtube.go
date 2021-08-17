package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/nemphi/sento"
)

func (a *agata) playYoutube(bot *sento.Bot, info sento.HandleInfo, url *url.URL, gs *guildState) (songInfo, *exec.Cmd, io.ReadCloser, error) {

	videoUrl := "https://www.youtube.com/watch?v="
	if url != nil {
		if strings.HasPrefix(url.Path, "/playlist") {
			ytRes, err := a.youtubePlaylist(url.Query().Get("list"), "")
			if err != nil {
				return songInfo{}, nil, nil, err
			}

			gs.Lock()
			for i := 0; i < len(ytRes.Items); i++ {
				if i == 0 {
					videoUrl += ytRes.Items[i].Snippet.ResourceID.VideoID
				} else {
					gs.queue.Add(sento.HandleInfo{
						Trigger:        info.Trigger,
						GuildID:        info.GuildID,
						ChannelID:      info.ChannelID,
						MessageID:      info.MessageID,
						AuthorID:       info.AuthorID,
						MessageContent: "https://www.youtube.com/watch?v=" + ytRes.Items[i].Snippet.ResourceID.VideoID,
					})
				}
			}
			gs.Unlock()
			bot.Send(info, fmt.Sprintf("Added %v songs", len(ytRes.Items)))
		} else {
			videoUrl += url.Query().Get("v")
		}

	} else {
		ytRes, err := a.youtubeSearch(info.MessageContent)
		if err != nil {
			return songInfo{}, nil, nil, err
		}
		if len(ytRes.Items) < 1 {
			bot.Send(info, "Bobo tonto no hay video pa ti")
			return songInfo{}, nil, nil, constError("no video")
		}
		videoUrl += ytRes.Items[0].ID.VideoID
	}
	ytdlCmd := exec.Command(
		"youtube-dl",
		videoUrl,
		"--http-chunk-size", "10M",
		"--no-playlist",
		"-q",
		"-o", "-",
	)
	ytdlReader, err := ytdlCmd.StdoutPipe()
	if err != nil {
		return songInfo{}, nil, nil, err
	}
	err = ytdlCmd.Start()
	if err != nil {
		return songInfo{}, nil, nil, err
	}
	return songInfo{url: videoUrl}, ytdlCmd, ytdlReader, nil
}

// Agata makes a Youtube search and returns the first result
func (a *agata) youtubeSearch(query string) (*youtubeResponse, error) {
	res, err := http.Get("https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&maxResults=1&q=" + url.QueryEscape(query) + "&key=" + a.youtubeKey)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	ytRes := &youtubeResponse{}
	err = json.NewDecoder(res.Body).Decode(ytRes)
	if err != nil {
		return nil, err
	}
	return ytRes, nil
}

func (a *agata) youtubePlaylist(id string, nextToken string) (*youtubePlaylistResponse, error) {
	res, err := http.Get("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=50&playlistId=" + url.QueryEscape(id) + "&pageToken=" + nextToken + "&key=" + a.youtubeKey)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	ytRes := &youtubePlaylistResponse{}
	err = json.NewDecoder(res.Body).Decode(ytRes)
	if err != nil {
		return nil, err
	}
	if ytRes.NextPageToken != "" {
		nextPage, err := a.youtubePlaylist(id, ytRes.NextPageToken)
		if err != nil {
			return nil, err
		}
		ytRes.Items = append(ytRes.Items, nextPage.Items...)
	}
	return ytRes, nil
}

type constError string

func (e constError) Error() string {
	return string(e)
}

type youtubeResponse struct {
	Kind          string `json:"kind,omitempty"`
	Etag          string `json:"etag,omitempty"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	RegionCode    string `json:"regionCode,omitempty"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults,omitempty"`
		ResultsPerPage int `json:"resultsPerPage,omitempty"`
	} `json:"pageInfo,omitempty"`
	Items []struct {
		Kind string `json:"kind,omitempty"`
		Etag string `json:"etag,omitempty"`
		ID   struct {
			Kind    string `json:"kind,omitempty"`
			VideoID string `json:"videoId,omitempty"`
		} `json:"id,omitempty"`
		Snippet struct {
			PublishedAt time.Time `json:"publishedAt,omitempty"`
			ChannelID   string    `json:"channelId,omitempty"`
			Title       string    `json:"title,omitempty"`
			Description string    `json:"description,omitempty"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url,omitempty"`
					Width  int    `json:"width,omitempty"`
					Height int    `json:"height,omitempty"`
				} `json:"default,omitempty"`
				Medium struct {
					URL    string `json:"url,omitempty"`
					Width  int    `json:"width,omitempty"`
					Height int    `json:"height,omitempty"`
				} `json:"medium,omitempty"`
				High struct {
					URL    string `json:"url,omitempty"`
					Width  int    `json:"width,omitempty"`
					Height int    `json:"height,omitempty"`
				} `json:"high,omitempty"`
			} `json:"thumbnails,omitempty"`
			ResourceID struct {
				Kind    string `json:"kind,omitempty"`
				VideoID string `json:"video_id,omitempty"`
			} `json:"resource_id,omitempty"`
			ChannelTitle         string    `json:"channelTitle,omitempty"`
			LiveBroadcastContent string    `json:"liveBroadcastContent,omitempty"`
			PublishTime          time.Time `json:"publishTime,omitempty"`
		} `json:"snippet,omitempty"`
	} `json:"items,omitempty"`
}

type youtubePlaylistResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken"`
	Items         []struct {
		Kind    string `json:"kind"`
		Etag    string `json:"etag"`
		ID      string `json:"id"`
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
				Standard struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"standard"`
			} `json:"thumbnails"`
			ChannelTitle string `json:"channelTitle"`
			PlaylistID   string `json:"playlistId"`
			Position     int    `json:"position"`
			ResourceID   struct {
				Kind    string `json:"kind"`
				VideoID string `json:"videoId"`
			} `json:"resourceId"`
			VideoOwnerChannelTitle string `json:"videoOwnerChannelTitle"`
			VideoOwnerChannelID    string `json:"videoOwnerChannelId"`
		} `json:"snippet,omitempty"`
	} `json:"items"`
	PageInfo struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
}
