export type JobStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface SourceDef {
  url: string;
  format?: string;
}

export interface ValidationRule {
  field: string;
  type: string;
  required: boolean;
}

export interface TransformationRule {
  field: string;
  target_field: string;
  operation: string;
}

export interface AggregationDef {
  type: string;
  field: string;
}

export interface ExportTarget {
  type: string;
  destination: string;
  format: string;
}

export interface ConcurrencyOptions {
  max_workers: number;
  batch_size: number;
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
