package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	content, _ := ioutil.ReadFile("internal/commands/commands.go")
	str := string(content)

	str = strings.Replace(str, "Chat:      chatID,", "MessageSource: types.MessageSource{Chat: chatID, IsFromMe: true},", -1)
	str = strings.Replace(str, "IsFromMe:  true,", "", -1)
	
	str = strings.Replace(str, "Chat: chatID, IsFromMe: true,", "MessageSource: types.MessageSource{Chat: chatID, IsFromMe: true},", -1)
	
	str = strings.Replace(str, "return sendResp, err", "return &sendResp, err", 1) // wait, sendResp is whatsmeow.SendResponse, not pointer?

	ioutil.WriteFile("internal/commands/commands.go", []byte(str), 0644)
	fmt.Println("Done")
}
