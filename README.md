# hypertune-go

A Go SDK for [Hypertune](https://www.hypertune.com).

## Quickstart

First, install the `hypertune-go-gen` tool that generates the type-safe Hypertune Go client by running

```bash
go get -tool github.com/hypertunehq/hypertune-go/cmd/hypertune-go-gen
```

Second, set the `HYPERTUNE_TOKEN` environment variable to your project token which you can find in the Settings panel of your project.

Then, generate the client by running

```bash
go tool hypertune-go-gen --token=${HYPERTUNE_TOKEN} --outputFileDir=pkg/hypertune
```

Alternatively you can add the following go generate directive to one of your go files to automatically re-generate the client when you run `go generate ./...`.

```go
//go:generate go tool hypertune-go-gen --token=${HYPERTUNE_TOKEN} --outputFileDir=pkg/hypertune
```

Finally instantiate the client and start evaluating your flags.

```go
package main

import (
    "fmt"
	"log"
    "os"

    // Update to your project path.
    "github.com/myTeam/myProject/pkg/hypertune"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
    var token = os.Getenv("HYPERTUNE_TOKEN")
	source, err := hypertune.CreateSource(&token)
	if err != nil {
		return err
	}
	defer source.Close()

	source.WaitForInitialization()

	rootNode := source.Root(hypertune.RootArgs{
		Context: hypertune.Context{
			Environment: hypertune.Development,
			User: hypertune.User{
				Id:    "123",
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
		},
	})

	fmt.Printf("ExampleFlag: %v\n", rootNode.ExampleFlag(false))

	return nil
}
```
