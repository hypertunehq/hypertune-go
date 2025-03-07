package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"

	_ "github.com/hypertunehq/hypertune-go"
	"github.com/urfave/cli/v3"
	"golang.org/x/mod/modfile"
)

const hypertuneGoPath = "github.com/hypertunehq/hypertune-go"

var (
	edgeBaseURL     = "https://edge.hypertune.com"
	sdkVersion      = ""
	outputFileDir   = "generated"
	queryFilePath   = ""
	token           = ""
	branchName      = "main"
	packageName     = "hypertune"
	includeToken    = false
	includeInitData = false
)

func main() {
	cmd := &cli.Command{
		Name:  "hypertune-go-gen",
		Usage: "Tool for generating Hypertune client code for Go.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "edgeBaseURL",
				Hidden:      true,
				Value:       edgeBaseURL,
				Destination: &edgeBaseURL,
			},
			&cli.StringFlag{
				Name:        "sdkVersion",
				Usage:       "Version of the hypertune-go package your app uses. Defaults to value specified in the go.mod file if present.",
				Destination: &sdkVersion,
			},
			&cli.StringFlag{
				Name:        "outputFileDir",
				Usage:       "The directory to write the generated files to.",
				Value:       outputFileDir,
				Destination: &outputFileDir,
			},
			&cli.StringFlag{
				Name:        "queryFilePath",
				Usage:       "File path to the GraphQL initialization query.",
				Destination: &queryFilePath,
			},
			&cli.StringFlag{
				Name:        "token",
				Sources:     cli.EnvVars("HYPERTUNE_TOKEN"),
				Usage:       "Project token.",
				Destination: &token,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "branchName",
				Sources:     cli.EnvVars("HYPERTUNE_BRANCH_NAME"),
				Usage:       "Project branch to use.",
				Value:       branchName,
				Destination: &branchName,
			},
			&cli.StringFlag{
				Name:        "packageName",
				Usage:       "Package name to use in the generated code.",
				Value:       packageName,
				Destination: &packageName,
			},
			&cli.BoolFlag{
				Name:        "includeToken",
				Usage:       "Include the project token in the generated code.",
				Value:       includeToken,
				Destination: &includeToken,
			},
			&cli.BoolFlag{
				Name:        "includeInitData",
				Usage:       "Embed a static snapshot of your flag logic in the generated code so the SDK can reliably, locally and instantly initialize first, before fetching the latest logic from the server, and can function even if the server is unreachable.",
				Value:       includeInitData,
				Destination: &includeInitData,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return run(ctx)
		},
	}
	context, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Run(context, os.Args); err != nil {
		log.Fatal(err.Error())
	}
}

func run(ctx context.Context) error {
	if outputFileDir == "" {
		return errors.New("outputFileDir must be specified")
	}
	if packageName == "" {
		return errors.New("packageName must be specified")
	}

	if err := parseSdkVersion(); err != nil {
		return fmt.Errorf("failed to parse sdk version: %w", err)
	}

	var query *string
	if queryFilePath != "" {
		queryContentBytes, err := os.ReadFile(queryFilePath)
		if err != nil {
			return fmt.Errorf("failed to read query file: %w", err)
		}
		queryContent := string(queryContentBytes)
		query = &queryContent
	}

	codegenUrl, err := url.Parse(edgeBaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse edgeBaseURL: %w", err)
	}
	response, err := doCodegenRequest(ctx, codegenUrl, query)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputFileDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	for _, fileResp := range response.Files {
		if filepath.Ext(fileResp.Name) == ".go" {
			formattedContent, err := format.Source([]byte(fileResp.Content))
			if err == nil {
				fileResp.Content = string(formattedContent)
			}
		}
		filePath := filepath.Join(outputFileDir, fileResp.Name)
		if err = os.WriteFile(filePath, []byte(fileResp.Content), os.ModePerm); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fileResp.Name, err)
		}
		fmt.Printf("Created file %s\n", filePath)
	}
	for _, message := range response.Messages {
		if message.Metadata != nil {
			meta, _ := json.MarshalIndent(message.Metadata, "", "  ")
			fmt.Printf("%s %s %+s\n", message.Level, message.Level, string(meta))
		} else {
			fmt.Printf("%s %s\n", message.Level, message.Level)
		}
	}
	return nil
}

func parseSdkVersion() error {
	if sdkVersion != "" {
		return nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	var fileContent []byte
	for dir != "/" {
		fileContent, err = os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			break
		}
		dir = filepath.Dir(dir)
	}
	if fileContent == nil {
		return fmt.Errorf("go.mod file not found please run the command from a directory containing go.mod file or specify --sdkVersion manually")
	}
	mod, err := modfile.Parse("go.mod", fileContent, nil)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod file: %w", err)
	}
	index := slices.IndexFunc(mod.Require, func(r *modfile.Require) bool {
		return r.Mod.Path == hypertuneGoPath
	})
	if index == -1 {
		return fmt.Errorf("%s doesn't contain %s package", filepath.Join(dir, "go.mod"), hypertuneGoPath)
	}
	sdkVersion = strings.TrimPrefix(mod.Require[index].Mod.Version, "v")
	return nil
}

func doCodegenRequest(ctx context.Context, codegenUrl *url.URL, query *string) (*CodegenResponse, error) {
	urlQuery := codegenUrl.Query()
	body, err := json.Marshal(map[string]any{
		"sdkType":         "go",
		"language":        "go",
		"query":           query,
		"includeToken":    includeToken,
		"includeFallback": includeInitData,
		"sdkVersion":      sdkVersion,
		"packageName":     packageName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	urlQuery.Set("body", url.QueryEscape(string(body)))
	urlQuery.Set("token", url.QueryEscape(token))
	urlQuery.Set("branch", branchName)

	codegenUrl.RawQuery = urlQuery.Encode()
	codegenUrl.Path = "/codegen"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, codegenUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize codegen request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make codegen request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ = io.ReadAll(resp.Body)
		return nil, fmt.Errorf("codegen request failed with code: %d and response: %s", resp.StatusCode, string(body))
	}

	var response CodegenResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode codegen response: %w", err)
	}
	return &response, nil
}

type CodegenResponse struct {
	Files    []CodegenFile    `json:"files"`
	Messages []CodegenMessage `json:"messages"`
}

type CodegenFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type CodegenMessage struct {
	Level    LogLevel `json:"level"`
	Message  string   `json:"message"`
	Metadata any      `json:"metadata"`
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "Debug"
	LogLevelError LogLevel = "Error"
	LogLevelInfo  LogLevel = "Info"
	LogLevelWarn  LogLevel = "Warn"
)
