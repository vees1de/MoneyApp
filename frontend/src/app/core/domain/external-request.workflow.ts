import type { RoleCode } from '@core/auth/auth.types';

export type ExternalRequestStatus =
  | 'draft'
  | 'manager_approval'
  | 'hr_approval'
  | 'approved'
  | 'rejected'
  | 'revision_requested'
  | 'completed'
  | string;

export type ExternalRequestAction = 'submit' | 'approve' | 'reject' | 'request_revision';

export function externalRequestStatusLabel(status: ExternalRequestStatus): string {
  const map: Record<string, string> = {
    draft: 'Черновик',
    manager_approval: 'На согласовании у руководителя',
    hr_approval: 'На согласовании у HR',
    approved: 'Согласовано',
    rejected: 'Отклонено',
    revision_requested: 'Требует доработки',
    completed: 'Завершено',
  };

  return map[status] ?? status;
}

export function canDoExternalRequestAction(
  role: RoleCode,
  status: ExternalRequestStatus,
  action: ExternalRequestAction,
): boolean {
  if (action === 'submit') {
    return role === 'employee' && (status === 'draft' || status === 'revision_requested');
  }

  if (action === 'approve') {
    return (role === 'manager' && status === 'manager_approval') || (role === 'hr' && status === 'hr_approval');
  }

  if (action === 'reject' || action === 'request_revision') {
    return (role === 'manager' && status === 'manager_approval') || (role === 'hr' && status === 'hr_approval');
  }

  return false;
}

export function externalRequestsScopeByRole(role: RoleCode): 'my' | 'team' | 'all' {
  if (role === 'employee' || role === 'trainer') return 'my';
  if (role === 'manager') return 'team';
  return 'all';
}
