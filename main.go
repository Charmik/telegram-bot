package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	_ "os"
	"strconv"
	"strings"
	"time"
)

var bot = createBot()
var updateId = 0
var CHAT_IDS_FILE = "chatIds.txt"

func main() {
	//scheduleUpdates()
	getUpdates()

	url := "https://yandex.ru/pogoda/moscow"
	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		panic(err)
	}
	// do this now so it won't be forgotten
	defer resp.Body.Close()
	// reads html as a slice of bytes
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	// show the HTML code as a string %s
	//fmt.Printf("%s\n", html)
	s := string(html[:])

	index := strings.Index(s, "мм рт. ст.")
	pressure_mm := html[index-29 : index-26]
	//fmt.Printf("%s\n", pressure_mm)

	indexTemp := strings.Index(s, "<span class=\"temp__value\">")
	strTemp := html[indexTemp+26:]
	temperature := strTemp[:strings.Index(string(strTemp[:]), "<")]
	//fmt.Printf("%s\n", temperature)

	indexWind := strings.Index(s, "<span class=\"wind-speed\">")
	windStr := html[indexWind+25:]
	windSpeed := windStr[:strings.Index(string(windStr[:]), "<")]
	//fmt.Printf("%s\n", windSpeed)
	windDirection := windStr[strings.Index(string(windStr[:]), "title=\"")+7:]
	windDirection = windDirection[:strings.Index(string(windDirection[:]), "\"")]
	windDirection = windDirection[12:]
	//fmt.Printf("%s\n", windDirection)

	humidityIndex := strings.Index(s, "<i class=\"icon icon_humidity-white term__fact-icon\"></i>")
	humidity := html[humidityIndex+56:]
	humidity = humidity[:strings.Index(string(humidity[:]), "<")]
	//fmt.Printf("%s\n", humidity)

	conditionIndex := strings.Index(s, "link__condition")
	condition := html[conditionIndex+56:]
	condition = condition[:strings.Index(string(condition[:]), "<")]
	condition = condition[strings.Index(string(condition[:]), ">")+1:]
	//fmt.Printf("%s\n", condition)

	saveToFile(string(pressure_mm[:]), string(temperature[:]), string(humidity[:]), string(windSpeed[:]), string(windDirection[:]), string(condition[:]))

	sendAll()
}

func saveToFile(pressure_mm string, temperature string, humidity string, wind_speed string, wind_dir string, condition string) {
	f, err := os.Create("currentTemperature.txt")
	if err != nil {
		fmt.Printf("error creating file: %v", err)
		return
	}
	defer f.Close()

	loc, _ := time.LoadLocation("Europe/Moscow")
	time := time.Now().In(loc).Add(time.Hour*time.Duration(0) +
		time.Minute*time.Duration(10) +
		time.Second*time.Duration(0))

	str := fmt.Sprintf("%02d", time.Hour()) + ":" + fmt.Sprintf("%02d", time.Minute()) + "\n\n" +
		fmt.Sprintf("%-17s", "Давление") + pressure_mm + " мм.рт.ст.\n" +
		fmt.Sprintf("%-17s", "Температура") + temperature + "\n" +
		fmt.Sprintf("%-17s", "Отн. влажность") + humidity + "\n" +
		fmt.Sprintf("%-17s", "Ветер") + wind_speed + " м/c, " + wind_dir + "\n\n" +
		condition
		//mapConditionToRussian(condition)

	fmt.Println(str)
	_, _ = f.WriteString(str)
}

func sendAll() {
	/*
		resp_body, err := getJsonFromYandexWeather()

		if err != nil {
			fmt.Printf("couldn't call get from yandex-api", err)
			return
		}
		var jsonResponse interface{}
		parseError := json.Unmarshal(resp_body, &jsonResponse)
		fmt.Print("parseError: ")
		fmt.Println(parseError)
		fmt.Print("response: ")
		fmt.Println(jsonResponse)

		jsonMap := jsonResponse.(map[string]interface{})
		handle(jsonMap)
	*/

	bytes, _ := ioutil.ReadFile(CHAT_IDS_FILE)
	str := string(bytes)
	chatIds := strings.Split(str, "\n")
	for _, chatIdStr := range chatIds {
		if len(chatIdStr) == 0 {
			continue
		}
		fmt.Println("send to chatId", chatIdStr)
		chatId, _ := strconv.ParseInt(chatIdStr, 10, 64)
		upload := tgbotapi.NewDocumentUpload(chatId, "currentTemperature.txt")
		bot.Send(upload)

		loc, _ := time.LoadLocation("Europe/Moscow")
		time := time.Now().In(loc)
		if time.Hour() == 11 {
			message := tgbotapi.NewMessage(chatId, "http://old.meteoinfo.ru/forecasts")
			bot.Send(message)
		}
	}
}

func scheduleUpdates() {
	s := gocron.NewScheduler()
	s.Every(5).Seconds().Do(getUpdates)
	<-s.Start()
}

func getUpdates() {
	fmt.Println("getUpdates updateId=", updateId)
	config := tgbotapi.UpdateConfig{Offset: updateId + 1, Limit: 100500, Timeout: 100}
	updates, e := bot.GetUpdates(config)
	if e != nil {
		fmt.Printf("couldn't get updates from bot", e)
	}
	for _, update := range updates {
		if update.Message.Text == "/start" {
			chatId := update.Message.Chat.ID
			addNewChatId(chatId)
		}
		if update.UpdateID > updateId {
			updateId = update.UpdateID
		}
		fmt.Println("got update test:", update.Message.Text)
	}
}

