package main

import (
	"crypto/tls"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"regexp"
	"strings"
	"time"
)

type IRCBot struct {
	*Bot
	IRCConnection *irc.Connection
}

func NewIRCBot(bot *Bot) *IRCBot {
	ircBot := &IRCBot{Bot: bot}
	nickname := "Nisaba"
	if bot.Config.Nickname != nil {
		nickname = *bot.Config.Nickname
	}

	irccon := irc.IRC(nickname, nickname)
	irccon.VerboseCallbackHandler = bot.Config.Debug != nil && *bot.Config.Debug
	irccon.Debug = bot.Config.Debug != nil && *bot.Config.Debug

	useSSL := bot.Config.UseSSL != nil && *bot.Config.UseSSL
	irccon.UseTLS = useSSL
	validateSSL := bot.Config.ValidateSSL != nil && *bot.Config.ValidateSSL
	if useSSL {
		irccon.TLSConfig = &tls.Config{InsecureSkipVerify: !validateSSL}
		if validateSSL {
			irccon.TLSConfig.ServerName = bot.Config.Server
		}
	}

	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(bot.Config.Channel) })
	irccon.AddCallback("PRIVMSG", ircBot.handleMessage)

	ircBot.IRCConnection = irccon
	return ircBot
}

func (ircBot *IRCBot) handleMessage(e *irc.Event) {
	if blockedUsers[e.Nick] {
		return
	}

	message := e.Message()
	re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(*ircBot.Config.Nickname) + `[:,]?\s?(.*)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(message))
	if len(matches) > 1 {
		user := e.Nick
		entireMessage := matches[1]
		if strings.HasPrefix(entireMessage, "!") {
			handleCommands(ircBot.Bot, strings.Fields(entireMessage)[0], strings.Join(strings.Fields(entireMessage)[1:], " "), user, ircBot.sendIRCMessage)
		} else {
			ircBot.processMessage(user, entireMessage)
		}
	}
}

func (ircBot *IRCBot) sendIRCMessage(channel, message string) {
	ircBot.IRCConnection.Privmsg(channel, message)
}

func (ircBot *IRCBot) processMessage(user, message string) {
	if len(message) == 0 {
		return
	}
	ircBot.sendIRCMessage(ircBot.Config.Channel, fmt.Sprintf("%s: I will think about that and be back with you shortly.", user))
	go func() {
		response := ircBot.callAPI(message)
		ircBot.sendMessage(user, response)
	}()
}

func (ircBot *IRCBot) sendMessage(user, response string) {
	messages := splitMessage(response, *ircBot.Config.MessageSize)
	delay := time.Duration(*ircBot.Config.Delay) * time.Second
	for i, msg := range messages {
		if i == 0 {
			ircBot.sendIRCMessage(ircBot.Config.Channel, fmt.Sprintf("%s: %s", user, msg))
		} else {
			time.Sleep(delay)
			ircBot.sendIRCMessage(ircBot.Config.Channel, msg)
		}
	}
}

func (ircBot *IRCBot) ConnectAndListen() {
	serverAndPort := fmt.Sprintf("%s:%s", ircBot.Config.Server, *ircBot.Config.Port)
	if err := ircBot.IRCConnection.Connect(serverAndPort); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	ircBot.IRCConnection.Loop()
}
