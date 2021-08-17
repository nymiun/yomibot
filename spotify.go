package main

import (
	"encoding/base64"
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

func (a *agata) playSpotify(bot *sento.Bot, info sento.HandleInfo, url *url.URL, gs *guildState) (songInfo, *exec.Cmd, io.ReadCloser, error) {

	newInfo := sento.HandleInfo{
		Trigger:   info.Trigger,
		GuildID:   info.GuildID,
		ChannelID: info.ChannelID,
		MessageID: info.MessageID,
		AuthorID:  info.AuthorID,
	}
	if url != nil {
		if strings.HasPrefix(url.Path, "/album") {
			spotRes, err := a.spotifyAlbum(strings.TrimPrefix(url.Path, "/album/"), "")
			if err != nil {
				return songInfo{}, nil, nil, err
			}
			gs.Lock()
			for i := 0; i < len(spotRes.Items); i++ {
				if i != 0 {
					gs.queue.Add(sento.HandleInfo{
						Trigger:        info.Trigger,
						GuildID:        info.GuildID,
						ChannelID:      info.ChannelID,
						MessageID:      info.MessageID,
						AuthorID:       info.AuthorID,
						MessageContent: spotRes.Items[i].Name + " - " + spotRes.Items[i].Artists[0].Name + " lyrics",
					})
				}
			}
			gs.Unlock()
			bot.Send(info, fmt.Sprintf("Added %v songs", len(spotRes.Items)))
			newInfo.MessageContent = spotRes.Items[0].Name + " - " + spotRes.Items[0].Artists[0].Name + " lyrics"
		} else if strings.HasPrefix(url.Path, "/playlist") {
			spotRes, err := a.spotifyPlaylist(strings.TrimPrefix(url.Path, "/playlist/"), "")
			if err != nil {
				return songInfo{}, nil, nil, err
			}
			gs.Lock()
			for i := 0; i < len(spotRes.Items); i++ {
				if i != 0 {
					gs.queue.Add(sento.HandleInfo{
						Trigger:        info.Trigger,
						GuildID:        info.GuildID,
						ChannelID:      info.ChannelID,
						MessageID:      info.MessageID,
						AuthorID:       info.AuthorID,
						MessageContent: spotRes.Items[i].Track.Name + " - " + spotRes.Items[i].Track.Artists[0].Name + " lyrics",
					})
				}
			}
			gs.Unlock()
			bot.Send(info, fmt.Sprintf("Added %v songs", len(spotRes.Items)))
			newInfo.MessageContent = spotRes.Items[0].Track.Name + " - " + spotRes.Items[0].Track.Artists[0].Name + " lyrics"
		} else if strings.HasPrefix(url.Path, "/artist") {
			spotRes, err := a.spotifyArtist(strings.TrimPrefix(url.Path, "/artist/"))
			if err != nil {
				return songInfo{}, nil, nil, err
			}
			gs.Lock()
			for i := 0; i < len(spotRes.Tracks); i++ {
				if i != 0 {
					gs.queue.Add(sento.HandleInfo{
						Trigger:        info.Trigger,
						GuildID:        info.GuildID,
						ChannelID:      info.ChannelID,
						MessageID:      info.MessageID,
						AuthorID:       info.AuthorID,
						MessageContent: spotRes.Tracks[i].Name + " - " + spotRes.Tracks[i].Artists[0].Name + " lyrics",
					})
				}
			}
			gs.Unlock()
			bot.Send(info, fmt.Sprintf("Added %v songs", len(spotRes.Tracks)))
			newInfo.MessageContent = spotRes.Tracks[0].Name + " - " + spotRes.Tracks[0].Artists[0].Name + " lyrics"
		} else if strings.HasPrefix(url.Path, "/track") {
			songID := strings.TrimPrefix(url.Path, "/track/")
			spotRes, err := a.spotifyTrack(songID)
			if err != nil {
				return songInfo{}, nil, nil, err
			}
			newInfo.MessageContent = spotRes.Name + " - " + spotRes.Artists[0].Name + " lyrics"
		}
	}

	return a.playYoutube(bot, newInfo, nil, gs)
}

type spotifyLoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (a *agata) getSpotifyToken(bot *sento.Bot, isTick bool) error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		bot.LogError(err.Error())
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.spotifyClientID+":"+a.spotifyClientSecret)))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		bot.LogError(err.Error())
		return err
	}
	defer res.Body.Close()
	spotLog := &spotifyLoginResponse{}
	err = json.NewDecoder(res.Body).Decode(spotLog)
	if err != nil {
		bot.LogError(err.Error())
		return err
	}
	if !isTick {
		go func(a *agata, bot *sento.Bot, spotLog *spotifyLoginResponse) {
			tch := time.Tick(time.Duration(spotLog.ExpiresIn) * time.Second)
			for {
				select {
				case <-tch:
					a.getSpotifyToken(bot, true)
				}
			}
		}(a, bot, spotLog)
	}
	a.spotifyAccessToken = spotLog.AccessToken
	return nil
}

