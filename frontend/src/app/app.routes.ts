import { Routes } from '@angular/router';

import { authGuard } from '@core/auth/auth.guard';
import { dashboardRedirectGuard } from '@core/auth/dashboard-redirect.guard';
import { permissionGuard } from '@core/auth/feature.guard';
import { guestGuard } from '@core/auth/guest.guard';
import { PERMISSIONS } from '@core/auth/permissions';
import { roleGuard } from '@core/auth/role.guard';
import { COURSE_INTAKE_MANAGER_APPROVAL_ENABLED } from '@core/config/feature-flags';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'dashboard' },

  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () =>
      import('@pages/public/login/login.page').then((m) => m.PublicLoginPageComponent),
  },
  {
    path: 'forbidden',
    loadComponent: () =>
      import('@pages/public/forbidden/forbidden.page').then((m) => m.PublicForbiddenPageComponent),
  },
  {
    path: 'error',
    loadComponent: () =>
      import('@pages/public/error/error.page').then((m) => m.PublicErrorPageComponent),
  },
  {
    path: 'not-found',
    loadComponent: () =>
      import('@pages/public/not-found/not-found.page').then((m) => m.PublicNotFoundPageComponent),
  },

  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@core/layout/app-shell/app-shell.component').then((m) => m.AppShellComponent),
    children: [
      {
        path: 'dashboard',
        canActivate: [dashboardRedirectGuard],
        pathMatch: 'full',
        loadComponent: () =>
          import('@pages/public/error/error.page').then((m) => m.PublicErrorPageComponent),
      },
      {
        path: 'dashboard/test-role',
        loadComponent: () =>
          import('@pages/dashboard/test-role/test-role.page').then(
            (m) => m.DashboardTestRolePageComponent,
          ),
      },
      {
        path: 'playground',
        loadComponent: () =>
          import('@pages/dashboard/test-role/test-role.page').then(
            (m) => m.DashboardTestRolePageComponent,
          ),
      },

      {
        path: 'dashboard/employee',
        canActivate: [roleGuard(['employee'])],
        loadComponent: () =>
          import('@pages/dashboard/employee/employee.page').then(
            (m) => m.DashboardEmployeePageComponent,
          ),
      },
      {
        path: 'dashboard/manager',
        canActivate: [roleGuard(['manager'])],
        loadComponent: () =>
          import('@pages/dashboard/manager/manager.page').then(
            (m) => m.DashboardManagerPageComponent,
          ),
      },
      {
        path: 'dashboard/hr',
        canActivate: [roleGuard(['hr'])],
        loadComponent: () =>
          import('@pages/dashboard/hr/hr.page').then((m) => m.DashboardHrPageComponent),
      },
      {
        path: 'dashboard/trainer',
        canActivate: [roleGuard(['trainer'])],
        loadComponent: () =>
          import('@pages/dashboard/trainer/trainer.page').then(
            (m) => m.DashboardTrainerPageComponent,
          ),
      },
      {
        path: 'dashboard/admin',
        canActivate: [roleGuard(['admin'])],
        loadComponent: () =>
          import('@pages/dashboard/admin/admin.page').then((m) => m.DashboardAdminPageComponent),
      },
      {
        path: 'testadmin',
        canActivate: [roleGuard(['admin', 'hr'])],
        loadComponent: () =>
          import('@pages/admin/testadmin/testadmin.page').then(
            (m) => m.AdminTestAdminPageComponent,
          ),
      },

      {
        path: 'catalog',
        canActivate: [permissionGuard([PERMISSIONS.coursesRead])],
        loadComponent: () =>
          import('@pages/catalog/list/list.page').then((m) => m.CatalogListPageComponent),
      },
      {
        path: 'catalog/:courseId',
        canActivate: [permissionGuard([PERMISSIONS.coursesRead])],
        loadComponent: () =>
          import('@pages/catalog/detail/detail.page').then((m) => m.CatalogDetailPageComponent),
      },
      {
        path: 'intakes',
        loadComponent: () =>
          import('@pages/intakes/list/list.page').then((m) => m.IntakesListPageComponent),
      },
      {
        path: 'intakes/new',
        canActivate: [roleGuard(['hr', 'admin'])],
        loadComponent: () =>
          import('@pages/intakes/new/new.page').then((m) => m.IntakesNewPageComponent),
      },
      {
        path: 'intakes/:intakeId',
        loadComponent: () =>
          import('@pages/intakes/detail/detail.page').then((m) => m.IntakeDetailPageComponent),
      },
      {
        path: 'applications/my',
        loadComponent: () =>
          import('@pages/applications/my/my.page').then((m) => m.ApplicationsMyPageComponent),
      },
      ...(COURSE_INTAKE_MANAGER_APPROVAL_ENABLED
        ? [
            {
              path: 'applications/pending-manager',
              canActivate: [roleGuard(['manager', 'admin'])],
              loadComponent: () =>
                import('@pages/applications/pending-manager/pending-manager.page').then(
                  (m) => m.ApplicationsPendingManagerPageComponent,
                ),
            },
          ]
        : []),
      {
        path: 'suggestions',
        loadComponent: () =>
          import('@pages/suggestions/list/list.page').then((m) => m.SuggestionsListPageComponent),
      },
      {
        path: 'suggestions/new',
        loadComponent: () =>
          import('@pages/suggestions/new/new.page').then((m) => m.SuggestionsNewPageComponent),
      },
      {
        path: 'suggestions/:suggestionId',
        loadComponent: () =>
          import('@pages/suggestions/detail/detail.page').then(
            (m) => m.SuggestionDetailPageComponent,
          ),
      },
      {
        path: 'my-learning',
        canActivate: [
          permissionGuard([PERMISSIONS.enrollmentsRead, PERMISSIONS.enrollmentsManage]),
        ],
        loadComponent: () =>
          import('@pages/my-learning/list/list.page').then((m) => m.MyLearningListPageComponent),
      },
      {
        path: 'learning/certificates',
        canActivate: [
          permissionGuard([PERMISSIONS.certificatesVerify, PERMISSIONS.enrollmentsRead]),
        ],
        loadComponent: () =>
          import('@pages/learning/certificates/certificates.page').then(
            (m) => m.LearningCertificatesPageComponent,
          ),
      },
      {
        path: 'learning/:enrollmentId',
        canActivate: [
          permissionGuard([PERMISSIONS.enrollmentsRead, PERMISSIONS.enrollmentsManage]),
        ],
        loadComponent: () =>
          import('@pages/learning/enrollment-detail/enrollment-detail.page').then(
            (m) => m.LearningEnrollmentDetailPageComponent,
          ),
      },

      {
        path: 'external-requests',
        canActivate: [
          permissionGuard([
            PERMISSIONS.externalRequestsCreate,
            PERMISSIONS.externalRequestsReadOwn,
            PERMISSIONS.externalRequestsReadAll,
          ]),
        ],
        loadComponent: () =>
          import('@pages/external-requests/list/list.page').then(
            (m) => m.ExternalRequestsListPageComponent,
          ),
      },
      {
        path: 'external-requests/new',
        canActivate: [permissionGuard([PERMISSIONS.externalRequestsCreate])],
        loadComponent: () =>
          import('@pages/external-requests/new/new.page').then(
            (m) => m.ExternalRequestsNewPageComponent,
          ),
      },
      {
        path: 'external-requests/:requestId',
        canActivate: [
          permissionGuard([
            PERMISSIONS.externalRequestsReadOwn,
            PERMISSIONS.externalRequestsReadAll,
            PERMISSIONS.externalRequestsApproveManager,
            PERMISSIONS.externalRequestsApproveHr,
          ]),
        ],
        loadComponent: () =>
          import('@pages/external-requests/detail/detail.page').then(
            (m) => m.ExternalRequestsDetailPageComponent,
          ),
      },

      {
        path: 'approvals/inbox',
        canActivate: [
          permissionGuard([
            PERMISSIONS.externalRequestsApproveManager,
            PERMISSIONS.externalRequestsApproveHr,
          ]),
        ],
        loadComponent: () =>
          import('@pages/approvals/inbox/inbox.page').then((m) => m.ApprovalsInboxPageComponent),
      },
      {
        path: 'approvals/:requestId',
        canActivate: [
          permissionGuard([
            PERMISSIONS.externalRequestsApproveManager,
            PERMISSIONS.externalRequestsApproveHr,
          ]),
        ],
        loadComponent: () =>
          import('@pages/approvals/decision/decision.page').then(
            (m) => m.ApprovalsDecisionPageComponent,
          ),
      },

      {
        path: 'calendar',
        canActivate: [permissionGuard([PERMISSIONS.enrollmentsRead])],
        loadComponent: () =>
          import('@pages/calendar/overview/overview.page').then(
            (m) => m.CalendarOverviewPageComponent,
          ),
      },
      {
        path: 'calendar/conflicts',
        canActivate: [permissionGuard([PERMISSIONS.enrollmentsRead])],
        loadComponent: () =>
          import('@pages/calendar/conflicts/conflicts.page').then(
            (m) => m.CalendarConflictsPageComponent,
          ),
      },

      {
        path: 'notifications',
        canActivate: [permissionGuard([PERMISSIONS.notificationsManage])],
        loadComponent: () =>
          import('@pages/notifications/center/center.page').then(
            (m) => m.NotificationsCenterPageComponent,
          ),
      },
      {
        path: 'profile',
        pathMatch: 'full',
        redirectTo: 'profile/overview',
      },
      {
        path: 'profile/overview',
        canActivate: [permissionGuard([PERMISSIONS.coursesRead])],
        loadComponent: () =>
          import('@pages/profile/overview/overview.page').then(
            (m) => m.ProfileOverviewPageComponent,
          ),
      },

      {
        path: 'reports/overview',
        canActivate: [
          permissionGuard([PERMISSIONS.analyticsReadHr, PERMISSIONS.analyticsReadManager]),
        ],
        loadComponent: () =>
          import('@pages/reports/overview/overview.page').then(
            (m) => m.ReportsOverviewPageComponent,
          ),
      },
      {
        path: 'reports/progress',
        canActivate: [
          permissionGuard([PERMISSIONS.analyticsReadHr, PERMISSIONS.analyticsReadManager]),
        ],
        loadComponent: () =>
          import('@pages/reports/progress/progress.page').then(
            (m) => m.ReportsProgressPageComponent,
          ),
      },
      {
        path: 'reports/budget',
        canActivate: [permissionGuard([PERMISSIONS.analyticsReadHr])],
        loadComponent: () =>
          import('@pages/reports/budget/budget.page').then((m) => m.ReportsBudgetPageComponent),
      },
      {
        path: 'reports/traceability',
        canActivate: [
          permissionGuard([PERMISSIONS.analyticsReadHr, PERMISSIONS.analyticsReadManager]),
        ],
        loadComponent: () =>
          import('@pages/reports/traceability/traceability.page').then(
            (m) => m.ReportsTraceabilityPageComponent,
          ),
      },
      {
        path: 'reports/export',
        canActivate: [
          permissionGuard([PERMISSIONS.analyticsReadHr, PERMISSIONS.analyticsReadManager]),
        ],
        loadComponent: () =>
          import('@pages/reports/export/export.page').then((m) => m.ReportsExportPageComponent),
      },

      {
        path: 'university/directions',
        canActivate: [permissionGuard([PERMISSIONS.programsManage])],
        loadComponent: () =>
          import('@pages/university/directions/directions.page').then(
            (m) => m.UniversityDirectionsPageComponent,
          ),
      },
      {
        path: 'university/programs/:programId',
        canActivate: [permissionGuard([PERMISSIONS.programsManage])],
        loadComponent: () =>
          import('@pages/university/programs/programs.page').then(
            (m) => m.UniversityProgramsPageComponent,
          ),
      },
      {
        path: 'university/groups',
        canActivate: [permissionGuard([PERMISSIONS.programsManage])],
        loadComponent: () =>
          import('@pages/university/groups/groups.page').then(
            (m) => m.UniversityGroupsPageComponent,
          ),
      },
      {
        path: 'university/feedback/:courseId',
        canActivate: [permissionGuard([PERMISSIONS.programsManage])],
        loadComponent: () =>
          import('@pages/university/feedback/feedback.page').then(
            (m) => m.UniversityFeedbackPageComponent,
          ),
      },

      {
        path: 'admin/users',
        canActivate: [permissionGuard([PERMISSIONS.usersRead, PERMISSIONS.usersWrite])],
        loadComponent: () =>
          import('@pages/admin/users/users.page').then((m) => m.AdminUsersPageComponent),
      },
      {
        path: 'admin/roles',
        canActivate: [permissionGuard([PERMISSIONS.rolesManage])],
        loadComponent: () =>
          import('@pages/admin/roles/roles.page').then((m) => m.AdminRolesPageComponent),
      },

      { path: '**', redirectTo: 'dashboard' },
    ],
  },

  { path: '**', redirectTo: 'dashboard' },
];
