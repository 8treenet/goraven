package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
)

func newMainRunner(main *MainAgent, runnerAgent *adk.ChatModelAgent, enableStreaming bool) *MainRunner {
	result := &MainRunner{}

	return result
}

func RegisterRunnerHold(sessionId string) {
}

func ClearRunnerHold(sessionId string) {
}

func RegisterRunner(sessionId string, runner *MainRunner) {
}

func GetRunner(sessionId string) (*MainRunner, bool) {
	return nil, false
}

func DeleteRunner(sessionId string) {
}

type MainRunner struct {
}

type RunnerCompleteEvent struct {
	Reply string
}

func (runner *MainRunner) Terminat() {
}

func (runner *MainRunner) IsStopped() bool {
	return true
}

func (runner *MainRunner) GetFetchStatus() (status int) {

	return 0
}

func (runner *MainRunner) StopFetchStatus() {
}

func (runner *MainRunner) StartFetch() <-chan *SSEEvent {
	return nil
}

func (runner *MainRunner) OnComplete(onComplete func(*RunnerCompleteEvent)) {
}

func (runner *MainRunner) Query(runctx context.Context, content string) (err error) {
	return
}