func (a *agata) spotifyTrack(id string) (*spotifyTrackResponse, error) {
	url := "https://api.spotify.com/v1/tracks/" + id
	res, err := a.spotifyRequest(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	spotRes := &spotifyTrackResponse{}
	err = json.NewDecoder(res.Body).Decode(spotRes)
	if err != nil {
		return nil, err
	}
	return spotRes, nil
}

func (a *agata) spotifyAlbum(id string, nextUrl string) (*spotifyAlbumTracksResponse, error) {
	url := "https://api.spotify.com/v1/albums/" + id + "/tracks?limit=50"
	if nextUrl != "" {
		url = nextUrl
	}
	res, err := a.spotifyRequest(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	spotRes := &spotifyAlbumTracksResponse{}
	err = json.NewDecoder(res.Body).Decode(spotRes)
	if err != nil {
		return nil, err
	}

	if spotRes.Next != "" {
		nextPage, err := a.spotifyAlbum(id, spotRes.Next)
		if err != nil {
			return nil, err
		}
		spotRes.Items = append(spotRes.Items, nextPage.Items...)
	}
	return spotRes, nil
}

func (a *agata) spotifyPlaylist(id string, nextUrl string) (*spotifyPlaylistResponse, error) {
	url := "https://api.spotify.com/v1/playlists/" + id + "/tracks?limit=100"
	if nextUrl != "" {
		url = nextUrl
	}
	res, err := a.spotifyRequest(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	spotRes := &spotifyPlaylistResponse{}
	err = json.NewDecoder(res.Body).Decode(spotRes)
	if err != nil {
		return nil, err
	}

	if spotRes.Next != "" {
		nextPage, err := a.spotifyPlaylist(id, spotRes.Next)
		if err != nil {
			return nil, err
		}
		spotRes.Items = append(spotRes.Items, nextPage.Items...)
	}
	return spotRes, nil
}

func (a *agata) spotifyArtist(id string) (*spotifyArtistResponse, error) {
	url := "https://api.spotify.com/v1/artists/" + id + "/top-tracks?market=US"
	res, err := a.spotifyRequest(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	spotRes := &spotifyArtistResponse{}
	err = json.NewDecoder(res.Body).Decode(spotRes)
	if err != nil {
		return nil, err
	}
	return spotRes, nil
}

func (a *agata) spotifyRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+a.spotifyAccessToken)
	return http.DefaultClient.Do(req)
}

type spotifyTrackResponse struct {
	Album struct {
		AlbumType string `json:"album_type"`
		Artists   []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		Type                 string `json:"type"`
		URI                  string `json:"uri"`
	} `json:"album"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []string `json:"available_markets"`
	DiscNumber       int      `json:"disc_number"`
	DurationMs       int      `json:"duration_ms"`
	Explicit         bool     `json:"explicit"`
	ExternalIds      struct {
		Isrc string `json:"isrc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href        string `json:"href"`
	ID          string `json:"id"`
	IsLocal     bool   `json:"is_local"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewURL  string `json:"preview_url"`
	TrackNumber int    `json:"track_number"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
}

type spotifyAlbumTracksResponse struct {
	Href  string `json:"href"`
	Items []struct {
		Artists []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		DiscNumber       int      `json:"disc_number"`
		DurationMs       int      `json:"duration_ms"`
		Explicit         bool     `json:"explicit"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		PreviewURL  string `json:"preview_url"`
		TrackNumber int    `json:"track_number"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
	} `json:"items"`
	Limit    int         `json:"limit"`
	Next     string      `json:"next"`
	Offset   int         `json:"offset"`
	Previous interface{} `json:"previous"`
	Total    int         `json:"total"`
}

type spotifyPlaylistResponse struct {
	Href  string `json:"href"`
	Items []struct {
		AddedAt time.Time `json:"added_at"`
		AddedBy struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"added_by"`
		IsLocal bool `json:"is_local"`
		Track   struct {
			Album struct {
				AlbumType string `json:"album_type"`
				Artists   []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					Height int    `json:"height"`
					URL    string `json:"url"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       int      `json:"disc_number"`
			DurationMs       int      `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string `json:"href"`
			ID          string `json:"id"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
		} `json:"track,omitempty"`
	} `json:"items"`
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
}

type spotifyArtistResponse struct {
	Tracks []struct {
		Album struct {
			AlbumType string `json:"album_type"`
			Artists   []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Width  int    `json:"width"`
			} `json:"images"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"album"`
		Artists []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		DiscNumber       int      `json:"disc_number"`
		DurationMs       int      `json:"duration_ms"`
		Explicit         bool     `json:"explicit"`
		ExternalIds      struct {
			Isrc string `json:"isrc"`
		} `json:"external_ids"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		Popularity  int    `json:"popularity"`
		PreviewURL  string `json:"preview_url"`
		TrackNumber int    `json:"track_number"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
	} `json:"tracks"`
}
