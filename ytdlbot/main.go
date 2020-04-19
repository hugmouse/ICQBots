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
	"time"
)

func main() {
	r := regexp.MustCompile("http(?:s?):\\/\\/(?:www\\.)?youtu(?:be\\.com\\/watch\\?v=|\\.be\\/)([\\w\\-\\_]*)(&(amp;)?‌​[\\w\\?‌​=]*)?")
	title := regexp.MustCompile("<title>(.*?)</title>")
	ctx := context.Background()

	bot, err := icq.NewBot("TOKEN_HERE")
	if err != nil {
		log.Println("wrong token")
	}

	updates := bot.GetUpdatesChannel(ctx)

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
					continue
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
					continue
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
					continue
				}

				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					continue
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					continue
				}

				err = resp.Body.Close()
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
				}

				youtubeTitle := title.Find(body)
				path := string(youtubeTitle[7:len(youtubeTitle)-8]) + ".opus"

				cmdd := exec.Command("youtube-dl", "--add-metadata",
					"--extract-audio", "--print-json", "-o", path, update.Payload.Text)

				cmdd.Start()

				done := make(chan error)
				go func() {
					done <- cmdd.Wait()
				}()

				// Timeout timer
				timeout := time.After(10 * time.Second)

				select {
				case <-timeout:
					// Timeout happened first,
					// kill the process and print a message
					cmdd.Process.Kill()
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply("Sorreh, timed out (timeout is 10 seconds)")
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					continue
				case err := <-done:
					// Command completed before timeout
					// Print output and error if it exists
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
						err = update.Payload.Message().Reply(err.Error())
						if err != nil {
							logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
						}
						continue
					}
				}

				file, err := os.Open(path)

				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					err = update.Payload.Message().Reply(err.Error())
					if err != nil {
						logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					}
					continue
				}

				err = bot.NewFileMessage(update.Payload.Chat.ID, file).Send()
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					continue
				}

				err = file.Close()
				if err != nil {
					logrus.WithFields(logrus.Fields{"EventID": update.EventID}).Errorln(err)
					continue
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
