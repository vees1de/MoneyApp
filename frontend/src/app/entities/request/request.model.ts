export type ExternalRequestStatus =
  | 'draft'
  | 'manager_approval'
  | 'hr_approval'
  | 'approved'
  | 'rejected'
  | 'revision_requested'
  | 'completed'
  | string;

export interface Request {
  id: string;
  request_no: string;
  employee_user_id: string;
  department_id?: string | null;
  title: string;
  provider_name?: string | null;
  course_url?: string | null;
  program_description?: string | null;
  planned_start_date?: string | null;
  planned_end_date?: string | null;
  duration_hours?: string | null;
  cost_amount: string;
  currency: string;
  business_goal?: string | null;
  employee_comment?: string | null;
  manager_comment?: string | null;
  hr_comment?: string | null;
  status: ExternalRequestStatus;
  calendar_conflict_status?: string | null;
  budget_check_status?: string | null;
  approved_at?: string | null;
  rejected_at?: string | null;
  created_at: string;
  updated_at: string;
}
