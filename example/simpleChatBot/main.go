package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/wejick/gochain/chain/conversation"
	"github.com/wejick/gochain/model"
	_openai "github.com/wejick/gochain/model/openAI"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type .quit to exit")

	var authToken = os.Getenv("OPENAI_API_KEY")
	chatModel := _openai.NewOpenAIChatModel(authToken, "", _openai.GPT3Dot5Turbo0301)
	memory := []model.ChatMessage{}
	streamingChannel := make(chan model.ChatMessage, 100)
	convoChain := conversation.NewConversationChain(chatModel, memory, "You're helpful chatbot that answer human question very concisely")
	convoChain.AppendToMemory(model.ChatMessage{Role: model.ChatMessageRoleAssistant, Content: "Hi, My name is GioAI"})

	for {
		fmt.Print("User : ")
		chat, _ := reader.ReadString('\n')

		// Remove newline character from the command string
		chat = chat[:len(chat)-1]

		if chat == ".quit" {
			break
		}

		var output string
		go func() {
			var err error
			output, err = convoChain.SimpleRun(context.Background(), chat, model.WithIsStreaming(true), model.WithStreamingChannel(streamingChannel))
			if err != nil {
				fmt.Println("error " + err.Error())
				return
			}
		}()
		fmt.Print("AI : ")
		for {
			value, ok := <-streamingChannel
			if ok && !model.IsStreamFinished(value) {
				fmt.Print(value.Content)
			} else {
				fmt.Println("")
				break
			}
		}
		convoChain.AppendToMemory(model.ChatMessage{Role: model.ChatMessageRoleUser, Content: chat})
		convoChain.AppendToMemory(model.ChatMessage{Role: model.ChatMessageRoleAssistant, Content: output})
	}

	fmt.Println("Program exited.")
}
