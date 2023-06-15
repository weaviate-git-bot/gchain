// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package model

import (
	"context"
	"sync"
)

// Ensure, that ChatModelMock does implement ChatModel.
// If this is not the case, regenerate this file with moq.
var _ ChatModel = &ChatModelMock{}

// ChatModelMock is a mock implementation of ChatModel.
//
//	func TestSomethingThatUsesChatModel(t *testing.T) {
//
//		// make and configure a mocked ChatModel
//		mockedChatModel := &ChatModelMock{
//			CallFunc: func(ctx context.Context, prompt string, options ...func(*Option)) (string, error) {
//				panic("mock out the Call method")
//			},
//			ChatFunc: func(ctx context.Context, messages []ChatMessage, options ...func(*Option)) (ChatMessage, error) {
//				panic("mock out the Chat method")
//			},
//		}
//
//		// use mockedChatModel in code that requires ChatModel
//		// and then make assertions.
//
//	}
type ChatModelMock struct {
	// CallFunc mocks the Call method.
	CallFunc func(ctx context.Context, prompt string, options ...func(*Option)) (string, error)

	// ChatFunc mocks the Chat method.
	ChatFunc func(ctx context.Context, messages []ChatMessage, options ...func(*Option)) (ChatMessage, error)

	// calls tracks calls to the methods.
	calls struct {
		// Call holds details about calls to the Call method.
		Call []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Prompt is the prompt argument value.
			Prompt string
			// Options is the options argument value.
			Options []func(*Option)
		}
		// Chat holds details about calls to the Chat method.
		Chat []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Messages is the messages argument value.
			Messages []ChatMessage
			// Options is the options argument value.
			Options []func(*Option)
		}
	}
	lockCall sync.RWMutex
	lockChat sync.RWMutex
}

// Call calls CallFunc.
func (mock *ChatModelMock) Call(ctx context.Context, prompt string, options ...func(*Option)) (string, error) {
	if mock.CallFunc == nil {
		panic("ChatModelMock.CallFunc: method is nil but ChatModel.Call was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Prompt  string
		Options []func(*Option)
	}{
		Ctx:     ctx,
		Prompt:  prompt,
		Options: options,
	}
	mock.lockCall.Lock()
	mock.calls.Call = append(mock.calls.Call, callInfo)
	mock.lockCall.Unlock()
	return mock.CallFunc(ctx, prompt, options...)
}

// CallCalls gets all the calls that were made to Call.
// Check the length with:
//
//	len(mockedChatModel.CallCalls())
func (mock *ChatModelMock) CallCalls() []struct {
	Ctx     context.Context
	Prompt  string
	Options []func(*Option)
} {
	var calls []struct {
		Ctx     context.Context
		Prompt  string
		Options []func(*Option)
	}
	mock.lockCall.RLock()
	calls = mock.calls.Call
	mock.lockCall.RUnlock()
	return calls
}

// Chat calls ChatFunc.
func (mock *ChatModelMock) Chat(ctx context.Context, messages []ChatMessage, options ...func(*Option)) (ChatMessage, error) {
	if mock.ChatFunc == nil {
		panic("ChatModelMock.ChatFunc: method is nil but ChatModel.Chat was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Messages []ChatMessage
		Options  []func(*Option)
	}{
		Ctx:      ctx,
		Messages: messages,
		Options:  options,
	}
	mock.lockChat.Lock()
	mock.calls.Chat = append(mock.calls.Chat, callInfo)
	mock.lockChat.Unlock()
	return mock.ChatFunc(ctx, messages, options...)
}

// ChatCalls gets all the calls that were made to Chat.
// Check the length with:
//
//	len(mockedChatModel.ChatCalls())
func (mock *ChatModelMock) ChatCalls() []struct {
	Ctx      context.Context
	Messages []ChatMessage
	Options  []func(*Option)
} {
	var calls []struct {
		Ctx      context.Context
		Messages []ChatMessage
		Options  []func(*Option)
	}
	mock.lockChat.RLock()
	calls = mock.calls.Chat
	mock.lockChat.RUnlock()
	return calls
}
