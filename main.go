package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tdakkota/gnhentai"
	"github.com/tdakkota/gnhentai/api"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var (
	Token              string
	Channel            string
	Client             *api.Client
	GetNumberRegex     *regexp.Regexp
	RemoveMentionRegex *regexp.Regexp
)

func main() {
	GetNumberRegex, _ = regexp.Compile("[0-9]{5,6}")
	RemoveMentionRegex, _ = regexp.Compile("<(@&?!?|#)[0-9]+?>")
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&Channel, "c", "", "Output Channel")
	flag.Parse()
	Client = api.NewClient()

	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		panic(err)
	}
	discord.AddHandler(messageCreate)
	discord.Identify.Intents = discordgo.IntentsGuildMessages
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Connected to Discord!")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	err = discord.Close()
	if err != nil {
		return
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// find doujin
	content := RemoveMentionRegex.ReplaceAllString(m.Content, "")
	match := GetNumberRegex.FindString(content)
	id, _ := strconv.ParseInt(match, 10, 32)
	doujin, err := Client.ByID(int(id))
	if err != nil {
		return
	}
	// extra things
	tags := ""
	for _, tag := range doujin.Tags {
		tags += fmt.Sprint("[", tag.Name, "(", tag.Count, ")]", "(", gnhentai.BaseNHentaiLink, tag.URL, ")", ", ")
	}
	tags = strings.TrimSuffix(tags, ", ")
	cover := gnhentai.CoverLink(doujin.MediaID, gnhentai.FormatFromImage(doujin.Images.Cover))
	_, err = s.ChannelMessageSendEmbed(Channel, &discordgo.MessageEmbed{
		Title: "Digits found: " + match,
		Image: &discordgo.MessageEmbedImage{
			URL: cover,
		},
		URL:         "https://nhentai.net/g/" + match,
		Description: tags + "\n\n" + fmt.Sprint("https://discord.com/channels/", m.GuildID, "/", m.ChannelID, "/", m.ID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "English",
				Value:  doujin.Title.English,
				Inline: true,
			},
			{
				Name:   "Japanese",
				Value:  doujin.Title.Japanese,
				Inline: true,
			},
			{
				Name:   "Pretty",
				Value:  doujin.Title.Pretty,
				Inline: true,
			},
			{
				Name:   "Pages",
				Value:  strconv.Itoa(doujin.NumPages),
				Inline: true,
			},
			{
				Name:   "Favorites",
				Value:  strconv.Itoa(doujin.NumFavorites),
				Inline: true,
			},
		},
	})
	if err != nil {
		return
	}
}
