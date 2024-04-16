package main

import (
	"bufio"
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
	"reflect"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Channel     string  `json:"channel"`
	Server      string  `json:"server"`
	Nickname    *string `json:"nickname"`
	Port        *string `json:"port"`
	UseSSL      *bool   `json:"use_ssl"`
	ValidateSSL *bool   `json:"validate_ssl"`
	Commands    *bool   `json:"commands"`
	Debug       *bool   `json:"debug"`
	APIURL      *string `json:"api_url"`
	APIKey      *string `json:"api_key"`
	APIMode     *string `json:"api_mode"`
	MessageSize *int    `json:"message_size"`
}

type Options struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	MinP             *float64 `json:"min_p,omitempty"`
	NPredict         *int     `json:"n_predict,omitempty"`
	NKeep            *int     `json:"n_keep,omitempty"`
	TfsZ             *float64 `json:"tfs_z,omitempty"`
	TypicalP         *float64 `json:"typical_p,omitempty"`
	RepeatPenalty    *float64 `json:"repeat_penalty,omitempty"`
	RepeatLastN      *int     `json:"repeat_last_n,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	Mirostat         *int     `json:"mirostat,omitempty"`
	MirostatTau      *float64 `json:"mirostat_tau,omitempty"`
	MirostatEta      *float64 `json:"mirostat_eta,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	NProbs           *int     `json:"n_probs,omitempty"`
	SlotID           *int     `json:"slot_id,omitempty"`
	PenalizeNL       *bool    `json:"penalize_nl,omitempty"`
	IgnoreEOS        *bool    `json:"ignore_eos,omitempty"`
	CachePrompt      *bool    `json:"cache_prompt,omitempty"`
	PenaltyPrompt    *string  `json:"penalty_prompt,omitempty"`
	SystemPrompt     *string  `json:"system_prompt,omitempty"`
}

type Bot struct {
	Config        Config
	Options       *Options
	IRCConnection *irc.Connection
	IsAvailable   bool
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func getConfigFilePath(fileName string) string {
	configDir := "config"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fileName
	} else {
		return filepath.Join(configDir, fileName)
	}
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

	// Validate mandatory fields
	if config.Server == "" {
		log.Fatalf("Mandatory configuration missing: 'server' is not set in config.json")
	}
	if config.Channel == "" {
		log.Fatalf("Mandatory configuration missing: 'channel' is not set in config.json")
	}

	// Set defaults for optional fields if not present
	if config.Nickname == nil {
		defaultNickname := "Nisaba"
		config.Nickname = &defaultNickname
	}
	if config.Port == nil {
		defaultPort := "6667"
		config.Port = &defaultPort
	}
	if config.APIURL == nil {
		defaultAPIURL := "http://localhost:8080/v1/chat/completions"
		config.APIURL = &defaultAPIURL
	}
	if config.APIKey == nil {
		defaultAPIKey := "null"
		config.APIKey = &defaultAPIKey
	}
	if config.APIMode == nil {
		defaultAPIMode := "chat"
		config.APIMode = &defaultAPIMode
	}
	if config.UseSSL == nil {
		defaultUseSSL := false
		config.UseSSL = &defaultUseSSL
	}
	if config.ValidateSSL == nil {
		defaultValidateSSL := false
		config.ValidateSSL = &defaultValidateSSL
	}
	if config.Commands == nil {
		defaultCommands := true
		config.Commands = &defaultCommands
	}
	if config.MessageSize == nil {
		defaultMessageSize := 400
		config.MessageSize = &defaultMessageSize
	}

	return config
}

func loadOptions(fileName string) (*Options, error) {
	filePath := getConfigFilePath(fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var opts Options
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&opts); err != nil {
		return nil, err
	}
	return &opts, nil
}

var blockedUsers map[string]bool

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

func createMessageHistory() {
	filePath := getConfigFilePath("history.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		var history []Message
		systemPromptContent := loadSystemPrompt()
		if systemPromptContent != "" {
			initialSystemMessage := Message{Role: "system", Content: systemPromptContent}
			history = append(history, initialSystemMessage)
		}
		fileContent, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			log.Fatalf("Error encoding message history: %v", err)
		}
		if err := ioutil.WriteFile(filePath, fileContent, 0644); err != nil {
			log.Fatalf("Error writing initial message history: %v", err)
		}
	}
}

func loadMessageHistory() []Message {
	filePath := getConfigFilePath("history.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		createMessageHistory()
	}

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading message history: %v", err)
	}
	var history []Message
	if err := json.Unmarshal(fileContent, &history); err != nil {
		log.Fatalf("Error parsing message history: %v", err)
	}
	return history
}

func saveMessageHistory(newMessages []Message) {
	filePath := getConfigFilePath("history.txt")
	existingHistory := loadMessageHistory()
	updatedHistory := append(existingHistory, newMessages...)

	fileContent, err := json.MarshalIndent(updatedHistory, "", "  ")
	if err != nil {
		log.Fatalf("Error encoding message history: %v", err)
	}

	if err := ioutil.WriteFile(filePath, fileContent, 0644); err != nil {
		log.Fatalf("Error writing message history: %v", err)
	}
}

