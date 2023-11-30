package enums

// 上报类型
type ReportType string

const (
	ReportTypeSource      ReportType = "source"      // 请求发起方
	ReportTypeDestination ReportType = "destination" // 请求目标方
)
