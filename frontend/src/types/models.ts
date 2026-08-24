export type JobStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';

export interface SourceDef {
  url: string;
  type?: string;
  json_array_path?: string;
}

export interface ValidationRule {
  field: string;
  rule: string;
  min_value?: number;
  max_value?: number;
}

export interface TransformationRule {
  field: string;
  action: string;
  default_value?: string;
  fill_with_avg?: boolean;
}

export interface AggregationDef {
  type: string;
  field: string;
  output_name: string;
}

export interface ExportTarget {
  type: string;
  target?: string;
}

export interface ConcurrencyOptions {
  validation_workers: number;
  transform_workers: number;
}

export interface JobConfig {
  sources: SourceDef[];
  validations?: ValidationRule[];
  transformations?: TransformationRule[];
  aggregations?: AggregationDef[];
  export_targets?: ExportTarget[];
  concurrency?: ConcurrencyOptions;
}

export interface Job {
  id: string;
  config: JobConfig;
  status: JobStatus;
  total_records: number;
  processed_records: number;
  error_count: number;
  metrics: Record<string, any>;
  created_at: string;
  updated_at: string;
  finished_at?: string;
}

export interface ProgressResponse {
  status: JobStatus;
  percent_complete: number;
  processed_records: number;
  records_per_second: number;
  error_count: number;
  stage_latencies: Record<string, string>;
  start_time: string;
  end_time?: string;
}

export interface ExportResponse {
  status: string;
  urls: string[];
}
