package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"net/http"
	_ "os"
	"strings"
)

var bot = createBot()

func handler(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Path[1:]
	split := strings.Split(s, "/")
	if len(split) < 2 {
		fmt.Fprintf(w, "not enought parameters, need 2, telegram-user & message")
		return
	}
	bytes, e := ioutil.ReadAll(r.Body)
	if e == nil {
	}

	name := split[0]
	if name == "charm" {
		sendPhotoToCharm(bytes)
		sendMessageToCharm(split[1])
	}
	if name == "shumik" {
		sendMessageToShumik(split[1])
		sendPhotoToShumik(bytes)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createBot() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI("605258699:AAG3EeFM-ETkvJ0NivCGs7K8YpNWAwtYqUE")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot
}

func sendMessageToCharm(text string) {
	message := tgbotapi.NewMessage(150789681, text)
	bot.Send(message)
}

func sendPhotoToCharm(bytes []byte) {
	fileBytes := tgbotapi.FileBytes{"name", bytes}
	message := tgbotapi.NewPhotoUpload(150789681, fileBytes)
	bot.Send(message)
}

func sendPhotoToShumik(bytes []byte) {
	fileBytes := tgbotapi.FileBytes{"name", bytes}
	message := tgbotapi.NewPhotoUpload(146395526, fileBytes)
	bot.Send(message)
}

func sendMessageToShumik(text string) {
	message := tgbotapi.NewMessage(146395526, text)
	bot.Send(message)
}
