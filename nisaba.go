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
    "path/filepath"
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
    UseContext  bool   `json:"use_context"`
    APIURL      string `json:"api_url"`
    APIKey      string `json:"api_key"`
    APIMode     string `json:"api_mode"`
    Channel     string `json:"channel"`
    MaxMessageSize int `json:"max_message_size"`
}

type Bot struct {
    Config        Config
    IRCConnection *irc.Connection
    IsAvailable   bool
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

var blockedUsers map[string]bool

func getConfigFilePath(fileName string) string {
    configDir := "config"
    configPath := filepath.Join(configDir, fileName)
    if _, err := os.Stat(configPath); err == nil {
        return configPath
    }
    return fileName
}

func loadConfig() Config {
    var config Config
    configPath := getConfigFilePath("config.json")
    file, err := os.Open(configPath)
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

func loadBlockedUsers() {
    blockedUsers = make(map[string]bool)
    filePath := getConfigFilePath("blocklist.txt")
    file, err := os.Open(filePath)
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
    filePath := getConfigFilePath("systemprompt.txt")
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return ""
        }
        log.Fatalf("Error reading system prompt file: %v", err)
    }
    return string(content)
}

func loadMessageHistory() []Message {
    var history []Message
    filePath := getConfigFilePath("history.txt")

    _, err := os.Stat(filePath)
    if os.IsNotExist(err) {
        systemPromptContent := loadSystemPrompt()
        if systemPromptContent != "" {
            initialSystemMessage := Message{Role: "system", Content: systemPromptContent}
            history = append(history, initialSystemMessage)
            saveMessageHistory(history)
        }
        return history
    }

    file, err := ioutil.ReadFile(filePath)
    if err != nil {
        log.Fatalf("Error reading message history: %v", err)
    }
    if err := json.Unmarshal(file, &history); err != nil {
        log.Fatalf("Error parsing message history: %v", err)
    }
    return history
}

func saveMessageHistory(history []Message) {
    filePath := getConfigFilePath("history.txt")
    file, err := json.MarshalIndent(history, "", "  ")
    if err != nil {
        log.Fatalf("Error encoding message history: %v", err)
    }
    if err := ioutil.WriteFile(filePath, file, 0644); err != nil {
        log.Fatalf("Error writing message history: %v", err)
    }
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

func (bot *Bot) callAPI(query string, useContext bool) string {
    var responseContent string

    if bot.Config.APIMode == "query" {
        systemPrompt := loadSystemPrompt()
        payload := map[string]interface{}{
            "prompt":      query,
            "stream":      false,
            "use_context": useContext,
        }
        if systemPrompt != "" {
            payload["system_prompt"] = systemPrompt
        }
    } else if bot.Config.APIMode == "chat" {
        newUserMessage := Message{Role: "user", Content: query}
        history := loadMessageHistory()
        history = append(history, newUserMessage)
        saveMessageHistory(history)

        messagesPayload := make([]map[string]interface{}, len(history))
        for i, msg := range history {
            messagesPayload[i] = map[string]interface{}{"role": msg.Role, "content": msg.Content}
        }

        payload := map[string]interface{}{
            "messages": messagesPayload,
            "stream":   false,
            "use_context": useContext,
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
            responseContent = response.Choices[0].Message.Content

            history = append(history, Message{Role: "assistant", Content: responseContent})
            saveMessageHistory(history)
            return responseContent
        } else {
            log.Printf("API response does not contain expected message structure or is empty.")
            return "I'm sorry, I am having some trouble accessing the archives."
        }
    }

    return responseContent
}

func (bot *Bot) handleMessage(e *irc.Event) {
    if !bot.IsAvailable || blockedUsers[e.Nick] {
        return
    }

    message := e.Message()

    re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(bot.Config.Nickname) + `[:,]?\s?(!\w+)?\s*(.*)`)
    matches := re.FindStringSubmatch(strings.TrimSpace(message))

    if len(matches) > 0 {
        bot.IsAvailable = false
        user := e.Nick

        command := matches[1]
        query := matches[2]

        useContext := bot.Config.UseContext

        switch command {
        case "!search":
            useContext = true
            bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: I will search through the archives and see what I can find.", user))
            go func() {
                response := bot.callAPI(query, useContext)
                bot.sendMessage(user, response)
                bot.IsAvailable = true
            }()
        case "!clear":
            err := os.Remove("history.txt")
            if err != nil {
                bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: I can't clear my recent memory. It may already be empty.", user))
            } else {
                bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: My recent memory has been cleared.", user))
            }
            bot.IsAvailable = true
        case "!system":
            newSystemMessage := Message{
                Role:    "system",
                Content: query,
            }
            history := loadMessageHistory()
            history = append(history, newSystemMessage)
            saveMessageHistory(history)
            bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: My system instructions have been updated.", user))
            bot.IsAvailable = true
        default:
            if query != "" {
                bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: I will think about that and be back with you shortly.", user))
                go func() {
                    response := bot.callAPI(query, useContext)
                    bot.sendMessage(user, response)
                    bot.IsAvailable = true
                }()
            } else {
                bot.IsAvailable = true
            }
        }
    }
}

func (bot *Bot) sendMessage(user, response string) {
    messages := splitMessage(response, bot.Config.MaxMessageSize)
    for i, msg := range messages {
        if i == 0 {
            bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: %s", user, msg))
        } else {
            bot.IRCConnection.Privmsg(bot.Config.Channel, msg)
        }
        time.Sleep(1 * time.Second)
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
