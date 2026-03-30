import type { PermissionCode } from './permissions';

export type RoleCode = 'admin' | 'hr' | 'manager' | 'employee' | 'trainer' | string;

export interface EmployeeProfile {
  id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  middle_name?: string | null;
  position_title?: string | null;
  department_id?: string | null;
  employment_status: string;
  timezone?: string | null;
  outlook_email?: string | null;
  created_at: string;
  updated_at: string;
}

export interface IdentityUserView {
  id: string;
  email: string;
  status: string;
  is_email_verified: boolean;
  last_login_at?: string | null;
  created_at: string;
  updated_at: string;
  roles: RoleCode[];
  permissions: PermissionCode[];
  employee_profile?: EmployeeProfile | null;
}

export interface MeResponse {
  user: IdentityUserView;
}

export interface RecentAction {
  id: string;
  label: string;
  route: string;
  at: string;
}
