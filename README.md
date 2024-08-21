# Disgomoji

##### First malware of my series called "Recreating Malware"
##### This project is proof of concept and should not be used with unethical intentions

## Description
I decided to start my own series called "Recreating Malware", where i find interesting malware from around the internet and try to recreate it. For the first one i choose malware called [Disgomoji](https://www.volexity.com/blog/2024/06/13/disgomoji-malware-used-to-target-indian-government/)
***
This malware was created by thread actor under alias UTA0137 to target Indian goverment. It is a modified version of the public project [discord-c2](https://github.com/bmdyy/discord-c2), which uses the messaging service Discord for command and control (C2), making use of emojis for its C2 communication. Malware came in to the computer using phishing zip file, containing downloader script impersonating valid document, which downloaded Disgomoji with actual document which was showed to the user. Primarily it is targeted towards goverment entities of India, who use custom linux distribution named BOSS as their daily desktop. It was found that it utilizes DirtyPipe (CVE-2022-0847).
#### Quick Analysis
Malware comes in phishing zip file, after opening file inside it, it downloads golang writen payload vmcoreinfo[^1] (instance of Disgomoji, just renamed) and actual file which is showed to the user. The payload is dropped in a hidden folder named .x86_64-linux-gnu in the user’s home directory. Disgomoji then authorizes to discord server using hardcoded token and guild id, in which it creates separate channel for current user. The channel name format is sess-%s-%s, where the first %s value is the operating system of the infected machine, and the second %s is formatted using the victim’s username. 
On startup it sends check-in message containing:
- Hostname
- Username
- Internal IP
- Operating System
- Current working directory

Persistence is maintaned using cron with @reboot entry so it can survive reboots. It also downloads simple script named uevent_seqnum.sh and executes it. Scripts job is to check for any connected USB devices and if so, download all content from it to local computer, so they can be retrieved later.
#### Commands
Disgomoji listens to commands in its own dedicated channel. C2 communication is emoji based, where attacker controls payload by sending emojis into the chanell. While Disgomoji is processing command it react with "Clock" emoji (🕐). After command finishes clock emoji is deleted and "Check mark button" emoji (✅) is added as reaction.
Bellow are listed avaible command to which Disgomoji listens:
| Emoji | Command description |
|-------|------------------|
|  🏃‍♂️  | Execute a command on the victim’s device. This command receives an argument, which is the command to execute. |
|  📸  | Take a screenshot of the victim’s screen and upload it to the command channel as an attachment. |
|  👇  | Download files from the victim’s device and upload them to the command channel as attachments. This command receives one argument, which is the path of the file. |
|  ☝️  | Upload a file to the victim’s device. The file to upload is attached along with this emoji. |
|  👉  | Upload a file from the victim’s device to Oshi (oshi[.]at), a remote file-storage service. This command receives an argument, which is the name of the file to upload. |
|  👈  | Upload a file from the victim’s device to transfer[.]sh, a remote file-sharing service. This command receives an argument, which is the name of the file to upload. |
|  🔥  | Find and send all files matching a pre-defined extension list that are present on the victim’s device. Files with the following extensions are exfiltrated: CSV, DOC, ISO, JPG, ODP, ODS, ODT, PDF, PPT, RAR, SQL, TAR, XLS, ZIP |
|  🦊  | Zip all Firefox profiles on the victim’s device. These files can be retrieved by the attacker later. |
|  💀  | Terminate the malware process using os.Exit(). |

## Instalation
> run: go run main.go

> compile: go build main.go

## Testing
Right now project is tested on windows but later switch to linux, as original disgomoji was made for
Some parts are coppied from discord-c2 as this malware is inspired by
**Sadly i found out, that transfer.sh is currently not avaible, instead of it i will be using temp.sh**

- [x] Add sending startup message
- [x] Add command 🏃‍♂️
- [x] Add command 📸
- [x] Add command 👇
- [x] Add command ☝️
- [x] Add command 👉
- [x] Add command 👈
- [x] Add command 🔥
- [ ] Add command 🦊
- [x] Add command 💀
- [ ] Create DirtyPipe exploit
- [x] Create cron persistence
- [x] Create uevent_seqnum.sh
- [ ] Optimalization
- [ ] Comments

## Resources
- [Volexity.com](https://www.volexity.com/blog/2024/06/13/disgomoji-malware-used-to-target-indian-government)
- [TheHackerNews.com](https://thehackernews.com/2024/06/pakistani-hackers-use-disgomoji-malware.html)
- [Dirty Pipe](https://github.com/AlexisAhmed/CVE-2022-0847-DirtyPipe-Exploits)

[^1]: Instance of Disgomoji on Virustotal: https://www.virustotal.com/gui/file/d9f29a626857fa251393f056e454dfc02de53288ebe89a282bad38d03f614529
