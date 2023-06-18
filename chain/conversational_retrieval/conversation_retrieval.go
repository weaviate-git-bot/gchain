package conversational_retrieval

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/wejick/gochain/callback"
	basechain "github.com/wejick/gochain/chain"
	"github.com/wejick/gochain/datastore"
	"github.com/wejick/gochain/model"
	"github.com/wejick/gochain/prompt"
	"github.com/wejick/gochain/textsplitter"
)

type answerOrLookupOutput struct {
	Question            string `json:"question"`
	Query               string `json:"query"`
	Intent              string `json:"intent"`
	Answer              string `json:"answer"`
	ConversationContext string `json:"conversation_context"`
	Lookup              bool   `json:"lookup"`
}

// ConversationalRetrievalChain conversation with ability to lookup data
type ConversationalRetrievalChain struct {
	chatModel           model.ChatModel // only allow using ChatModel
	memory              []model.ChatMessage
	retriever           datastore.Retriever
	textSplitter        textsplitter.TextSplitter
	callbackManager     *callback.Manager
	instructionTemplate *prompt.PromptTemplate
	answerTemplate      *prompt.PromptTemplate
	maxToken            int
}

func NewConversationalRetrievalChain(
	chatModel model.ChatModel,
	memory []model.ChatMessage,
	retriever datastore.Retriever,
	textSplitter textsplitter.TextSplitter,
	callbackManager *callback.Manager,
	firstSystemPrompt string,
	maxToken int,
	verbose bool,
) (chain *ConversationalRetrievalChain) {
	instructionTemplate, _ := prompt.NewPromptTemplate("instruction", instruction)
	answerTemplate, _ := prompt.NewPromptTemplate("answer", answeringInstruction)
	memory = append(memory, model.ChatMessage{Role: model.ChatMessageRoleSystem, Content: firstSystemPrompt})

	if verbose {
		callbackManager.RegisterCallback(basechain.CallbackChainEnd, callback.VerboseCallback)
	}
	if maxToken == 0 {
		maxToken = 1000
	}
	return &ConversationalRetrievalChain{
		chatModel:           chatModel,
		memory:              memory,
		retriever:           retriever,
		textSplitter:        textSplitter,
		callbackManager:     callbackManager,
		instructionTemplate: instructionTemplate,
		answerTemplate:      answerTemplate,
		maxToken:            maxToken,
	}
}

// Run expect chat["input"] as input, and put the result to output["output"]
func (C *ConversationalRetrievalChain) Run(ctx context.Context, chat map[string]string, options ...func(*model.Option)) (output map[string]string, err error) {
	if _, ok := chat["input"]; !ok {
		return output, errors.New("input[\"input\"] is not specified")
	}
	output = make(map[string]string)
	var answerOrLookup answerOrLookupOutput

	// trigger CallbackChainStart
	C.callbackManager.TriggerEvent(ctx, basechain.CallbackChainStart, callback.CallbackData{
		EventName:    basechain.CallbackChainStart,
		FunctionName: "ConversationalRetrievalChain.Run",
		Input:        chat,
		Output:       output})

	inputChat := model.ChatMessage{Role: model.ChatMessageRoleUser, Content: chat["input"]}

	answerOrLookup, err = C.answerOrLookup(ctx, chat["input"], options...)
	if err != nil {
		return
	}

	// trigger CallbackChainEnd, using lambda to defer the execution
	defer func(data callback.CallbackData) {
		C.callbackManager.TriggerEvent(ctx, basechain.CallbackChainEnd, data)
	}(callback.CallbackData{
		EventName:    basechain.CallbackChainEnd,
		FunctionName: "ConversationalRetrievalChain.Run",
		Input:        chat,
		Output:       output,
		Data:         answerOrLookup})

	// when we get direct answer
	if !answerOrLookup.Lookup {
		C.AppendToMemory(inputChat)
		C.AppendToMemory(model.ChatMessage{Role: model.ChatMessageRoleAssistant, Content: answerOrLookup.Answer})
		output["output"] = answerOrLookup.Answer

		return
	}

	// when we need to look up
	retrieverOutput, err := C.retriever.Search(ctx, "Indonesia", answerOrLookup.Query)
	if err != nil {
		return
	}
	var retrieverResult string
	for _, resp := range retrieverOutput {
		if data, ok := resp.(map[string]interface{}); ok {
			retrieverResult += data["text"].(string)
		}
	}

	answer, err := C.answerFromDoc(ctx, answerOrLookup, retrieverResult, options...)

	// append answer to memory
	C.AppendToMemory(inputChat)
	C.AppendToMemory(model.ChatMessage{Role: model.ChatMessageRoleAssistant, Content: answer})
	if err != nil {
		return
	}

	output["output"] = answer

	return
}

// answerOrLookup will return answer if it can, or return lookup query
// This approach is faster because we will be able to get answer directly when possible
func (C *ConversationalRetrievalChain) answerOrLookup(ctx context.Context, input string, options ...func(*model.Option)) (output answerOrLookupOutput, err error) {
	convoHistory := model.FlattenChatMessages(C.memory)
	instructionPrompt, err := C.instructionTemplate.FormatPrompt(map[string]string{"question": input, "history": convoHistory})
	if err != nil {
		return
	}

	response, err := C.chatModel.Chat(ctx, []model.ChatMessage{{Role: model.ChatMessageRoleUser, Content: instructionPrompt}}, options...)
	if err != nil {
		return
	}
	errUnmarshall := json.Unmarshal([]byte(response.Content), &output)
	if errUnmarshall != nil {
		return
	}

	return
}

// answerFromDoc based on the given context, will retrieve data and use it to answer question using llm
// this one off query is the key to make this more cost effective and save token usage
func (C *ConversationalRetrievalChain) answerFromDoc(ctx context.Context, context answerOrLookupOutput, doc string, options ...func(*model.Option)) (output string, err error) {
	// cut to max token
	if C.textSplitter.Len(doc) > C.maxToken {
		doc = C.textSplitter.SplitText(doc, C.maxToken, 0)[0]
	}

	b, err := json.Marshal(context)
	if err != nil {
		log.Println(err)
		return
	}

	instructionPrompt, err := C.answerTemplate.FormatPrompt(map[string]string{"doc": doc, "context": string(b)})
	if err != nil {
		log.Print(err)
		return
	}
	message := model.ChatMessage{
		Role:    model.ChatMessageRoleUser,
		Content: instructionPrompt,
	}
	AIResponse, err := C.chatModel.Chat(ctx, []model.ChatMessage{message}, options...)
	if err != nil {
		return
	}

	output = AIResponse.Content

	return
}

// AppendMemory to add conversation to the memory
func (C *ConversationalRetrievalChain) AppendToMemory(message model.ChatMessage) {
	C.memory = append(C.memory, message)
}
