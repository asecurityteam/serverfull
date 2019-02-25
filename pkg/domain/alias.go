package domain

import (
	"github.com/asecurityteam/runhttp"
	"github.com/aws/aws-lambda-go/lambda"
)

// Logger is an alias for the chosen project logging library
// which is, currently, logevent. All references in the project
// should be to this name rather than logevent directly.
type Logger = runhttp.Logger

// LogFn extracts a logger from the context.
type LogFn = runhttp.LogFn

// Stat is an alias for the chosen project metrics library
// which is, currently, xstats. All references in the project
// should be to this name rather than xstats directly.
type Stat = runhttp.Stat

// StatFn extracts a metrics client from the context.
type StatFn = runhttp.StatFn

// Handler is an executable lambda function and is an alias
// for the type of the same name in the AWS Lambda SDK for
// go.
type Handler = lambda.Handler
