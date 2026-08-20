package models

// defines the entire execution specification for the pipeline
type JobConfig struct {
	Sources         []SourceDef          `json:"sources"`
	Validations     []ValidationRule     `json:"validations"`
	Transformations []TransformationRule `json:"transformations"`
	Aggregations    []AggregationDef     `json:"aggregations"`
	ExportTargets   []ExportTarget       `json:"export_targets,omitempty"`
	Concurrency     ConcurrencyOptions   `json:"concurrency"`
}

type ExportTarget struct {
	Type   string `json:"type" example:"database"` // "database", "csv", "json"
	Target string `json:"target,omitempty" example:"job_exported_records"`
}

type SourceDef struct {
	Type          string `json:"type" example:"json"`
	URL           string `json:"url" example:"https://randomuser.me/api/?results=10"`
	JSONArrayPath string `json:"json_array_path,omitempty" example:"results"`
}

type ValidationRule struct {
	Field    string   `json:"field" example:"gender"`
	Rule     string   `json:"rule" example:"not_empty"` // "not_empty", "is_numeric", "range"
	MinValue *float64 `json:"min_value,omitempty" example:"0"`
	MaxValue *float64 `json:"max_value,omitempty" example:"100"`
}

type TransformationRule struct {
	Field        string `json:"field" example:"Weight(Pounds)"`
	Action       string `json:"action" example:"convert_to_float"` // "convert_to_int", "convert_to_float", "fill_empty"
	DefaultValue string `json:"default_value,omitempty" example:"0"`
	FillWithAvg  bool   `json:"fill_with_avg,omitempty" example:"false"`
}

type AggregationDef struct {
	Type       string `json:"type" example:"count"` // "sum", "average", "count"
	Field      string `json:"field" example:"gender"`
	OutputName string `json:"output_name" example:"total_users"`
}

type ConcurrencyOptions struct {
	ValidationWorkers int `json:"validation_workers" example:"5"`
	TransformWorkers  int `json:"transform_workers" example:"5"`
}
