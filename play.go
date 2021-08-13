package main

import (
	"encoding/binary"
	"io"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/nemphi/sento"
	"github.com/patrickmn/go-cache"
	"layeh.com/gopus"
)

// Technically the below settings can be adjusted however that poses
// a lot of other problems that are not handled well at this time.
// These below values seem to provide the best overall performance
const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

type songInfo struct {
	url string
}

func (a *agata) play(bot *sento.Bot, info sento.HandleInfo) error {
	vs, err := bot.Sess().State.VoiceState(info.GuildID, info.AuthorID)
	if err != nil {
		return err
	}
	retryCount := 0

	var v *discordgo.VoiceConnection
	var gs *guildState

	gsi, exists := a.guildMap.Get(info.GuildID)
	if !exists {
	retry:
		v, err = bot.Sess().ChannelVoiceJoin(info.GuildID, vs.ChannelID, false, true)
		if err != nil {
			bot.LogError(err.Error())
			if retryCount < 3 {
				retryCount++
				time.Sleep(time.Second)
				goto retry
			}
			// TODO: Send Err mesg to channel
			return err
		}
		gs = &guildState{
			voice:   v,
			stopper: make(chan struct{}, 1),
			pauser:  make(chan struct{}, 1),
			resumer: make(chan struct{}, 1),
			queue:   arraylist.New(),
		}
	} else {
		gs = gsi.(*guildState)
		gs.Lock()
		v = gs.voice
		gs.Unlock()
	}

	a.guildMap.Set(info.GuildID, gs, cache.DefaultExpiration)

	var songReader io.ReadCloser
	var fetcherCmd *exec.Cmd

	var song songInfo

	url, err := url.Parse(info.MessageContent)

	if err == nil && (strings.Contains(url.Host, "youtube.com") || strings.Contains(url.Host, "youtu.be")) {
		song, fetcherCmd, songReader, err = a.playYoutube(bot, info)
		if err != nil {
			return err
		}
	} else {
		song, fetcherCmd, songReader, err = a.playYoutube(bot, info)
		if err != nil {
			return err
		}
	}

	gs.Lock()
	if gs.playing {
		gs.queue.Add(info)
		gs.Unlock()
		return nil
	}
	gs.fetcherCmd = fetcherCmd
	gs.fetcherOut = songReader
	gs.Unlock()

	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:", // Input
		"-threads", "2",
		"-vn",                                  // No video
		"-af", "loudnorm=I=-16:LRA=11:TP=-1.5", // Normalization filter
		"-f", "s16le", // Format
		"-ar", "48K", // Samplerate
		"-b:a", "96K", // Bitrate
		"-ac", "2", // Audio Channels
		"pipe:", // Output
	)
	ffmpegInput, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	ffmpegOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	gs.Lock()
	gs.ffmpegCmd = cmd
	gs.Unlock()
	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		_, err := io.Copy(ffmpegInput, songReader)
		if err != nil && !strings.Contains(err.Error(), "broken pipe") {
			bot.LogError(err.Error())
		}
		gs.Lock()
		if !gs.stopping {
			ffmpegInput.Close()
		}
		gs.Unlock()
	}()

	send := make(chan []int16, 2)
	go func() {
		for {
			// read data from ffmpeg stdout
			audiobuf := make([]int16, frameSize*channels)
			err = binary.Read(ffmpegOut, binary.LittleEndian, &audiobuf)
			if err == io.EOF {
				send <- nil
				ffmpegOut.Close()
				return
			}
			if err == io.ErrUnexpectedEOF {
				send <- audiobuf
				continue
			}
			if err != nil {
				send <- nil
				bot.LogError(err.Error())
				return
			}
			send <- audiobuf
		}
	}()

	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		bot.LogError(err.Error())
	}
	opusEncoder.SetBitrate(96000)
	gs.Lock()
	gs.encoder = opusEncoder
	gs.playing = true
	gs.Unlock()

	v.Speaking(true)

	bot.Send(info, "Playing "+song.url)
	for {
		if !v.Ready || v.OpusSend == nil {
			continue
		}
		select {

		// read pcm from chan, exit if channel is closed.
		case recv := <-send:
			if recv == nil {
				close(send)
				goto yes
			}
			// try encoding pcm frame with Opus
			opus, err := opusEncoder.Encode(recv, frameSize, maxBytes)
			if err != nil {
				bot.LogError(err.Error())
				return err
			}
			v.OpusSend <- opus
		case <-gs.pauser:
			gs.Lock()
			gs.paused = true
			gs.Unlock()
			<-gs.resumer
		case <-gs.stopper:
			goto yes
		}
	}
yes:
	gs.Lock()
	if gs.looping {
		defer a.play(bot, info)
	}
	if !gs.queue.Empty() && !gs.looping {
		nxtInfo, exists := gs.queue.Get(0)
		if exists {
			defer a.play(bot, nxtInfo.(sento.HandleInfo))
		}
		gs.queue.Remove(0)
	}
	gs.playing = false
	v.Speaking(false)
	fetcherCmd.Wait()
	cmd.Wait()
	if gs.leaving {
		v.Disconnect()
	}
	gs.Unlock()
	return nil
}
