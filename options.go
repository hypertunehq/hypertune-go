package hypertune

import "time"

type sdkOptions struct {
	BranchName              *string
	InitDataRefreshInterval *time.Duration
	LogsFlushInterval       *time.Duration
	EdgeBaseURL             *string
	RemoteLoggingBaseURL    *string
}

type Option = func(*sdkOptions)

// WithBranchName controls the branch name the SDK will use to fetch data from.
// The default value is "main".
func WithBranchName(branchName string) Option {
	return func(options *sdkOptions) {
		options.BranchName = &branchName
	}
}

// WithInitDataRefreshInterval controls how often the SDK will check for updates.
// Default is 2 seconds. If you set this to 0, the SDK will not check for updates.
func WithInitDataRefreshInterval(interval time.Duration) Option {
	return func(options *sdkOptions) {
		options.InitDataRefreshInterval = &interval
	}
}

// WithLogsFlushInterval controls how often the SDK will automatically flush logs
// in the background. The default value is 2 seconds. If you set this to 0,
// the SDK will not flush logs in the background. This can still be done by manually
// calling `node.FlushLogs()`.
func WithLogsFlushInterval(interval time.Duration) Option {
	return func(options *sdkOptions) {
		options.LogsFlushInterval = &interval
	}
}

func WithEdgeBaseURL(url string) Option {
	return func(options *sdkOptions) {
		options.EdgeBaseURL = &url
	}
}

func WithRemoteLoggingBaseURL(baseUrl string) Option {
	return func(options *sdkOptions) {
		options.RemoteLoggingBaseURL = &baseUrl
	}
}

func parseOptions(opts []Option) map[string]any {
	options := &sdkOptions{}
	for _, option := range opts {
		option(options)
	}

	configMap := map[string]interface{}{
		"language": language,
	}

	if options.BranchName != nil {
		configMap["branch_name"] = *options.BranchName
	}
	if options.InitDataRefreshInterval != nil {
		configMap["init_data_refresh_interval_ms"] = *options.InitDataRefreshInterval / time.Millisecond
	}
	if options.LogsFlushInterval != nil {
		configMap["logs_flush_interval_ms"] = *options.LogsFlushInterval / time.Millisecond
	}
	if options.EdgeBaseURL != nil {
		configMap["edge_base_url"] = *options.EdgeBaseURL
	}
	if options.RemoteLoggingBaseURL != nil {
		configMap["remote_logging_base_url"] = *options.RemoteLoggingBaseURL
	}
	return configMap
}
