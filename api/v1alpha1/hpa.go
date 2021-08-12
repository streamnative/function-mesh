package v1alpha1

type BuiltinHPARule string

const (
	AverageUtilizationCPUPercent80 BuiltinHPARule = "AverageUtilizationCPUPercent80"
	AverageUtilizationCPUPercent50 BuiltinHPARule = "AverageUtilizationCPUPercent50"
	AverageUtilizationCPUPercent20 BuiltinHPARule = "AverageUtilizationCPUPercent20"

	AverageUtilizationMemoryPercent80 BuiltinHPARule = "AverageUtilizationMemoryPercent80"
	AverageUtilizationMemoryPercent50 BuiltinHPARule = "AverageUtilizationMemoryPercent50"
	AverageUtilizationMemoryPercent20 BuiltinHPARule = "AverageUtilizationMemoryPercent20"
)
