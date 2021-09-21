package discord

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ImDevinC/go-radio/pkg/jsrlive"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

const prefix = "!jsrlive"

type Client struct {
	sess  *discordgo.Session
	token string
}

var voiceChannels = make(map[string]*discordgo.VoiceConnection)

func NewClient(token string) (*Client, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	client := Client{
		sess:  discord,
		token: token,
	}
	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)
	return &client, nil
}

func (c *Client) Run() {
	err := c.sess.Open()
	if err != nil {
		panic(err)
	}
	fmt.Println("Radio is live...")
	lock := make(chan os.Signal, 1)
	signal.Notify(lock, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-lock
	c.sess.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, prefix)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Received message")
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	var err error
	if strings.HasPrefix(m.Content, fmt.Sprintf("%s join ", prefix)) {
		err = joinVoiceChannel(s, m)
	} else if strings.HasPrefix(m.Content, fmt.Sprintf("%s station ", prefix)) {
		err = startRadio(s, m)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func joinVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) error {
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return err
	}
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		return err
	}
	channelName := strings.TrimPrefix(m.Content, prefix+" join ")
	var channelID string
	for _, ch := range g.Channels {
		if ch.Type != discordgo.ChannelTypeGuildVoice && strings.ToLower(ch.Name) != channelName {
			continue
		}
		channelID = ch.ID
	}
	if channelID == "" {
		return errors.New("failed to find channel")
	}
	vc, err := s.ChannelVoiceJoin(c.GuildID, channelID, false, true)
	if err != nil {
		return err
	}
	voiceChannels[c.GuildID] = vc
	return nil
}

func startRadio(s *discordgo.Session, m *discordgo.MessageCreate) error {
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return err
	}
	vc := voiceChannels[c.GuildID]
	if vc == nil {
		return errors.New("not joined to a voice channel")
	}
	songs, err := jsrlive.GetSongs("ultraremixes")
	if err != nil {
		return err
	}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"
	songURL := jsrlive.FormatSong("ultraremixes", songs[0])

	encodingSession, err := dca.EncodeFile(songURL, options)
	if err != nil {
		return err
	}
	defer encodingSession.Cleanup()

	vc.Speaking(true)
	defer vc.Speaking(false)

	d := make(chan error)
	stream := dca.NewStream(encodingSession, vc, d)
	for {
		err = <-d
		if err != nil && err != io.EOF {
			return err
		}
		frame, err := encodingSession.OpusFrame()
		if err != nil {
			return err
		}
		vc.OpusSend <- frame
		finished, err := stream.Finished()
		if err != nil {
			return err
		}
		if err == io.EOF || finished {
			break
		}
	}
	return nil
}
