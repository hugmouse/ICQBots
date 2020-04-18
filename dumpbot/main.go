package main

import (
    "context"
    "encoding/json"
    "github.com/mail-ru-im/bot-golang"
    "github.com/tidwall/pretty"
    "log"
)

func main() {
    bot, err := botgolang.NewBot("TOKEN HERE")
    if err != nil {
        log.Println("wrong token")
    }
    ctx, _ := context.WithCancel(context.Background())
    updates := bot.GetUpdatesChannel(ctx)
    for update := range updates {
        if update.Type == botgolang.NEW_MESSAGE {
            switch update.Payload.Message().Text {
            case "/help":
                // do nothing
            case "/start":
                err = update.Payload.Message().Reply("Send any message and I'll answer you with a message dump!")
                if err != nil {
                    log.Println(err)
                }
            case "/stop":
                // do nothing
            }
            jsn, _ := json.Marshal(update)
            formatted := "```" + string(pretty.Pretty(jsn)) + "```"
            err = update.Payload.Message().Reply(formatted)
            if err != nil {
                log.Println(err)
            }

            if len(update.Payload.Parts) > 0 && update.Payload.Parts[0].Type == botgolang.FILE {
                msg := bot.NewFileMessageByFileID(update.Payload.Chat.ID, update.Payload.Parts[0].Payload.FileID)
                err = msg.Send()
                if err != nil {
                    log.Println(err)
                }
            }
        }
    }
}