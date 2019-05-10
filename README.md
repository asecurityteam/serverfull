<a id="markdown-serverfull---a-lambda-simulator-for-go" name="serverfull---a-lambda-simulator-for-go"></a>
# Serverfull - A Lambda Simulator For Go

<!-- TOC -->

- [Serverfull - A Lambda Simulator For Go](#serverfull---a-lambda-simulator-for-go)
    - [Overview](#overview)
    - [Quick Start](#quick-start)
    - [Features And Limitations](#features-and-limitations)
        - [HTTP API Options](#http-api-options)
        - [Function Loaders](#function-loaders)
        - [Running In Mock Mode](#running-in-mock-mode)
        - [Building Lambda Binaries](#building-lambda-binaries)
    - [Configuration](#configuration)
    - [Status](#status)
    - [Planned/Proposed Features](#plannedproposed-features)
    - [Contributing](#contributing)
        - [Building And Testing](#building-and-testing)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview

This projects is a toolkit for leveraging Lambda functions outside of the usual
serverless offering in AWS. Bundled with the project is an HTTP server that
implements the [AWS Lambda Invoke
API](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html) and is able to
serve any function that is compatible with the [lambda go
sdk](https://github.com/aws/aws-lambda-go).

Generally speaking, if you want to use AWS Lambda then you should probably use AWS
Lambda rather this simulator. This project is specialized and was conceived to
enable:

-   Teams who want to adopt a serverless style of development but without full access
    to a serverless/FaaS provider such as those who are developing on bare metal or
    private cloud infrastructures.

-   Teams who are ready to migrate away from AWS Lambda but aren't ready to rewrite
    large portions of existing code such as those who are either changing cloud
    providers or moving to EC2 in order to fine-tune the performance characteristics
    of their runtimes.

If you're looking for a local development tool that supports AWS Lambda then you
would likely be better served by using tools like
[docker-lambda](https://github.com/lambci/docker-lambda) and [AWS SAM
Local](https://aws.amazon.com/blogs/aws/new-aws-sam-local-beta-build-and-test-serverless-applications-locally/).

<a id="markdown-quick-start" name="quick-start"></a>
## Quick Start

Start by defining a normal lambda function. For example, here is one from the AWS Go
SDK project:

```golang
package main

// hello is lifted straight from the aws-lambda-go README.md file.
// This is can be called like:
//
//      curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
func hello() (string, error) {
    return "Hello ƛ!", nil
}
```

Then attach that function to map and generate an instance of Serverfull:

```golang
package main

import (
    "context"
    "fmt"

    "github.com/asecurityteam/serverfull"
    "github.com/asecurityteam/settings"
)

func hello() (string, error) {
    return "Hello ƛ!", nil
}

func main() {
    ctx := context.Background()
    functions := map[string]serverfull.Function{
            // The keys of this map represent the function name and will be
            // accessed using the URL parameter of the Invoke API call.
            // These names are arbitrary and user defined. They do not need
            // to match the name of the code function. Think of these map keys
            // as being the ARN of a lambda.
            "hello":    serverfull.NewFunction(hello),
            // Note that we use the AWS Lambda SDK for Go here to wrap the function.
            // This is because the serverfull functions are _actual_ Lambda functions
            // and are 100% compatible with AWS Lambda.
    }
    // Create any implementation of the settings.Source interface. Here we
    // use the environment variable source.
    source, err := settings.NewEnvSource(os.Environ())
    if err != nil {
        panic(err.Error())
    }
    // Wrap the map in a static function loader. Other loading options are
    // discussed later in the docs.
    fetcher := &serverfull.StaticFetcher{Functions:functions}
    // Start the runtime.
    if err := serverfull.Start(ctx, source, fetcher); err != nil {
        panic(err.Error())
    }
}
```

If you run this code then you can make a request to invoke the `hello` function and
see the result.

```sh
curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
```

Alternatively, the AWS CLI may be used as well:

```sh
aws lambda invoke \
    --endpoint-url="http://localhost:8080" \
    --function-name="hello" \
    output.txt && \
    cat output.txt && \
    rm output.txt
```

<a id="markdown-features-and-limitations" name="features-and-limitations"></a>
## Features And Limitations

<a id="markdown-http-api-options" name="http-api-options"></a>
### HTTP API Options

The `X-Amz-Invocation-Type` header can be used, as described in the actual [AWS
Lambda Invoke API](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html), to
alter the execution behavior. The only aspects of the API that are not simulated yet
are:

-   The "Tail" option for the LogType header does not cause the response to include
    partial logs.

-   The "Qualifier" parameter is currently ignored and the reported execution version
    is always "latest".

-   The "Function-Error" header is always "Unhandled" in the event of an exception.

The API is compatible enough with AWS Lambda that the AWS CLI, as well as all AWS
SDKs that support Lambda features, can be used after adjusting the endpoint value.

<a id="markdown-function-loaders" name="function-loaders"></a>
### Function Loaders

The project currently only supports using a static mapping of functions. A future
feature we are considering is the addition of the `CreateFunction` and
`UpdateFunction` API endpoints that would leverage S3 for persistence and loading.
This would enable teams who want to continue using the AWS CLI for managing
deployments to do so.

<a id="markdown-running-in-mock-mode" name="running-in-mock-mode"></a>
### Running In Mock Mode

When building a binary with `go build` or running with `go run` you can pass in a
build flag that swaps out all lambda functions registered with the server with a
mocked version. The mock version will always return an empty version of the output
struct and a `nil` error if either of those are returned from the source function. To
enable this mode use:

```bash
go build \
    -ldflags "-X github.com/asecurityteam/serverfull.MockMode=true" \
    main.go
```

Alternatively, if you prefer to manage these values through a runtime source, like
environment variables, then you have a few options. The project exports a global flag
variable called `serverfull.MockMode` which is same variable modified by the build
flag. It defaults to an empty string to indicate that mocking is disabled. Rather
than using the build flag you can set that value to any non-empty string to enable
the mock mode. For example:

```golang
serverfull.MockMode = os.Getenv("RUN_IN_MOCK_MODE")
serverfull.Start(ctx, source, fetcher) // Now running in mock mode.
```

If you would prefer not to rely on process wide flags then you also have the option
of using individual runtime functions directly. There are methods called
`serverfull.StartHTTPMock()` and `serverfull.StartHTTP()` which rely on no flags. For
example:

```golang
if os.Getenv("RUN_IN_MOCK_MODE") == "" {
    serverfull.StartHTTP(ctx, source, fetcher)
} else {
    serverfull.StartHTTPMock(ctx, source, fetcher)
}
```

<a id="markdown-building-lambda-binaries" name="building-lambda-binaries"></a>
### Building Lambda Binaries

In the same manner that you can enable mock mode you can also enable a native lambda
mode. The lambda mode causes the runtime to bypass the HTTP server that implements
the Invoke API and runs `lambda.Start()` on a target function instead. This is
provided for teams who want to build binaries that are compatible with AWS Lambda. To
do this you must pass in two build flags:

```bash
go build \
    -ldflags "-X github.com/asecurityteam/serverfull.BuildMode=lambda -X github.com/asecurityteam/serverfull.TargetFunction=hello" \
    main.go
```

The native lambda build can only target a single function at a time so you _must_
specify the target function name. The resulting binary, when executed, will use
whatever loading strategy you've defined to fetch a function matching the name
defined as `TargetFunction` and pass it to the AWS lambda SDK `lambda.Start()`
method. This recreates exactly what a "normal" lambda build would do except that
there is a loading step. To keep consistent with lambda expectations we recommend
only using the static loader for this process because the loading will be done at
startup time.

Just like the mock flag for the HTTP runtime, you may bypass the build flag in one of
two ways if you want to manage the behavior switch differently. The first alternative
is modifying the global flags with:

```golang
serverfull.BuildMode = serverfull.BuildModeLambda
serverfull.TargetFunction = os.Getenv("TARGET_FUNCTION")
serverfull.Start(ctx, source, fetcher)
```

The second alternative is to use the specific lambda runtime methods directly:

```golang
targetFunction := os.Getenv("TARGET_FUNCTION")
if os.Getenv("RUN_IN_LAMBDA_MODE") == "" {
    serverfull.StartLambda(ctx, source, fetcher, targetFunc)
} else {
    serverfull.StartHTTP(ctx, source, fetcher)
}
```

The lambda build mode also supports running the function in mock mode using exactly
the same build flags and variables as demonstrated in the mock mode section.

<a id="markdown-configuration" name="configuration"></a>
## Configuration

This project uses [settings](https://github.com/asecurityteam/settings) for managing
configuration values and [runhttp](https://github.com/asecurityteam/runhttp) to
manage the runtime behaviors such as logs, metrics, and shutdown signaling when
running or building in HTTP mode. All of the configuration described in the `runhttp`
README is valid with the notable exception that this project adds a `serverfull` path
prefix to all lookups. This means where `runhttp` will have a `RUNTIME_LOGGING_LEVEL`
variable then this project will have a `SERVERFULL_RUNTIME_LOGGING_LEVEL` variable.

For more advanced changes we recommend you use the `NewRouter` and `Start` methods as
examples of how the system is composed. To add features such as authentication,
additional metrics, retries, or other features suitable as middleware we recommend
you use a project like [transportd](https://github.com/asecurityteam/transportd) or a
more in-depth "service-mesh" type of proxy rather than modifying this project
directly.

<a id="markdown-status" name="status"></a>
## Status

This project is in incubation which means we are not yet operating this tool in
production and the interfaces are subject to change.

<a id="markdown-plannedproposed-features" name="plannedproposed-features"></a>
## Planned/Proposed Features

-   Replication of AWS CloudWatch metrics for lambda when running in HTTP mode.
-   Automated injection of the log and stat clients into the context when running in
    lambda mode.
-   Ability to trigger errors in mock mode.
-   Ability to provide static or random values for mock outputs instead of only zero
    values.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

<a id="markdown-building-and-testing" name="building-and-testing"></a>
### Building And Testing

We publish a docker image called [SDCLI](https://github.com/asecurityteam/sdcli) that
bundles all of our build dependencies. It is used by the included Makefile to help
make building and testing a bit easier. The following actions are available through
the Makefile:

-   make dep

    Install the project dependencies into a vendor directory

-   make lint

    Run our static analysis suite

-   make test

    Run unit tests and generate a coverage artifact

-   make integration

    Run integration tests and generate a coverage artifact

-   make coverage

    Report the combined coverage for unit and integration tests

<a id="markdown-license" name="license"></a>
### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a patch. If
you are an individual you can fill out the [individual
CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the [corporate
CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
