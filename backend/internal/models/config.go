package models

// defines the entire execution specification for the pipeline
type JobConfig struct {
	Sources         []SourceDef          `json:"sources"`
	Validations     []ValidationRule     `json:"validations"`
	Transformations []TransformationRule `json:"transformations"`
	Aggregations    []AggregationDef     `json:"aggregations"`
	Concurrency     ConcurrencyOptions   `json:"concurrency"`
}

type SourceDef struct {
	Type          string `json:"type" example:"csv"` // "csv", "json"
	URL           string `json:"url" example:"https://covid.ourworldindata.org/data/owid-covid-data.csv"`
	JSONArrayPath string `json:"json_array_path,omitempty" example:"$"`
}

type ValidationRule struct {
	Field    string   `json:"field" example:"new_cases"`
	Rule     string   `json:"rule" example:"is_numeric"` // "not_empty", "is_numeric", "range"
	MinValue *float64 `json:"min_value,omitempty" example:"0"`
	MaxValue *float64 `json:"max_value,omitempty" example:"1000"`
}

type TransformationRule struct {
	Field        string `json:"field" example:"new_cases"`
	Action       string `json:"action" example:"fill_empty"`            // "convert_to_int", "fill_empty", "flatten"
	DefaultValue string `json:"default_value,omitempty" example:"0"`    // Optional default value
	FillWithAvg  bool   `json:"fill_with_avg,omitempty" example:"true"` // Dynamically compute average and fill
}

type AggregationDef struct {
	Type       string `json:"type" example:"sum"` // "sum", "average", "count"
	Field      string `json:"field" example:"new_cases"`
	OutputName string `json:"output_name" example:"total_global_new_cases"`
}

type ConcurrencyOptions struct {
	ValidationWorkers int `json:"validation_workers" example:"5"`
	TransformWorkers  int `json:"transform_workers" example:"5"`
}
