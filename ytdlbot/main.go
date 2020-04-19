package main

import (
	"context"
	icq "github.com/mail-ru-im/bot-golang"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)
func main() {
	r := regexp.MustCompile("http(?:s?):\\/\\/(?:www\\.)?youtu(?:be\\.com\\/watch\\?v=|\\.be\\/)([\\w\\-\\_]*)(&(amp;)?‌​[\\w\\?‌​=]*)?")
	title := regexp.MustCompile("<title>(.*?)</title>")
	ctx := context.Background()

	bot, err := icq.NewBot("TOKEN_HERE", icq.BotDebug(true))
	if err != nil {
		log.Println("wrong token")
	}

	updates := bot.GetUpdatesChannel(ctx)

STARTOVER:
	for update := range updates {
		if update.Type == icq.NEW_MESSAGE {
			switch update.Payload.Text {
			case "/start":
				err = update.Payload.Message().Reply(`Send me a link to a Youtube video and I will send you an audio version of it as a file!

Пришли мне ссылку на видео с Youtube и я отправлю тебе его аудио версию файлом!`)
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply(err.Error())
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					goto STARTOVER
				}
			case "/help":
				// do nothing
			case "/stop":
				// do nothing

			default:
				if !r.MatchString(update.Payload.Text) {
					err = update.Payload.Message().Reply("Invalid url!")
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					goto STARTOVER
				}

				err = update.Payload.Message().Reply("Started downloading...")
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
				}

				resp, err := http.Get(update.Payload.Text)
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply(err.Error())
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					goto STARTOVER
				}

				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					goto STARTOVER
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					goto STARTOVER
				}

				err = resp.Body.Close()
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
				}

				youtubeTitle := title.Find(body)
				path := string(youtubeTitle[7:len(youtubeTitle)-8]) + ".opus"

				_, err = exec.Command("youtube-dl", "--add-metadata",
					"--extract-audio", "--print-json", "-o", path, update.Payload.Text).Output()

				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply(err.Error())
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					goto STARTOVER
				}

				file, err := os.Open(path)

				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply(err.Error())
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					goto STARTOVER
				}

				err = bot.NewFileMessage(update.Payload.Chat.ID, file).Send()
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					goto STARTOVER
				}

				err = file.Close()
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					goto STARTOVER
				}


				err = os.Remove(path)
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply(err.Error())
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
				}
			}
		}
	}
}
