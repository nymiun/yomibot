package main

import (
	"encoding/binary"
	"io"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/nemphi/sento"
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

func (a *agata) play(bot *sento.Bot, info sento.HandleInfo) error {
	vs, err := bot.Sess().State.VoiceState(info.GuildID, info.AuthorID)
	if err != nil {
		return err
	}
	retryCount := 0
retry:
	v, err := bot.Sess().ChannelVoiceJoin(info.GuildID, vs.ChannelID, false, true)
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

	var songReader io.ReadCloser

	killChan := make(chan struct{}, 1)

	url, err := url.Parse(info.MessageContent)

	videoUrl := "https://www.youtube.com/watch?v="
	if err == nil && (strings.Contains(url.Host, "youtube.com") || strings.Contains(url.Host, "youtu.be")) {
		songReader, err = a.playYoutube(bot, info, killChan)
		if err != nil {
			return err
		}
	} else {
		songReader, err = a.playYoutube(bot, info, killChan)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:", // Input
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

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		defer ffmpegInput.Close()
		_, err := io.Copy(ffmpegInput, songReader)
		if err != nil {
			bot.LogError(err.Error())
		}
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

	v.Speaking(true)

	bot.Send(info, "Playing "+videoUrl)
	for {
		if !v.Ready || v.OpusSend == nil {
			continue
		}
		// read pcm from chan, exit if channel is closed.
		recv := <-send
		if recv == nil {
			close(send)
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
	cmd.Process.Kill()
	killChan <- struct{}{}
	return nil
}
