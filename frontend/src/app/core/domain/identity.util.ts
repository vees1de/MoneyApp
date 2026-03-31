import type { IdentityUserView } from '@core/auth/auth.types';

export function identityUserDisplayName(user: IdentityUserView | null | undefined): string {
  if (!user) {
    return '—';
  }

  const profile = user.employee_profile;
  if (!profile) {
    return user.email;
  }

  return [profile.last_name, profile.first_name, profile.middle_name]
    .filter((item) => !!item)
    .join(' ')
    .trim();
}
