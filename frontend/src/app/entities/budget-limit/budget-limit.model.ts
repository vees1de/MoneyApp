export type BudgetUsageStatus = 'healthy' | 'attention' | 'risk' | 'over_limit';

export type TrainingExpenseStatus =
  | 'reserved'
  | 'approved'
  | 'completed'
  | 'rejected'
  | 'revision_returned'
  | 'cancelled';

export type TrainingBudgetCheckStatus =
  | 'within_budget'
  | 'employee_quota_exceeded'
  | 'department_budget_exceeded'
  | 'department_and_employee_exceeded';

export interface EmployeeTrainingQuota {
  employeeId: string;
  employeeName: string;
  employeeEmail?: string | null;
  departmentId: string;
  departmentName: string;
  quotaAmount: number;
  spentAmount: number;
  reservedAmount: number;
  usedAmount: number;
  remainingAmount: number;
  usagePercent: number;
  status: BudgetUsageStatus;
}

export interface TrainingExpense {
  id: string;
  requestId: string;
  departmentId: string;
  departmentName: string;
  employeeId: string;
  employeeName: string;
  employeeEmail?: string | null;
  courseTitle: string;
  providerName?: string | null;
  amount: number;
  currency: string;
  status: TrainingExpenseStatus;
  budgetCheckStatus: TrainingBudgetCheckStatus;
  requiresAdditionalApproval: boolean;
  approvalComment?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface DepartmentTrainingBudget {
  departmentId: string;
  departmentName: string;
  totalBudget: number;
  spentAmount: number;
  reservedAmount: number;
  usedAmount: number;
  remainingAmount: number;
  usagePercent: number;
  currency: string;
  status: BudgetUsageStatus;
  employees: EmployeeTrainingQuota[];
  expenses: TrainingExpense[];
}

export interface TrainingBudgetValidation {
  departmentId: string;
  employeeId: string;
  requestAmount: number;
  currency: string;
  employeeRemainingAmount: number;
  departmentRemainingAmount: number;
  employeeEnough: boolean;
  departmentEnough: boolean;
  requiresAdditionalApproval: boolean;
  status: TrainingBudgetCheckStatus;
  warnings: string[];
}

export interface TrackedTrainingRequest {
  requestId: string;
  title: string;
  providerName?: string | null;
  amount: number;
  currency: string;
  departmentId: string;
  departmentName: string;
  employeeId: string;
  employeeName: string;
  employeeEmail?: string | null;
  requiresAdditionalApproval: boolean;
  additionalApprovalComment?: string | null;
  lastKnownStatus: string;
  updatedAt: string;
}

export interface TrainingBudgetRequestView {
  trackedRequest: TrackedTrainingRequest;
  department: DepartmentTrainingBudget | null;
  employee: EmployeeTrainingQuota | null;
  validation: TrainingBudgetValidation;
  latestExpense: TrainingExpense | null;
  history: TrainingExpense[];
}

export interface TrainingBudgetPreview {
  department: DepartmentTrainingBudget | null;
  employee: EmployeeTrainingQuota | null;
  validation: TrainingBudgetValidation;
}

export interface DepartmentBudgetReportRow {
  departmentId: string;
  departmentName: string;
  totalBudget: number;
  spentAmount: number;
  reservedAmount: number;
  usedAmount: number;
  remainingAmount: number;
  usagePercent: number;
  status: BudgetUsageStatus;
  headcount: number;
  activeRequests: number;
  currency: string;
}

export interface EmployeeBudgetReportRow {
  employeeId: string;
  employeeName: string;
  departmentId: string;
  departmentName: string;
  quotaAmount: number;
  spentAmount: number;
  reservedAmount: number;
  usedAmount: number;
  remainingAmount: number;
  usagePercent: number;
  status: BudgetUsageStatus;
  activeRequests: number;
}

export interface CourseSpendReportRow {
  courseTitle: string;
  totalAmount: number;
  requestCount: number;
  uniqueEmployees: number;
  currency: string;
}

export interface BudgetLimit {
  departments: DepartmentTrainingBudget[];
  departmentReport: DepartmentBudgetReportRow[];
  employeeReport: EmployeeBudgetReportRow[];
  topCourses: CourseSpendReportRow[];
  history: TrainingExpense[];
}
