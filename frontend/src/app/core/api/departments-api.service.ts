import { Injectable } from '@angular/core';
import { Observable, map } from 'rxjs';

import type { Department } from '@entities/department';
import { AuthApiService } from '@core/auth/auth-api.service';

@Injectable({ providedIn: 'root' })
export class DepartmentsApiService {
  constructor(private readonly authApi: AuthApiService) {}

  getCurrentUserDepartment(): Observable<Department | null> {
    return this.authApi.me().pipe(
      map((payload) => {
        const departmentId = payload.user.employee_profile?.department_id;
        if (!departmentId) {
          return null;
        }

        // Backend currently has no public /departments endpoint.
        return {
          id: departmentId,
          status: 'active',
        };
      }),
    );
  }
}