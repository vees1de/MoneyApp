export type EnrollmentStatus =
  | 'enrolled'
  | 'in_progress'
  | 'completed'
  | 'canceled'
  | string;

export interface Enrollment {
  id: string;
  course_id: string;
  user_id: string;
  assignment_id?: string | null;
  source: string;
  status: EnrollmentStatus;
  enrolled_at: string;
  started_at?: string | null;
  completed_at?: string | null;
  deadline_at?: string | null;
  last_activity_at?: string | null;
  completion_percent: string;
  is_mandatory: boolean;
  created_at: string;
  updated_at: string;
}
