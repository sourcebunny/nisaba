package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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
	Delay       *int    `json:"delay"`
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
	Config      Config
	Options     *Options
	IsAvailable bool
}

func NewBot(config Config) *Bot {
	return &Bot{
		Config:      config,
		IsAvailable: true,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var sendMessage func(channel, message string)

func getConfigFilePath(fileName string) string {
	if profileDir != "" {
		profilePath := filepath.Join("profiles", profileDir, fileName)
		if _, err := os.Stat(profilePath); err == nil {
			return profilePath
		}
	}

	configDir := "config"
	defaultPath := filepath.Join(configDir, fileName)
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}

	if _, err := os.Stat(fileName); err == nil {
		return fileName
	}

	return filepath.Join(configDir, fileName)
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
	if config.Delay == nil || *config.Delay < 1 {
		defaultDelay := 3
		config.Delay = &defaultDelay
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

var profileDir string

func loadProfile(bot *Bot, profileName string, user string) {
	if profileName == "" {
		profileDir = ""
		opts, err := loadOptions("options.json")
		if err == nil {
			bot.Options = opts
			log.Printf("Options file for API endpoint has been reloaded with default settings.")
		}
		loadMessageHistory()
		loadBlockedUsers()
		sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Profile directory has been reset to default settings.", user))
	} else if match, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, profileName); match {
		dirPath := filepath.Join("profiles", profileName)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: The directory does not exist: '%s'.", user, dirPath))
		} else {
			profileDir = profileName
			opts, err := loadOptions("options.json")
			if err == nil {
				bot.Options = opts
				log.Printf("Options file for API endpoint has been reloaded for profile '%s'.", profileName)
			}
			loadMessageHistory()
			loadBlockedUsers()
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Configuration directory is set to '%s'.", user, dirPath))
		}
	} else {
		sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Invalid directory name. Only alphanumeric characters are allowed.", user))
	}
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

func loadReminderPrompt() string {
	filePath := getConfigFilePath("reminderprompt.txt")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		log.Fatalf("Error reading reminder prompt file: %v", err)
	}
	return string(content)
}

func getHistoryFilePath() string {
	var filePath = "history.txt"

	if profileDir != "" {
		customPath := filepath.Join("profiles", profileDir, "history.txt")
		if _, err := os.Stat(filepath.Join("profiles", profileDir)); err == nil {
			filePath = customPath
		}
	} else {
		configPath := filepath.Join("config", "history.txt")
		if _, err := os.Stat(filepath.Join("config")); err == nil {
			filePath = configPath
		}
	}

	return filePath
}

func createMessageHistory() {
	filePath := getHistoryFilePath()
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

func loadMessageHistory() []Message {
	filePath := getHistoryFilePath()
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
	filePath := getHistoryFilePath()
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

func saveHistoryArchive(index int) (int, error) {
	basePath := getHistoryFilePath()

	baseDir, baseName := filepath.Split(basePath)
	extension := filepath.Ext(baseName)
	baseName = baseName[:len(baseName)-len(extension)]

	if index == 0 {
		// Find the highest existing file and increment by one
		for i := 9999; i >= 1; i-- {
			historyArchivePath := fmt.Sprintf("%s%s.%d%s", baseDir, baseName, i, extension)
			if _, err := os.Stat(historyArchivePath); !os.IsNotExist(err) {
				index = i + 1
				break
			}
		}
		if index == 0 {
			index = 1
		}
		if index > 9999 {
			return 0, fmt.Errorf("maximum history file index reached")
		}
	} else if index < 1 || index > 9999 {
		return 0, fmt.Errorf("index out of range")
	}

	historyArchivePath := fmt.Sprintf("%s%s.%d%s", baseDir, baseName, index, extension)

	content, err := ioutil.ReadFile(basePath)
	if err != nil {
		return index, err
	}

	err = ioutil.WriteFile(historyArchivePath, content, 0644)
	return index, err
}

func loadHistoryArchive(index int) (int, error) {
	basePath := getHistoryFilePath()

	baseDir, baseName := filepath.Split(basePath)
	extension := filepath.Ext(baseName)
	baseName = baseName[:len(baseName)-len(extension)]

	if index == 0 {
		// Find the highest existing file
		for i := 9999; i >= 1; i-- {
			historyArchivePath := fmt.Sprintf("%s%s.%d%s", baseDir, baseName, i, extension)
			if _, err := os.Stat(historyArchivePath); !os.IsNotExist(err) {
				index = i
				break
			}
		}
		if index == 0 {
			return 0, fmt.Errorf("no history file found to load")
		}
	} else if index < 1 || index > 9999 {
		return 0, fmt.Errorf("index out of range")
	}

	historyArchivePath := fmt.Sprintf("%s%s.%d%s", baseDir, baseName, index, extension)

	content, err := ioutil.ReadFile(historyArchivePath)
	if err != nil {
		return index, err
	}

	err = ioutil.WriteFile(basePath, content, 0644)
	return index, err
}

func (bot *Bot) callAPI(query string) string {
	bot.IsAvailable = false
	defer func() { bot.IsAvailable = true }()

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

			// Append the reminder prompt if it exists
			reminderPrompt := loadReminderPrompt()
			if reminderPrompt != "" {
				reminderMessage := Message{Role: "system", Content: reminderPrompt}
				saveMessageHistory([]Message{reminderMessage})
			}
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

func handleCommands(bot *Bot, command, query, user string) {
	switch command {
	case "!clear":
		historyFilePath := getHistoryFilePath()
		if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: I can't clear my recent memory. It may already be empty.", user))
		} else {
			createMessageHistory()
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: My recent memory has been cleared.", user))
		}
	case "!system":
		newSystemMessage := Message{Role: "system", Content: query}
		saveMessageHistory([]Message{newSystemMessage})
		sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Specified system prompt will be attached to the next message.", user))
	case "!options":
		optionsFile := fmt.Sprintf("options.%s.json", query)
		newOptions, err := loadOptions(optionsFile)
		if err != nil {
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Failed to load options from '%s'.", user, optionsFile))
		} else {
			bot.Options = newOptions
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Options loaded successfully from '%s'.", user, optionsFile))
		}
	case "!profile":
		loadProfile(bot, query, user)
	case "!save":
		index := 0
		autoSelect := false
		if query == "" || query == "0" {
			autoSelect = true
		} else {
			var err error
			index, err = strconv.Atoi(query)
			if err != nil || index < 1 || index > 9999 {
				sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Invalid index provided. Use a number between 1 and 9999, or zero for auto-selection.", user))
				return
			}
		}
		idxUsed, err := saveHistoryArchive(index)
		if err != nil {
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Error saving history: %s", user, err))
		} else {
			if autoSelect {
				index = idxUsed
			}
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: History successfully saved as history.%d.txt", user, index))
		}
	case "!load":
		index := 0
		autoSelect := false
		if query == "" || query == "0" {
			autoSelect = true
		} else {
			var err error
			index, err = strconv.Atoi(query)
			if err != nil || index < 1 || index > 9999 {
				sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Invalid index provided. Use a number between 1 and 9999, or zero for auto-selection.", user))
				return
			}
		}
		idxUsed, err := loadHistoryArchive(index)
		if err != nil {
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: Error loading history: %s", user, err))
		} else {
			if autoSelect {
				index = idxUsed
			}
			sendMessage(bot.Config.Channel, fmt.Sprintf("%s: History successfully loaded from history.%d.txt", user, index))
		}
	}
}

func splitMessage(response string, maxSize int) []string {
	var parts []string
	var currentSize int
	var currentPart bytes.Buffer

	// Regular expression to collapse multiple newlines
	re := regexp.MustCompile(`\n+`)
	normalizedResponse := re.ReplaceAllString(response, "\n")

	for _, runeValue := range normalizedResponse {
		if currentSize+len(string(runeValue)) > maxSize || runeValue == '\n' {
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
				currentSize = 0
			}
			// Skip directly appending newline to avoid empty strings
			if runeValue != '\n' {
				currentPart.WriteRune(runeValue)
				currentSize += len(string(runeValue))
			}
		} else {
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

	ircBot := NewIRCBot(bot)
	sendMessage = ircBot.sendIRCMessage
	ircBot.ConnectAndListen()
}
