# Serverfull - A Lambda Simulator For Go
[![GoDoc](https://godoc.org/github.com/asecurityteam/serverfull?status.svg)](https://godoc.org/github.com/asecurityteam/serverfull)

[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=bugs)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=code_smells)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=coverage)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=duplicated_lines_density)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=ncloc)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=alert_status)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=security_rating)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=sqale_index)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=asecurityteam_serverfull&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=asecurityteam_serverfull)

<!-- TOC -->autoauto- [Serverfull - A Lambda Simulator For Go](#serverfull---a-lambda-simulator-for-go)auto    - [Overview](#overview)auto    - [Quick Start](#quick-start)auto    - [Features And Limitations](#features-and-limitations)auto        - [HTTP API Options](#http-api-options)auto        - [Function Loaders](#function-loaders)auto        - [Running In Mock Mode](#running-in-mock-mode)auto        - [Building Lambda Binaries](#building-lambda-binaries)auto    - [Configuration](#configuration)auto    - [Status](#status)auto    - [Planned/Proposed Features](#plannedproposed-features)auto    - [Contributing](#contributing)auto        - [Building And Testing](#building-and-testing)auto        - [License](#license)auto        - [Contributing Agreement](#contributing-agreement)autoauto<!-- /TOC -->

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
    "github.com/asecurityteam/settings/v2"
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
    if err := serverfull.StartHTTP(ctx, source, fetcher); err != nil {
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

## Features And Limitations

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

### Function Loaders

The project currently only supports using a static mapping of functions. A future
feature we are considering is the addition of the `CreateFunction` and
`UpdateFunction` API endpoints that would leverage S3 for persistence and loading.
This would enable teams who want to continue using the AWS CLI for managing
deployments to do so.

### Running In Mock Mode

Mock mode inspects the signatures of each function being served and runs a
version that only returns an empty version of the output and `nil` as the error.
Systems that want to leverage mock mode will need to do something like this:

```go
if mockMode {
    // Start the mock runtime.
    if err := serverfull.StartHTTPMock(ctx, source, fetcher); err != nil {
        panic(err.Error())
    }
    return
}
// Start the real runtime.
if err := serverfull.StartHTTP(ctx, source, fetcher); err != nil {
    panic(err.Error())
}
```

### Building Lambda Binaries

In the same manner that you can enable mock mode you can also enable a native
lambda mode. The lambda mode causes the runtime to bypass the HTTP server that
implements the Invoke API and runs `lambda.Start()` on a target function
instead. This is provided for teams who want to build binaries that are
compatible with AWS Lambda. To do this you would need to have something like:

```go
if lambdaMode {
    // Start the lambda mode runtime.
    if err := serverfull.StartLambda(ctx, source, fetcher, "myFunctionName"); err != nil {
        panic(err.Error())
    }
    return
}
// Start the serverfull mode runtime.
if err := serverfull.StartHTTP(ctx, source, fetcher); err != nil {
    panic(err.Error())
}
```

The native lambda build can only target a single function at a time so you _must_
specify the target function name. The resulting binary, when executed, will use
whatever loading strategy you've defined to fetch a function matching the name
defined as `TargetFunction` and pass it to the AWS lambda SDK `lambda.Start()`
method. This recreates exactly what a "normal" lambda build would do except that
there is a loading step. To keep consistent with lambda expectations we recommend
only using the static loader for this process because the loading will be done at
startup time.

The lambda build mode also supports running the function in mock mode by
using `StartLambdaMock`.

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

## Status

This project is in incubation which means we are not yet operating this tool in
production and the interfaces are subject to change.

## Planned/Proposed Features

-   Replication of AWS CloudWatch metrics for lambda when running in HTTP mode.
-   Automated injection of the log and stat clients into the context when running in
    lambda mode.
-   Ability to trigger errors in mock mode.
-   Ability to provide static or random values for mock outputs instead of only zero
    values.

## Contributing

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

### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a patch. If
you are an individual you can fill out the [individual
CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the [corporate
CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
