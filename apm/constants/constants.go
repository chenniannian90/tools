package constants

const (
	ApmRequestTotal    = "apm_request_total"
	ApmRequestDuration = "apm_request_duration"
)

// 我们注入的label
const (
	OperatorID      = "operator_id"
	StatusCode      = "status_code"
	Project         = "project"
	Report          = "report"
	Source          = "source"
	Destination     = "destination"
	DestinationType = "destination_type"
	Protocol        = "protocol"
)

const (
	Unknown               = "unknown"
	DefaultMetricsPath    = "/metrics"
	DefaultMetricsAddress = ":8080"

	MetaSource      = "Meta-Apm-Source"   // 请求 header 中, Meta-Apm-Source 记录发起请求的服务
	MetaOperator    = "Meta-Apm-Operator" // 请求 header 中, Meta-Apm-Operator 记录发起请求的操作
	ValidStatusCode = "-"                 // 如果是 http 协议, 服务器未返回 response 就出错了, 状态码会是这个状态
	ErrStatusCode   = "-1"                // 如果是通过 sdk 请求远端服务, 没有状态码, 只有 err, -1 表示出错; 0 表示正常
	OkStatusCode    = "0"
)

// 千万不要改这个
var LabelList = []string{OperatorID, StatusCode, Project, Report, Source, Destination, DestinationType, Protocol}
