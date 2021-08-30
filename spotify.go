package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nemphi/lavago"
	"github.com/nemphi/sento"
	"github.com/valyala/fasthttp"
)

func (a *agata) spotify(bot *sento.Bot, info sento.HandleInfo, gs *guildState, url *url.URL, lnode *lavago.Node, p *lavago.Player) (track *lavago.Track, err error) {
	if strings.HasPrefix(url.Path, "/album") {
		var spotRes *spotifyAlbumTracksResponse
		spotRes, err = a.spotifyAlbum(strings.TrimPrefix(url.Path, "/album/"), "")
		if err != nil {
			return nil, err
		}
		gs.Lock()
		for i := 0; i < len(spotRes.Items); i++ {
			if i != 0 {
				gs.queue.Add(rawTrack{
					songID: spotRes.Items[i].ID,
					title:  spotRes.Items[i].Name,
					artist: spotRes.Items[i].Artists[0].Name,
					url:    spotRes.Items[i].ExternalUrls.Spotify,
				})
			}
		}
		gs.Unlock()
		bot.Send(info, fmt.Sprintf("Added %v songs", len(spotRes.Items)))
		ci := &cacheItem{}
		a.db.Where(&cacheItem{SpotifyID: spotRes.Items[0].ID}).First(&ci)
		if ci.Track != "" {
			track = &lavago.Track{Track: ci.Track, Info: ci.Info}
			return
		}
		track, err = a.nodeSearchTrack(spotRes.Items[0].ID, spotRes.Items[0].Name, spotRes.Items[0].Artists[0].Name, spotRes.Items[0].ExternalUrls.Spotify)
	} else if strings.HasPrefix(url.Path, "/playlist") {
		var spotRes *spotifyPlaylistResponse
		spotRes, err = a.spotifyPlaylist(strings.TrimPrefix(url.Path, "/playlist/"), "")
		if err != nil {
			return nil, err
		}
		gs.Lock()
		added := false
		first := 0
		for i := 0; i < len(spotRes.Items); i++ {
			if spotRes.Items[i].Track.Name != "" {
				if !added {
					added = true
					first = i
				} else {
					gs.queue.Add(rawTrack{
						songID: spotRes.Items[i].Track.ID,
						title:  spotRes.Items[i].Track.Name,
						artist: spotRes.Items[i].Track.Artists[0].Name,
						url:    spotRes.Items[i].Track.ExternalUrls.Spotify,
					})
				}
			}
		}
		gs.Unlock()
		bot.Send(info, fmt.Sprintf("Added %v songs", len(spotRes.Items)))
		ci := &cacheItem{}
		a.db.Where(&cacheItem{SpotifyID: spotRes.Items[first].Track.ID}).First(&ci)
		if ci.Track != "" {
			track = &lavago.Track{Track: ci.Track, Info: ci.Info}
			return
		}
		track, err = a.nodeSearchTrack(spotRes.Items[first].Track.ID, spotRes.Items[first].Track.Name, spotRes.Items[first].Track.Artists[0].Name, spotRes.Items[first].Track.ExternalUrls.Spotify)
	} else if strings.HasPrefix(url.Path, "/artist") {
		var spotRes *spotifyArtistResponse
		spotRes, err = a.spotifyArtist(strings.TrimPrefix(url.Path, "/artist/"))
		if err != nil {
			return nil, err
		}
		gs.Lock()
		for i := 0; i < len(spotRes.Tracks); i++ {
			if i != 0 {
				gs.queue.Add(rawTrack{
					songID: spotRes.Tracks[i].ID,
					title:  spotRes.Tracks[i].Name,
					artist: spotRes.Tracks[i].Artists[0].Name,
					url:    spotRes.Tracks[i].ExternalUrls.Spotify,
				})
			}
		}
		gs.Unlock()
		bot.Send(info, fmt.Sprintf("Added %v songs", len(spotRes.Tracks)))
		ci := &cacheItem{}
		a.db.Where(&cacheItem{SpotifyID: spotRes.Tracks[0].ID}).First(&ci)
		if ci.Track != "" {
			track = &lavago.Track{Track: ci.Track, Info: ci.Info}
			return
		}
		track, err = a.nodeSearchTrack(spotRes.Tracks[0].ID, spotRes.Tracks[0].Name, spotRes.Tracks[0].Artists[0].Name, spotRes.Tracks[0].ExternalUrls.Spotify)
	} else if strings.HasPrefix(url.Path, "/track") {
		songID := strings.TrimPrefix(url.Path, "/track/")
		ci := &cacheItem{}
		a.db.Where(&cacheItem{SpotifyID: songID}).First(&ci)
		if ci.Track != "" {
			track = &lavago.Track{Track: ci.Track, Info: ci.Info}
			return
		}
		var spotRes *spotifyTrackResponse
		spotRes, err = a.spotifyTrack(songID)
		if err != nil {
			return nil, err
		}

		track, err = a.nodeSearchTrack(songID, spotRes.Name, spotRes.Artists[0].Name, spotRes.ExternalUrls.Spotify)
	}

	return track, err
}

func (a *agata) nodeSearchTrack(songID, title, artist, url string) (*lavago.Track, error) {
	sr, err := a.nodeSearchYTMusic(title, artist)
	if err != nil {
		return nil, err
	}
	switch sr.Status {
	case lavago.NoMatchesSearchStatus:
		sr2, err := a.nodeSearchYT(title, artist)
		if err != nil {
			return nil, err
		}
		switch sr2.Status {
		case lavago.NoMatchesSearchStatus:
			return nil, nil
		case lavago.SearchResultSearchStatus:
			track := sr2.Tracks[0]
			track.Info.Title = title
			track.Info.Author = artist
			track.Info.URL = url
			go a.db.Omit("ID", "Timestamp").Create(&cacheItem{
				SpotifyID: songID,
				Track:     track.Track,
				Info:      track.Info,
			})
			return track, nil
		default:
			return nil, nil
		}
	case lavago.SearchResultSearchStatus:
		track := sr.Tracks[0]
		track.Info.Title = title
		track.Info.Author = artist
		track.Info.URL = url
		go a.db.Omit("ID", "Timestamp").Create(&cacheItem{
			SpotifyID: songID,
			Track:     track.Track,
			Info:      track.Info,
		})
		return track, nil
	default:
		return nil, nil
	}
}

func (a *agata) nodeSearchYTMusic(title, artist string) (*lavago.SearchResult, error) {
	return a.lavaNode.Search(lavago.YouTubeMusic, title+" "+artist)
}

func (a *agata) nodeSearchYT(title, artist string) (*lavago.SearchResult, error) {
	return a.lavaNode.Search(lavago.YouTube, title+" "+artist+" audio")
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
	spotRes := &spotifyTrackResponse{}
	err = json.Unmarshal(res, spotRes)
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
	spotRes := &spotifyAlbumTracksResponse{}
	err = json.Unmarshal(res, spotRes)
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
	spotRes := &spotifyPlaylistResponse{}
	err = json.Unmarshal(res, spotRes)
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
	spotRes := &spotifyArtistResponse{}
	err = json.Unmarshal(res, spotRes)
	if err != nil {
		return nil, err
	}
	return spotRes, nil
}

func (a *agata) spotifyRequest(url string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.Add("Authorization", "Bearer "+a.spotifyAccessToken)
	res := fasthttp.AcquireResponse()
	err := fasthttp.Do(req, res) // TOO SLOWWWWWWWWWWWW
	if err != nil {
		return nil, err
	}
	data := res.Body()
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	return data, err
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
