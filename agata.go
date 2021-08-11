package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/nemphi/sento"
	"layeh.com/gopus"
)

type agata struct {
	youtubeKey string
}

func (a *agata) Start(_ *sento.Bot) (err error) { return }
func (a *agata) Stop(_ *sento.Bot) (err error)  { return }

func (a *agata) Name() string {
	return "Agata"
}

func (a *agata) Triggers() []string {
	return []string{
		"p",
		"play",
		// "s",
		// "stop",
		// "pause",
		// "resume",
		// "skip",
		// "next",
		// "seek",
		// "queue",
		// "q",
		// "leave",
		// "history",
		// "nowplaying",
		// "loop",
		// "speed",
		// "volume",
	}
}

func (a *agata) Handle(bot *sento.Bot, info sento.HandleInfo) (err error) {
	switch info.Trigger {
	case "p", "play":
		ch, err := bot.Sess().Channel(info.ChannelID)
		if err != nil {
			return err
		}
		g, err := bot.Sess().Guild(ch.GuildID)
		if err != nil {
			return err
		}

		vs, err := bot.Sess().State.VoiceState(g.ID, info.AuthorID)
		if err != nil {
			return err
		}

		v, err := bot.Sess().ChannelVoiceJoin(g.ID, vs.ChannelID, false, true)
		if err != nil {
			bot.LogError(err.Error())
			return err
		}

		ytRes, err := a.youtubeSearch(info.MessageContent)
		if err != nil {
			return err
		}
		client := youtube.Client{}
		video, err := client.GetVideo(ytRes.Items[0].ID.VideoID)
		if err != nil {
			return err
		}
		formats := video.Formats.AudioChannels(2)
		formats.Sort()
		stream, _, err := client.GetStream(video, &formats[0])
		if err != nil {
			return err
		}
		cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-vn", "-f", "s16le", "-ar", "48K", "-b:a", "96K", "-ac", "2", "pipe:1")
		cmd.Stdin = stream
		ffmpegOut, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		ffmpegbuf := bufio.NewReaderSize(ffmpegOut, 16384)
		err = cmd.Start()
		if err != nil {
			return err
		}
		defer cmd.Process.Kill()

		// Technically the below settings can be adjusted however that poses
		// a lot of other problems that are not handled well at this time.
		// These below values seem to provide the best overall performance
		const (
			channels  int = 2                   // 1 for mono, 2 for stereo
			frameRate int = 48000               // audio sampling rate
			frameSize int = 960                 // uint16 size of each audio frame
			maxBytes  int = (frameSize * 2) * 2 // max size of opus data
		)

		send := make(chan []int16, 2)
		go func() {
			for {
				// read data from ffmpeg stdout
				audiobuf := make([]int16, frameSize*channels)
				err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
				send <- audiobuf
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					close(send)
					return
				}
				if err != nil {
					close(send)
					bot.LogError(err.Error())
					return
				}
			}
		}()

		opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
		if err != nil {
			return err
		}

		v.Speaking(true)

		bot.Send(info, "Playing "+ytRes.Items[0].Snippet.Title)
		for {
			if !v.Ready || v.OpusSend == nil {
				continue
			}
			// read pcm from chan, exit if channel is closed.
			recv, ok := <-send
			if !ok {
				bot.LogInfo("SEND CLOSED")
				break
			}
			// try encoding pcm frame with Opus
			opus, err := opusEncoder.Encode(recv, frameSize, maxBytes)
			if err != nil {
				bot.LogError(err.Error())
				return err
			}
			v.OpusSend <- opus
		}

		v.Speaking(false)
		defer v.Disconnect()
	default:
		return
	}

	return
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
