export interface CalendarUpcomingEvent {
  id: string;
  title: string;
  start_at: string;
  end_at: string;
  status: string;
  meeting_url?: string | null;
  location?: string | null;
}

export interface ExternalRequest {
  id: string;
  request_no: string;
  employee_user_id: string;
  department_id?: string | null;
  title: string;
  provider_id?: string | null;
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
  status: string;
  calendar_conflict_status?: string | null;
  budget_check_status?: string | null;
  approved_at?: string | null;
  rejected_at?: string | null;
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

export interface BoardSummary {
  summary: {
    boards_total: number;
    tasks_total: number;
    active_total: number;
    completed_total: number;
    overdue_total: number;
  };
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
