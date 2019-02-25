<a id="markdown-serverfull---a-lambda-simulator-for-go" name="serverfull---a-lambda-simulator-for-go"></a>
# Serverfull - A Lambda Simulator For Go

*Status: Incubation*

<!-- TOC -->

- [Serverfull - A Lambda Simulator For Go](#serverfull---a-lambda-simulator-for-go)
    - [Overview](#overview)
    - [Using The Static Loader](#using-the-static-loader)
    - [Customizing The Runtime](#customizing-the-runtime)
    - [Contributing](#contributing)
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
[lambda go sdk](https://github.com/aws/aws-lambda-go). The system comes with a few
options for managing Lambda functions ranging from static builds to more dynamic,
push based updates of functions at runtime depending on how many features of AWS
Lambda you need to replicate and how much complexity you can tolerate when operating
the simulator.

Generally speaking, if you want to use AWS Lambda then you should probably use
AWS Lambda rather this simulator. This project is a bit specialized and was conceived
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

<a id="markdown-using-the-static-loader" name="using-the-static-loader"></a>
## Using The Static Loader

The simplest option for using Serverfull is with the static mapping version of the
function loader. This options requires you to create a `main.go` file that
imports Serverfull and creates a new runtime with the static map of functions you
want to serve:

```golang
package main

// This example demonstrates how the static HandlerFetcher may be used to
// create an instance of the runtime using direct imports of Lambda functions
// rather than using a dynamic loading system.

import (
	"context"
	"fmt"

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

func main() {
	handlers := map[string]domain.Handler{
			// The keys of this map represent the function name and will be
			// accessed using the URL parameter of the Invoke API call.
			// These names are arbitrary and user defined. They do not need
			// to match the name of the code function.
			"hello":    lambda.NewHandler(hello),
    }
	rt, err := serverfull.NewStatic(handlers)
	if err != nil {
		panic(err.Error())
	}
	if err := rt.Run(); err != nil {
		panic(err.Error())
	}
}
```

If you run this code, which is also provided for you in the `/cmd/example`
directory, then you can make a request to invoke the `hello` function and
see the result.

```sh
curl --request POST localhost:8080/2015-03-31/functions/hello/invocations
```

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

```sh
aws lambda invoke \
	--endpoint-url="http://localhost:8080" \
	--function-name="hello" \
    output.txt && \
    cat output.txt && \
    rm output.txt
```

<a id="markdown-customizing-the-runtime" name="customizing-the-runtime"></a>
## Customizing The Runtime

This project leverages [runhttp](https://github.com/asecurityteam/runhttp) to manage
the logs, metrics, and runtime behaviors of the HTTP server. Using the
`serverfull.NewStatic()` method will leverage environment variables to configure
the runtime as it is described in the `runhttp` documentation.

For more advanced customization, see the `serverfull.NewStatic()` method as an
example of how to create a custom build that is capable of modifying substantial
aspects of the project such as:

-   Adding custom middleware to the HTTP server
-   Adding custom routes our modifying the default /healthcheck route
-   Installing a custom Lambda loader
-   Using different configuration sources

<a id="markdown-contributing" name="contributing"></a>
## Contributing

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
