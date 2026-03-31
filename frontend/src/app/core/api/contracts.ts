export interface CalendarUpcomingEvent {
  id: string;
  title: string;
  start_at: string;
  end_at: string;
  status: string;
  meeting_url?: string | null;
  location?: string | null;
}

export interface AdminRole {
  id: string;
  code: string;
  name: string;
  description?: string | null;
  is_system: boolean;
}

export interface AdminPermission {
  id: string;
  code: string;
  module: string;
  action: string;
  description?: string | null;
}

export interface ExternalRequest {
  id: string;
  request_no: string;
  employee_user_id: string;
  employee_full_name?: string;
  employee_email?: string;
  department_id?: string | null;
  department_name?: string;
  title: string;
  provider_id?: string | null;
  provider_name?: string | null;
  course_url?: string | null;
  program_description?: string | null;
  planned_start_date?: string | null;
  planned_end_date?: string | null;
  duration_hours?: string | null;
  cost_amount?: string;
  currency?: string;
  business_goal?: string | null;
  employee_comment?: string | null;
  manager_comment?: string | null;
  hr_comment?: string | null;
  status: string;
  calendar_conflict_status?: string | null;
  budget_check_status?: string | null;
  current_approval_step_id?: string | null;
  current_approval_status?: string;
  current_approval_role_code?: string;
  current_approval_due_at?: string | null;
  current_approver_user_id?: string | null;
  current_approver_full_name?: string;
  approved_at?: string | null;
  rejected_at?: string | null;
  sent_to_revision_at?: string | null;
  training_started_at?: string | null;
  training_completed_at?: string | null;
  certificate_uploaded_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface PendingApprovalItem {
  request: ExternalRequest;
  current_step: {
    step_id: string;
    role_code: string;
    due_at?: string | null;
    approver_user_id: string;
  };
}

export interface LearningPlanItem {
  enrollment_id: string;
  course_id: string;
  title: string;
  status: string;
  completion_percent: string;
  deadline_at?: string | null;
}

export interface RecommendedCourseItem {
  course_id: string;
  title: string;
  short_description?: string | null;
}

export interface MyLearningPlan {
  summary: {
    total: number;
    in_progress: number;
    upcoming: number;
    completed_recently: number;
    recommended: number;
  };
  in_progress: LearningPlanItem[];
  upcoming: LearningPlanItem[];
  completed_recently: LearningPlanItem[];
  recommended: RecommendedCourseItem[];
}

export interface TeamPreviewItem {
  user_id: string;
  first_name: string;
  last_name: string;
  position_title?: string | null;
  department_id?: string | null;
}

export interface ManagerDashboard {
  stats: {
    team_size: number;
    pending_external_approvals: number;
    team_external_requests: number;
    team_course_requests: number;
  };
  team_preview: TeamPreviewItem[];
  pending_external_approvals: PendingApprovalItem[];
  team_external_requests: ExternalRequest[];
  team_course_requests: Array<{
    id: string;
    status: string;
    course_title: string;
    employee_full_name: string;
  }>;
}

export interface CourseAssignment {
  id: string;
  course_id: string;
  assignment_type: string;
  target_type: string;
  target_id: string;
  assigned_by: string;
  priority: string;
  reason?: string | null;
  start_at?: string | null;
  deadline_at?: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ApprovalWorkflowStep {
  id: string;
  workflow_id: string;
  step_order: number;
  role_code: string;
  approver_source: string;
  approver_user_id?: string | null;
  sla_hours?: number | null;
  is_required: boolean;
}

export interface ApprovalWorkflow {
  id: string;
  entity_type: string;
  name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  steps?: ApprovalWorkflowStep[];
}

export interface BudgetLimitRecord {
  id: string;
  scope_type: string;
  scope_id?: string | null;
  period_year: number;
  period_month?: number | null;
  limit_amount: string;
  currency: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface HrDashboardStats {
  users: number;
  courses: number;
  enrollments: number;
  external_requests: number;
  pending_approvals: number;
}

export interface BoardSummaryBoardItem {
  board_id: string;
  title: string;
  tasks_total: number;
  active_total: number;
  completed_total: number;
  overdue_total: number;
}

export interface BoardSummaryOverdueTask {
  task_id: string;
  board_id?: string | null;
  board_title?: string | null;
  title: string;
  deadline_at?: string | null;
  completed: boolean;
  archived: boolean;
}

export interface BoardSummary {
  source: string;
  status: string;
  summary: {
    boards_total: number;
    tasks_total: number;
    active_total: number;
    completed_total: number;
    overdue_total: number;
  };
  boards: BoardSummaryBoardItem[];
  overdue_tasks: BoardSummaryOverdueTask[];
}

export interface AppNotification {
  id: string;
  title: string;
  body: string;
  status: string;
  related_entity_type?: string | null;
  related_entity_id?: string | null;
  created_at: string;
}

export interface CourseIntake {
  id: string;
  course_id?: string | null;
  title: string;
  description?: string | null;
  opened_by: string;
  approver_id?: string | null;
  max_participants?: number | null;
  start_date?: string | null;
  end_date?: string | null;
  application_deadline?: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CourseApplication {
  id: string;
  intake_id: string;
  applicant_id: string;
  motivation?: string | null;
  status: string;
  manager_approver_id?: string | null;
  manager_comment?: string | null;
  manager_decided_at?: string | null;
  hr_approver_id?: string | null;
  hr_comment?: string | null;
  hr_decided_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CourseSuggestion {
  id: string;
  suggested_by: string;
  title: string;
  description?: string | null;
  external_url?: string | null;
  provider_name?: string | null;
  price?: string | null;
  price_currency: string;
  duration_hours?: string | null;
  approver_id?: string | null;
  status: string;
  reviewed_by?: string | null;
  review_comment?: string | null;
  reviewed_at?: string | null;
  intake_id?: string | null;
  created_at: string;
  updated_at: string;
}

export interface SuggestionOpenIntakeResponse {
  suggestion: CourseSuggestion;
  intake: CourseIntake;
}
