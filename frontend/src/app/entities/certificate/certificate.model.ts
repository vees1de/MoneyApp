export interface Certificate {
  id: string;
  user_id: string;
  course_id?: string | null;
  enrollment_id?: string | null;
  certificate_no?: string | null;
  issued_by?: string | null;
  issued_at?: string | null;
  expires_at?: string | null;
  status: string;
  file_id: string;
  uploaded_at: string;
  verified_at?: string | null;
  verified_by?: string | null;
  notes?: string | null;
}
