package enums

// 不同服务类型, 监控的指标不一样, 尽量正确填写

// 目标服务类型
type DestinationType string

const (
	DestinationTypeCommon DestinationType = "common" // 普通服务
)
