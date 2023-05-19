package chain

import (
	"context"

	"github.com/wejick/gochain/model"
)

//go:generate mockery --name BaseChain
type BaseChain interface {
	// Run does prediction of input of <string,string> and produce output of <string,string>
	// map of <string,string> of output and input to accomodate many possible usecases
	Run(ctx context.Context, input map[string]string, options ...func(*model.Option)) (output map[string]string, err error)

	// SimpleRun does prediction of input of string and produce output of string
	// this is to accomodate simple input / output usage
	SimpleRun(ctx context.Context, input string, options ...func(*model.Option)) (output string, err error)
}