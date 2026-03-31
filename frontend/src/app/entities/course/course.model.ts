export interface Course {
  id: string;
  type: string;
  source_type: string;
  title: string;
  slug?: string | null;
  short_description?: string | null;
  description?: string | null;
  provider_id?: string | null;
  category_id?: string | null;
  direction_id?: string | null;
  level?: string | null;
  duration_hours?: string | null;
  language?: string | null;
  is_mandatory_default: boolean;
  status: string;
  external_url?: string | null;
  price?: string | null;
  price_currency?: string | null;
  next_start_date?: string | null;
  thumbnail_file_id?: string | null;
  created_by?: string | null;
  updated_by?: string | null;
  published_at?: string | null;
  archived_at?: string | null;
  created_at: string;
  updated_at: string;
}
