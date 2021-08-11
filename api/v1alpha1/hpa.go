package v1alpha1

type BuiltinAutoScaler struct {
	Type string `json:"type,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Config *Config `json:"config,omitempty"`
}

type
