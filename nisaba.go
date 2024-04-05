package main

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "github.com/thoj/go-ircevent"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
    "bufio"
    "regexp"
)

type Config struct {
    Nickname    string `json:"nickname"`
    Server      string `json:"server"`
    Port        string `json:"port"`
    UseSSL      bool   `json:"use_ssl"`
    ValidateSSL bool   `json:"validate_ssl"`
    APIURL      string `json:"api_url"`
    APIKey      string `json:"api_key"`
    Channel     string `json:"channel"`
    MaxMessageSize int `json:"max_message_size"`
}

type Bot struct {
    Config        Config
    IRCConnection *irc.Connection
    IsAvailable   bool
}

func loadConfig() Config {
    var config Config
    file, err := os.Open("config.json")
    if err != nil {
        log.Fatalf("Error opening config file: %v", err)
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
        log.Fatalf("Error decoding config file: %v", err)
    }

    return config
}

var blockedUsers map[string]bool

func loadBlockedUsers() {
    blockedUsers = make(map[string]bool)
    fileName := "blocklist.txt"
    file, err := os.Open(fileName)
    if err != nil {
        if !os.IsNotExist(err) {
            log.Printf("Error opening block list file: %v", err)
        }
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        blockedUsers[strings.TrimSpace(scanner.Text())] = true
    }

    if err := scanner.Err(); err != nil {
        log.Printf("Error reading block list file: %v", err)
    }
}

func loadSystemPrompt() string {
    fileName := "systemprompt.txt"
    content, err := ioutil.ReadFile(fileName)
    if err != nil {
        if os.IsNotExist(err) {
            return ""
        }
        log.Fatalf("Error reading system prompt file: %v", err)
    }
    return string(content)
}

func NewBot(config Config) *Bot {
    bot := &Bot{
        Config:      config,
        IsAvailable: true,
    }
    irccon := irc.IRC(config.Nickname, config.Nickname)
    irccon.VerboseCallbackHandler = true
    irccon.Debug = true
    irccon.UseTLS = config.UseSSL
    if config.UseSSL {
        irccon.TLSConfig = &tls.Config{
            InsecureSkipVerify: !config.ValidateSSL,
        }
        if config.ValidateSSL {
            irccon.TLSConfig.ServerName = config.Server
        }
    }

    irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(config.Channel) })
    irccon.AddCallback("PRIVMSG", bot.handleMessage)

    bot.IRCConnection = irccon
    return bot
}

func (bot *Bot) callAPI(query string) string {
    systemPrompt := loadSystemPrompt()

    payload := map[string]interface{}{
        "prompt": query,
        "stream": false,
    }

    if systemPrompt != "" {
        payload["system_prompt"] = systemPrompt
    }

    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error encoding payload to JSON: %v", err)
        return "Error encoding request payload."
    }

    log.Printf("Sending payload: %s\n", string(payloadBytes))

    req, err := http.NewRequest("POST", bot.Config.APIURL, bytes.NewBuffer(payloadBytes))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer " + bot.Config.APIKey)

    if err != nil {
        log.Printf("Error creating request: %v", err)
        return "Error creating request."
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error making request to API: %v", err)
        return "Error sending request."
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response body: %v", err)
        return "Error reading response."
    }

    log.Printf("Received response: %s\n", string(body))

    var response struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    if err := json.Unmarshal(body, &response); err != nil {
        log.Printf("Error decoding response from API: %v", err)
        return "Error parsing response."
    }

    if len(response.Choices) > 0 && response.Choices[0].Message.Content != "" {
        return response.Choices[0].Message.Content
    } else {
        log.Printf("API response does not contain expected message structure or is empty.")
        return "I'm sorry, I am having some trouble accessing the archives."
    }
}

func (bot *Bot) handleMessage(e *irc.Event) {
    if !bot.IsAvailable || blockedUsers[e.Nick] {
        return
    }

    message := e.Message()

    re := regexp.MustCompile(`(?i)^(` + regexp.QuoteMeta(bot.Config.Nickname) + `[:,]?\s?)`)
    query := re.ReplaceAllString(strings.TrimSpace(message), "")

    if strings.HasPrefix(strings.ToLower(message), strings.ToLower(bot.Config.Nickname)) {
        bot.IsAvailable = false
        user := e.Nick

        bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: I will consult my knowledge and be back with you shortly.", user))

        go func() {
            response := bot.callAPI(query)
            messages := splitMessage(response, bot.Config.MaxMessageSize)

            for i, msg := range messages {
                if i == 0 {
                    bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: %s", user, msg))
                } else {
                    bot.IRCConnection.Privmsg(bot.Config.Channel, msg)
                }
                time.Sleep(1 * time.Second)
            }

            bot.IsAvailable = true
        }()
    }
}

func splitMessage(response string, maxSize int) []string {
    var parts []string
    var currentSize int
    var currentPart bytes.Buffer

    for _, runeValue := range response {
        if currentSize+len(string(runeValue)) > maxSize || runeValue == '\n' {
            parts = append(parts, currentPart.String())
            currentPart.Reset()
            currentSize = 0
        }
        if runeValue != '\n' {
            currentPart.WriteRune(runeValue)
            currentSize += len(string(runeValue))
        }
    }

    if currentPart.Len() > 0 {
        parts = append(parts, currentPart.String())
    }

    return parts
}

func main() {
    loadBlockedUsers()
    config := loadConfig()
    bot := NewBot(config)

    if err := bot.IRCConnection.Connect(fmt.Sprintf("%s:%s", config.Server, config.Port)); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	bot.IRCConnection.Loop()
}
