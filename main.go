package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"net/http"
	"strings"
)

var bot = createBot()

func handler(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Path[1:]
	split := strings.Split(s, "/")
	if (len(split) < 2) {
		fmt.Fprintf(w, "not enought parameters, need 2, botname & message")
		return
	}

	name := split[0]
	fmt.Fprintf(w, "Hi there, I love %s | %s", split[0], split[1])
	if name == "charm" {
		sendMessageToCharm(split[1])
	}
	if name == "shumik" {
		sendMessageToShumik(split[1])
	}
}

func main() {

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func createBot() (*tgbotapi.BotAPI) {
	bot, err := tgbotapi.NewBotAPI("605258699:AAG3EeFM-ETkvJ0NivCGs7K8YpNWAwtYqUE")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	/*
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
	*/
	return bot
}

func sendMessageToCharm(text string) {
	message := tgbotapi.NewMessage(150789681, text)
	bot.Send(message)
}

func sendMessageToShumik(text string) {
	message := tgbotapi.NewMessage(146395526, text)
	bot.Send(message)
}
