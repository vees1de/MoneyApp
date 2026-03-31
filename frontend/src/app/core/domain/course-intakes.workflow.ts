import type { RoleCode } from '@core/auth/auth.types';

export type CourseIntakeStatus = 'open' | 'closed' | 'canceled' | 'completed' | string;
export type CourseApplicationStatus =
  | 'pending'
  | 'pending_manager'
  | 'approved_by_manager'
  | 'approved'
  | 'rejected_by_manager'
  | 'rejected_by_hr'
  | 'withdrawn'
  | 'enrolled'
  | string;
export type CourseApplicationPaymentStatus = 'paid' | 'unpaid' | string;
export type CourseSuggestionStatus = 'pending' | 'approved' | 'rejected' | 'intake_opened' | string;

export function courseIntakeStatusLabel(status: CourseIntakeStatus): string {
  const labels: Record<string, string> = {
    open: 'Открыт',
    closed: 'Набор закрыт',
    canceled: 'Отменён',
    completed: 'Завершён',
  };

  return labels[status] ?? status;
}

export function courseApplicationStatusLabel(status: CourseApplicationStatus): string {
  const labels: Record<string, string> = {
    pending: 'Ожидает решения HR',
    pending_manager: 'На согласовании у руководителя',
    approved_by_manager: 'Согласовано руководителем',
    approved: 'Одобрено HR',
    rejected_by_manager: 'Отклонено руководителем',
    rejected_by_hr: 'Отклонено HR',
    withdrawn: 'Отозвано сотрудником',
    enrolled: 'Сотрудник взят',
  };

  return labels[status] ?? status;
}

export function courseApplicationPaymentStatusLabel(
  status: CourseApplicationPaymentStatus,
): string {
  const labels: Record<string, string> = {
    paid: 'Оплачен',
    unpaid: 'Не оплачен',
  };

  return labels[status] ?? status;
}

export function courseSuggestionStatusLabel(status: CourseSuggestionStatus): string {
  const labels: Record<string, string> = {
    pending: 'На ревью HR',
    approved: 'Одобрено',
    rejected: 'Отклонено',
    intake_opened: 'Набор открыт',
  };

  return labels[status] ?? status;
}

export function isIntakeManageRole(role: RoleCode): boolean {
  return role === 'hr' || role === 'admin';
}

export function canApplyToIntake(
  status: CourseIntakeStatus,
  hasExistingApplication: boolean,
): boolean {
  return status === 'open' && !hasExistingApplication;
}

export function canWithdrawApplication(status: CourseApplicationStatus): boolean {
  return status !== 'withdrawn' && status !== 'enrolled';
}

export function canManagerReviewApplication(status: CourseApplicationStatus): boolean {
  return status === 'pending_manager';
}

export function canHrReviewApplication(status: CourseApplicationStatus): boolean {
  return status === 'pending' || status === 'approved_by_manager';
}

export function canEnrollApplication(status: CourseApplicationStatus): boolean {
  return status === 'approved';
}

export function canReviewSuggestion(role: RoleCode): boolean {
  return isIntakeManageRole(role);
}

export function canOpenIntakeFromSuggestion(status: CourseSuggestionStatus): boolean {
  return status === 'pending' || status === 'approved';
}
