package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"strings"
	"bufio"
	"log" //Delete later
	"os/user"
	"runtime"
	"net"
)
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
func discordBot(token string){
	bot, err := discordgo.New(fmt.Sprint("Bot ", token))

	if err != nil {
		log.Fatal(err)
	}

	// Start websocket connection
	err = bot.Open()
	if err != nil {
		log.Fatal(err)
	}
	// Create message handler
	bot.AddHandler(func(session *discordgo.Session, message *discordgo.MessageCreate) {
		
	})

	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	stop := make(chan bool)
	<-stop
}
func getExecutablePath() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return exePath
}
func getInfo() map[string]string {
	var personalData map[string]string = make(map[string]string)

	hostname, _ := os.Hostname()
	personalData["hostname"] = hostname

	user, _ := user.Current()
	personalData["username"] = user.Username

	ip := getLocalIP()
	personalData["ip"] = ip

	workingDir, _ := os.Getwd()
	personalData["workingDir"] = workingDir

	operatingSystem := runtime.GOOS
	personalData["os"] = operatingSystem

	
	return personalData
}
func getLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "Unknown"
    }
    for _, address := range addrs {
		ipnet, ok := address.(*net.IPNet)
        if ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return "Unknown"
}

func main(){
	envVars, err := loadEnvFile(".env")
	if err != nil {
		log.Fatal(err)
	}

	var TOKEN = envVars["DISCORD_TOKEN"]
	discordBot(TOKEN)
}