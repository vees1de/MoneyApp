export const PERMISSIONS = {
  usersRead: 'users.read',
  usersWrite: 'users.write',
  rolesManage: 'roles.manage',
  coursesRead: 'courses.read',
  coursesWrite: 'courses.write',
  coursesAssign: 'courses.assign',
  enrollmentsRead: 'enrollments.read',
  enrollmentsManage: 'enrollments.manage',
  externalRequestsCreate: 'external_requests.create',
  externalRequestsReadOwn: 'external_requests.read_own',
  externalRequestsReadAll: 'external_requests.read_all',
  externalRequestsApproveManager: 'external_requests.approve_manager',
  externalRequestsApproveHr: 'external_requests.approve_hr',
  certificatesVerify: 'certificates.verify',
  programsManage: 'programs.manage',
  analyticsReadHr: 'analytics.read_hr',
  analyticsReadManager: 'analytics.read_manager',
  notificationsManage: 'notifications.manage',
  settingsManage: 'settings.manage',
  auditRead: 'audit.read',
} as const;

export type PermissionCode = (typeof PERMISSIONS)[keyof typeof PERMISSIONS] | string;
