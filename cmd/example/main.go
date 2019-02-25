package main

// This example demonstrates how the static HandlerFetcher may be used to
// create an instance of the runtime using direct imports of Lambda functions
// rather than using a dynamic loading system.

import (
	"context"
	"flag"
	"fmt"
	"os"

	serverfull "github.com/asecurityteam/serverfull/pkg"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/aws/aws-lambda-go/lambda"
)

// hello is lifted straight from the aws-lambda-go README.md file.
// This is can be called like:
//
//		curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
func hello() (string, error) {
	return "Hello ƛ!", nil
}

type helloYouInput struct {
	Name string `json:"name"`
}
type helloYouOutput struct {
	Greeting string `json:"greeting"`
}

// helloYou is an added example to show how functions that work with
// structured input and output behave.
// This is can be called like:
//
//		curl --request POST --data '{"name": "me"}' localhost:8080/2015-03-31/functions/helloYou/invocations
//
// The input data may be omitted or invalid in order to generate an error.
func helloYou(ctx context.Context, input helloYouInput) (helloYouOutput, error) {
	name := input.Name
	if name == "" {
		name = "ƛ"
	}
	return helloYouOutput{Greeting: fmt.Sprintf("Hello %s!", name)}, nil
}

func main() {
	handlers := map[string]domain.Handler{
		// The keys of this map represent the function name and will be
		// accessed using the URL parameter of the Invoke API call.
		// These names are arbitrary and user defined. They do not need
		// to match the name of the code function.
		"hello":    lambda.NewHandler(hello),
		"helloYou": lambda.NewHandler(helloYou),
	}

	// Handle the -h flag and print settings.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(serverfull.HelpStatic())
		return
	}

	rt, err := serverfull.NewStatic(handlers)
	if err != nil {
		panic(err.Error())
	}
	if err := rt.Run(); err != nil {
		panic(err.Error())
	}
}
