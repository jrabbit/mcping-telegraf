package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"github.com/whatupdave/mcping"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const version = "0.3.0"

type McPopList struct {
	Online int
	Users  []string
}

func (m McPopList) String() string {
	return fmt.Sprintf("%v", m.Online)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "$minecraft" {
		resp := DoPing()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(":bar_chart:  mc population: %v \n :people_wrestling: %v", resp.Online, resp.Users))
	}
}

func DiscordSetup() {
	const TARGET_PERMS uint = 67120144
	//manage messages, channels, nicknames, view channels
	cl_id := viper.GetString("discord.client_id")
	log.Printf("connect via https://discordapp.com/oauth2/authorize?client_id=%s&scope=bot&permissions=%s", cl_id, TARGET_PERMS)
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	//done := make(chan bool)
	mc_voice_channel := GetVCChannel(s)
	for {
		select {
		case <-ticker.C:
			pop := DoPing()
			s.UpdateStatus(0, fmt.Sprintf("mc pop: %v", pop.Online))
			s.ChannelEdit(mc_voice_channel.ID, fmt.Sprintf("minecraft population: %v", pop.Online))
		}
	}
}

func GetVCChannel(s *discordgo.Session) discordgo.Channel {
	guilds, _ := s.UserGuilds(10, "", "")
	guild, _ := s.Guild(guilds[0].ID)
	for _, c := range guild.Channels {
		if strings.HasPrefix(c.Name, "minecraft") {
			return *c
		}
	}
	var category discordgo.Channel
	for _, c := range guild.Channels {
		if strings.HasPrefix(c.Name, "Voice"){
			category = *c		
		}
	}  
	newChannel, err := s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{Name: "minecraft population: x", Type: discordgo.ChannelTypeGuildVoice, ParentID: category.ID})
	if err != nil {
		log.Fatalf("issue creating channel %v", err)
	}
	return *newChannel
}

func DoDiscord() {
	token := viper.GetString("discord.token")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("discord error: %v", err)
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(ready)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func main() {
	fmt.Printf("mcping-bin version %s\n", version)
	// ref https://github.com/pallets/click/blob/4da5e93cede17262424671208799bc6921dcfa36/click/utils.py#L368-L417
	viper.AddConfigPath("$HOME/.config/")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/Library/Application Support/")
	viper.AddConfigPath(os.Getenv("MCPING_CONF_DIR"))
	cwd, _ := os.Getwd()
	viper.AddConfigPath(cwd)
	if runtime.GOOS == "windows" {
		winAppData := os.Getenv("LOCALAPPDATA")
		viper.AddConfigPath(winAppData)
	}
	viper.SetConfigName("mcping")

	viper.SetDefault("minecraft_server", "localhost:25565")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("config warn: %v", err)
	}
	log.Printf("mcping config file: %s", viper.ConfigFileUsed())
	DoDiscord()

	//resp := DoPing()
	//log.Println("Mineplex has", resp.Online, "players online")
}

func DoPing() McPopList {
	mcServer := viper.GetString("minecraft_server")
	resp, mcErr := mcping.Ping(mcServer)
	if mcErr != nil {
		log.Printf("minecraft fail: %s", mcErr)
		log.Printf("minecraft host tried: %s", mcServer)
	}
	users := []string{}
	for _, u := range resp.Sample {
		users = append(users, u.Name)
	}
	return McPopList{Online: resp.Online, Users: users}
}