func addNewChatId(chatId int64) {
	if _, err := os.Stat(CHAT_IDS_FILE); err == nil {
		f, _ := os.OpenFile(CHAT_IDS_FILE, os.O_APPEND|os.O_WRONLY, 0600)
		writeChatId(f, chatId)
	} else if os.IsNotExist(err) {
		os.Create(CHAT_IDS_FILE)
		f, _ := os.OpenFile(CHAT_IDS_FILE, os.O_APPEND|os.O_WRONLY, 0600)
		writeChatId(f, chatId)
	}
}

func writeChatId(file *os.File, chatId int64) {
	bytes, _ := ioutil.ReadFile(CHAT_IDS_FILE)
	str := string(bytes)
	chatIdStr := strconv.FormatInt(chatId, 10)
	if !strings.Contains(str, chatIdStr) {
		_, _ = file.WriteString(chatIdStr + "\n")
	}
}

func getJsonFromYandexWeather() ([]byte, error) {
	client := &http.Client{}
	req, errGet := http.NewRequest(
		"GET", "https://api.weather.yandex.ru/v1/informers?lat=55.75396&lon=37.620393", nil)
	if errGet != nil {
		fmt.Printf("couldn't make get from yandex-api", errGet)
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("status: " + resp.Status + " ")
	return resp_body, err
}
func handle(m map[string]interface{}) {
	fact := (m["fact"]).(map[string]interface{})

	pressure_mm := fact["pressure_mm"].(float64)
	temperature := fact["temp"].(float64)
	humidity := fact["humidity"].(float64)
	wind_speed := fact["wind_speed"].(float64)
	wind_dir := fact["wind_dir"].(string)
	condition := fact["condition"].(string)
	fmt.Println(pressure_mm, "pressure_mm")
	fmt.Println(temperature, "temperature")
	fmt.Println(humidity, "humidity")
	fmt.Println(wind_speed, "wind_speed")
	fmt.Println(wind_dir, "wind_dir")
	fmt.Println(condition, "condition")

	f, err := os.Create("currentTemperature.txt")
	if err != nil {
		fmt.Printf("error creating file: %v", err)
		return
	}
	defer f.Close()

	time := time.Now().UTC()
	str := strconv.Itoa(time.Hour()) + ":" + strconv.Itoa(time.Minute()) + "\n" +
		fmt.Sprintf("%-17s", "Давление") + fmt.Sprintf("%.0f", pressure_mm) + "\n" +
		fmt.Sprintf("%-17s", "Температура") + fmt.Sprintf("%.0f", temperature) + "\n" +
		fmt.Sprintf("%-17s", "Отн. влажность") + fmt.Sprintf("%.0f", humidity) + "\n" +
		fmt.Sprintf("%-17s", "Ветер") + fmt.Sprintf("%.0f", wind_speed) + " м/c направление " + mapWindDirToRussian(wind_dir) + "\n" +
		"\n" +
		mapConditionToRussian(condition)

	fmt.Println(str)
	_, _ = f.WriteString(str)
}

func mapWindDirToRussian(windDir string) string {
	if windDir == "nw" {
		return "северо-западное"
	}
	if windDir == "n" {
		return "северое"
	}
	if windDir == "ne" {
		return "северо-восточное"
	}
	if windDir == "e" {
		return "восточное"
	}
	if windDir == "se" {
		return "юго-восточное"
	}
	if windDir == "s" {
		return "южное"
	}
	if windDir == "sw" {
		return "юго-западноее"
	}
	if windDir == "w" {
		return "западное"
	}
	if windDir == "c" {
		return "штиль"
	}
	return "неизвестное направление"
}

func mapConditionToRussian(condition string) string {
	if condition == "clear" {
		return "ясно"
	}
	if condition == "partly-cloudy" {
		return "малооблачно"
	}
	if condition == "cloudy" {
		return "облачно с прояснениями"
	}
	if condition == "overcast" {
		return "пасмурно"
	}
	if condition == "partly-cloudy-and-light-rain" {
		return "небольшой дождь"
	}
	if condition == "partly-cloudy-and-rain" {
		return "дождь"
	}
	if condition == "overcast-and-rain" {
		return "сильный дождь"
	}
	if condition == "overcast-thunderstorms-with-rain" {
		return "сильный дождь, гроза"
	}
	if condition == "cloudy-and-light-rain" {
		return "небольшой дождь"
	}
	if condition == "overcast-and-light-rain" {
		return "небольшой дождь"
	}
	if condition == "cloudy-and-rain" {
		return "дождь"
	}
	if condition == "overcast-and-wet-snow" {
		return "дождь со снегом"
	}
	if condition == "partly-cloudy-and-light-snow" {
		return "небольшой снег"
	}
	if condition == "partly-cloudy-and-snow" {
		return "снег"
	}
	if condition == "overcast-and-snow" {
		return "снегопад"
	}
	if condition == "cloudy-and-light-snow" {
		return "небольшой снег"
	}
	if condition == "overcast-and-light-snow" {
		return "небольшой снег"
	}
	if condition == "cloudy-and-snow" {
		return "снег"
	}
	return ""
}

func createBot() *tgbotapi.BotAPI {
	bytes, _ := ioutil.ReadFile("telegramKey.txt")
	botToken := string(bytes)

	bot, _ := tgbotapi.NewBotAPI(botToken)
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	return bot
}
