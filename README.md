<a id="markdown-serverfull---a-lambda-simulator-for-go" name="serverfull---a-lambda-simulator-for-go"></a>
# Serverfull - A Lambda Simulator For Go

<!-- TOC -->

- [Serverfull - A Lambda Simulator For Go](#serverfull---a-lambda-simulator-for-go)
    - [Overview](#overview)
    - [Quick Start](#quick-start)
    - [Features And Limitations](#features-and-limitations)
    - [Configuration](#configuration)
    - [Status](#status)
    - [Contributing](#contributing)
        - [Building And Testing](#building-and-testing)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview

This projects is a toolkit for leveraging Lambda functions outside of the usual
serverless offering in AWS. Bundled with the project is an HTTP server that implements
the
[AWS Lambda Invoke API](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html)
and is able to serve any function that is compatible with the
[lambda go sdk](https://github.com/aws/aws-lambda-go).

Generally speaking, if you want to use AWS Lambda then you should probably use
AWS Lambda rather this simulator. This project is specialized and was conceived
to enable:

-   Teams who want to adopt a serverless style of development but without
    full access to a serverless/FaaS provider such as those who are developing
    on bare metal or private cloud infrastructures.

-   Teams who are ready to migrate away from AWS Lambda but aren't ready to
    rewrite large portions of existing code such as those who are either changing
    cloud providers or moving to EC2 in order to fine-tune the performance
    characteristics of their runtimes.

If you're looking for a local development tool that supports AWS Lambda then you
would likely be better served by using tools like
[docker-lambda](https://github.com/lambci/docker-lambda) and
[AWS SAM Local](https://aws.amazon.com/blogs/aws/new-aws-sam-local-beta-build-and-test-serverless-applications-locally/).

<a id="markdown-quick-start" name="quick-start"></a>
## Quick Start

Start by defining a normal lambda function. For example, here is one from the AWS Go SDK project:

```golang
package main

// hello is lifted straight from the aws-lambda-go README.md file.
// This is can be called like:
//
//		curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
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

	serverfull "github.com/asecurityteam/serverfull/pkg"
	"github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/asecurityteam/settings"
)

func hello() (string, error) {
	return "Hello ƛ!", nil
}

func main() {
	ctx := context.Background()
	handlers := map[string]domain.Handler{
			// The keys of this map represent the function name and will be
			// accessed using the URL parameter of the Invoke API call.
			// These names are arbitrary and user defined. They do not need
			// to match the name of the code function. Think of these map keys
			// as being the ARN of a lambda.
			"hello":    lambda.NewHandler(hello),
			// Note that we use the AWS Lambda SDK for Go here to wrap the function.
			// This is because the handlers are _actual_ Lambda functions and
			// are 100% compatible with AWS Lambda.
	}
	// Create any implementation of the settings.Source interface. Here we
  	// use the environment variable source.
  	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	// Create a new instance of the runtime using the given source for
	// configuration values and the given static mapping of Lambda functions.
	rt, err := serverfull.NewStatic(ctx, source, handlers)
	if err != nil {
		panic(err.Error())
	}
	if err := rt.Run(); err != nil {
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

The `X-Amz-Invocation-Type` header can be used, as described in the actual
[AWS Lambda Invoke API](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html),
to alter the execution behavior. The only aspects of the API that are not
simulated yet are:

-   The "Tail" option for the LogType header does not cause the
    response to include partial logs.

-   The "Qualifier" parameter is currently ignored and the reported
    execution version is always "latest".

-   The "Function-Error" header is always "Unhandled" in the event
    of an exception.

The API is compatible enough with AWS Lambda that the AWS CLI, as well
as all AWS SDKs that support Lambda features, can be used after adjusting
the endpoint value.

Currently, this project performs no scheduling or orchestration of Lambda functions
across multiple hosts. Instead it relies on a static mapping of Lambda functions that
are exposed by each host running the service. There are no plans to add dynamic
scheduling at this time.

<a id="markdown-configuration" name="configuration"></a>
## Configuration

This project uses [settings](https://github.com/asecurityteam/settings) for managing
configuration values and [runhttp](https://github.com/asecurityteam/runhttp) to
manage the runtime behaviors such as logs, metrics, and shutdown signaling. All of
the configuration described in the `runhttp` README is valid with the notable exception
that this project adds a `serverfull` path prefix to all lookups. This means where
`runhttp` will have a `RUNTIME_LOGGING_LEVEL` variable then this project will have
a `SERVERFULL_RUNTIME_LOGGING_LEVEL` variable.

For more advanced changes we recommend you use the `NewRouter` and `NewStatic` methods
as examples of how the system is composed. To add features such as authentication,
additional metrics, retries, or other features suitable as middleware we recommend
you use a project like [transportd](https://github.com/asecurityteam/transportd) or
a more in-depth "service-mesh" type of proxy rather than modifying this project
directly.

<a id="markdown-status" name="status"></a>
## Status

This project is in incubation which means we are not yet operating this tool in production
and the interfaces are subject to change.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

<a id="markdown-building-and-testing" name="building-and-testing"></a>
### Building And Testing

We publish a docker image called [SDCLI](https://github.com/asecurityteam/sdcli) that
bundles all of our build dependencies. It is used by the included Makefile to help make
building and testing a bit easier. The following actions are available through the Makefile:

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

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
