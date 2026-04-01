export interface ColumnMeta {
  key: string;
  label: string;
  type: 'string' | 'number' | 'date' | 'currency';
  default: boolean;
}

export interface SourceMeta {
  key: string;
  label: string;
  columns: ColumnMeta[];
}

export interface SourcesResponse {
  sources: SourceMeta[];
}

export interface ExportFilters {
  employee_ids?: string[];
  intake_ids?: string[];
  department_ids?: string[];
  course_ids?: string[];
  category_ids?: string[];
  levels?: string[];
  statuses?: string[];
  price_min?: number;
  price_max?: number;
  price_currency?: string;
  date_from?: string;
  date_to?: string;
  date_field?: string;
  search?: string;
}

export interface SmartExportRequest {
  source: string;
  columns: string[];
  filters?: ExportFilters;
  sort_by?: string;
  sort_dir?: 'asc' | 'desc';
  format?: 'xlsx' | 'csv';
}
