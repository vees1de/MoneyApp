export enum UserRole {
  Administrator = 'ADMINISTRATOR',
  HrLnd = 'HR_LND',
  Manager = 'MANAGER',
  Employee = 'EMPLOYEE',
  Trainer = 'TRAINER',
}

export enum AppFeature {
  Dashboard = 'DASHBOARD',
  ExternalRequests = 'EXTERNAL_REQUESTS',
  Approvals = 'APPROVALS',
  Calendar = 'CALENDAR',
  Certificates = 'CERTIFICATES',
  Reports = 'REPORTS',
  University = 'UNIVERSITY',
  Notifications = 'NOTIFICATIONS',
  Profile = 'PROFILE',
  Catalog = 'CATALOG',
  AdminUsers = 'ADMIN_USERS',
  AdminRoles = 'ADMIN_ROLES',
}

export interface AuthUser {
  id: string;
  fullName: string;
  roles: UserRole[];
  features: AppFeature[];
}

export interface RecentAction {
  id: string;
  label: string;
  route: string;
  at: string;
}