func NewBot(config Config) *Bot {
	bot := &Bot{
		Config:      config,
		IsAvailable: true,
	}

	nickname := "Nisaba"
	if config.Nickname != nil {
		nickname = *config.Nickname
	}

	irccon := irc.IRC(nickname, nickname)

	debug := false
	if config.Debug != nil {
		debug = *config.Debug
	}
	irccon.VerboseCallbackHandler = debug
	irccon.Debug = debug

	useSSL := false
	if config.UseSSL != nil {
		useSSL = *config.UseSSL
	}
	irccon.UseTLS = useSSL

	validateSSL := false
	if config.ValidateSSL != nil {
		validateSSL = *config.ValidateSSL
	}
	if useSSL {
		irccon.TLSConfig = &tls.Config{
			InsecureSkipVerify: !validateSSL,
		}
		if validateSSL {
			irccon.TLSConfig.ServerName = config.Server
		}
	}

	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(config.Channel) })

	irccon.AddCallback("PRIVMSG", bot.handleMessage)

	bot.IRCConnection = irccon
	return bot
}

func (bot *Bot) callAPI(query string) string {
	var responseContent string
	var payload map[string]interface{}

	// Prepare payload for the API call based on the mode

	// Use "chat" for "/v1/chat/completions" endpoint
	// Use "query" for "/completion" endpoint

	if *bot.Config.APIMode == "chat" {
		history := loadMessageHistory()
		newUserMessage := Message{Role: "user", Content: query}
		history = append(history, newUserMessage)
		saveMessageHistory([]Message{newUserMessage})

		messagesPayload := make([]map[string]interface{}, len(history))
		for i, msg := range history {
			messagesPayload[i] = map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}
		}

		payload = map[string]interface{}{
			"messages": messagesPayload,
			"stream":   false,
		}
	} else if *bot.Config.APIMode == "query" {
		payload = map[string]interface{}{
			"prompt": query,
			"stream": false,
		}
	}

	// Include options from the Options struct
	if bot.Options != nil {
		val := reflect.ValueOf(*bot.Options)
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.IsNil() {
				payloadKey := strings.ToLower(typ.Field(i).Name)
				payload[payloadKey] = field.Elem().Interface()
			}
		}
	}

	// Serialize the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding payload to JSON: %v", err)
		return "Error encoding request payload."
	}

	// Sending the payload to the API
	log.Printf("Sending payload: %s\n", string(payloadBytes))
	req, err := http.NewRequest("POST", *bot.Config.APIURL, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*bot.Config.APIKey)

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

	// Reading the response from the API
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "Error reading response."
	}

	log.Printf("Received response: %s\n", string(body))

	// Parsing the response
	if *bot.Config.APIMode == "chat" {
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
			// Append the assistant's response to the message history
			responseMessage := Message{Role: "assistant", Content: responseContent}
			saveMessageHistory([]Message{responseMessage})
		}
	} else if *bot.Config.APIMode == "query" {
		// Directly parse the response content for query mode
		var response struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			log.Printf("Error decoding response from API: %v", err)
			return "Error parsing response."
		}
		responseContent = response.Content
	}

	return responseContent
}

func (bot *Bot) handleMessage(e *irc.Event) {
	if !bot.IsAvailable || blockedUsers[e.Nick] {
		return
	}

	message := e.Message()

	// Check if the message starts with the bot's nickname
	re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(*bot.Config.Nickname) + `[:,]?\s?(.*)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(message))

	if len(matches) > 1 {
		bot.IsAvailable = false
		user := e.Nick
		entireMessage := matches[1]

		// Split to check for commands
		parts := strings.Fields(entireMessage)
		if len(parts) == 0 {
			bot.IsAvailable = true
			return
		}

		firstWord, restOfMessage := parts[0], strings.Join(parts[1:], " ")

		if !*bot.Config.Commands && strings.HasPrefix(firstWord, "!") {
			bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: Commands are currently disabled.", user))
			bot.IsAvailable = true
			return
		}

		switch firstWord {
		case "!clear", "!system", "!options":
			handleCommands(bot, firstWord, restOfMessage, user)
		default:
			// Handle as a normal message if no command is detected
			query := entireMessage
			bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: I will think about that and be back with you shortly.", user))
			go func() {
				response := bot.callAPI(query)
				bot.sendMessage(user, response)
				bot.IsAvailable = true
			}()
		}
	}
}

func handleCommands(bot *Bot, command, query, user string) {
	switch command {
	case "!clear":
		err := os.Remove(getConfigFilePath("history.txt"))
		if err != nil {
			bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: I can't clear my recent memory. It may already be empty.", user))
		} else {
			bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: My recent memory has been cleared.", user))
		}
	case "!system":
		newSystemMessage := Message{Role: "system", Content: query}
		saveMessageHistory([]Message{newSystemMessage})
	case "!options":
		optionsFile := fmt.Sprintf("options.%s.json", query)
		newOptions, err := loadOptions(optionsFile)
		if err != nil {
			bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: Failed to load options from '%s'.", user, optionsFile))
		} else {
			bot.Options = newOptions
			bot.IRCConnection.Privmsg(bot.Config.Channel, fmt.Sprintf("%s: Options loaded successfully from '%s'.", user, optionsFile))
		}
	}
	bot.IsAvailable = true
}

func (bot *Bot) sendMessage(user, response string) {
	messages := splitMessage(response, *bot.Config.MessageSize)
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

	defaultOptions, err := loadOptions("options.json")
	if err != nil {
		log.Printf("No default options loaded: %v", err)
	} else {
		log.Println("Default options loaded successfully.")
	}

	bot := NewBot(config)
	bot.Options = defaultOptions

	serverAndPort := fmt.Sprintf("%s:%s", config.Server, *config.Port)
	if err := bot.IRCConnection.Connect(serverAndPort); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	bot.IRCConnection.Loop()
}
