import { Routes } from '@angular/router';

import { authGuard } from '@core/auth/auth.guard';
import { featureGuard } from '@core/auth/feature.guard';
import { roleGuard } from '@core/auth/role.guard';
import { AppFeature, UserRole } from '@core/auth/auth.types';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'home' },

  {
    path: 'login',
    loadComponent: () =>
      import('@pages/public/login/login.page').then((m) => m.PublicLoginPageComponent),
  },
  {
    path: 'forbidden',
    loadComponent: () =>
      import('@pages/public/forbidden/forbidden.page').then(
        (m) => m.PublicForbiddenPageComponent,
      ),
  },
  {
    path: 'error',
    loadComponent: () =>
      import('@pages/public/error/error.page').then((m) => m.PublicErrorPageComponent),
  },
  {
    path: 'not-found',
    loadComponent: () =>
      import('@pages/public/not-found/not-found.page').then(
        (m) => m.PublicNotFoundPageComponent,
      ),
  },

  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@core/layout/app-shell/app-shell.component').then((m) => m.AppShellComponent),
    children: [
      {
        path: 'home',
        canActivate: [featureGuard(AppFeature.Dashboard)],
        loadComponent: () => import('@pages/home/home.page').then((m) => m.HomePageComponent),
      },
      {
        path: 'dashboard',
        canActivate: [featureGuard(AppFeature.Dashboard)],
        loadComponent: () => import('@pages/home/home.page').then((m) => m.HomePageComponent),
      },

      {
        path: 'dashboard/employee',
        canActivate: [roleGuard([UserRole.Employee])],
        loadComponent: () =>
          import('@pages/dashboard/employee/employee.page').then(
            (m) => m.DashboardEmployeePageComponent,
          ),
      },
      {
        path: 'dashboard/manager',
        canActivate: [roleGuard([UserRole.Manager])],
        loadComponent: () =>
          import('@pages/dashboard/manager/manager.page').then(
            (m) => m.DashboardManagerPageComponent,
          ),
      },
      {
        path: 'dashboard/hr',
        canActivate: [roleGuard([UserRole.HrLnd])],
        loadComponent: () =>
          import('@pages/dashboard/hr/hr.page').then((m) => m.DashboardHrPageComponent),
      },
      {
        path: 'dashboard/trainer',
        canActivate: [roleGuard([UserRole.Trainer])],
        loadComponent: () =>
          import('@pages/dashboard/trainer/trainer.page').then(
            (m) => m.DashboardTrainerPageComponent,
          ),
      },
      {
        path: 'dashboard/admin',
        canActivate: [roleGuard([UserRole.Administrator])],
        loadComponent: () =>
          import('@pages/dashboard/admin/admin.page').then((m) => m.DashboardAdminPageComponent),
      },

      {
        path: 'catalog',
        canActivate: [featureGuard(AppFeature.Catalog)],
        loadComponent: () =>
          import('@pages/catalog/list/list.page').then((m) => m.CatalogListPageComponent),
      },
      {
        path: 'catalog/:courseId',
        canActivate: [featureGuard(AppFeature.Catalog)],
        loadComponent: () =>
          import('@pages/catalog/detail/detail.page').then((m) => m.CatalogDetailPageComponent),
      },
      {
        path: 'my-learning',
        canActivate: [featureGuard(AppFeature.Dashboard)],
        loadComponent: () =>
          import('@pages/my-learning/list/list.page').then((m) => m.MyLearningListPageComponent),
      },
      {
        path: 'learning/certificates',
        canActivate: [featureGuard(AppFeature.Certificates)],
        loadComponent: () =>
          import('@pages/learning/certificates/certificates.page').then(
            (m) => m.LearningCertificatesPageComponent,
          ),
      },
      {
        path: 'learning/:enrollmentId',
        canActivate: [featureGuard(AppFeature.Dashboard)],
        loadComponent: () =>
          import('@pages/learning/enrollment-detail/enrollment-detail.page').then(
            (m) => m.LearningEnrollmentDetailPageComponent,
          ),
      },

      {
        path: 'external-requests',
        canActivate: [featureGuard(AppFeature.ExternalRequests)],
        loadComponent: () =>
          import('@pages/external-requests/list/list.page').then(
            (m) => m.ExternalRequestsListPageComponent,
          ),
      },
      {
        path: 'external-requests/new',
        canActivate: [featureGuard(AppFeature.ExternalRequests)],
        loadComponent: () =>
          import('@pages/external-requests/new/new.page').then(
            (m) => m.ExternalRequestsNewPageComponent,
          ),
      },
      {
        path: 'external-requests/:requestId',
        canActivate: [featureGuard(AppFeature.ExternalRequests)],
        loadComponent: () =>
          import('@pages/external-requests/detail/detail.page').then(
            (m) => m.ExternalRequestsDetailPageComponent,
          ),
      },

      {
        path: 'approvals/inbox',
        canActivate: [featureGuard(AppFeature.Approvals)],
        loadComponent: () =>
          import('@pages/approvals/inbox/inbox.page').then((m) => m.ApprovalsInboxPageComponent),
      },
      {
        path: 'approvals/:requestId',
        canActivate: [featureGuard(AppFeature.Approvals)],
        loadComponent: () =>
          import('@pages/approvals/decision/decision.page').then(
            (m) => m.ApprovalsDecisionPageComponent,
          ),
      },

      {
        path: 'calendar',
        canActivate: [featureGuard(AppFeature.Calendar)],
        loadComponent: () =>
          import('@pages/calendar/overview/overview.page').then(
            (m) => m.CalendarOverviewPageComponent,
          ),
      },
      {
        path: 'calendar/conflicts',
        canActivate: [featureGuard(AppFeature.Calendar)],
        loadComponent: () =>
          import('@pages/calendar/conflicts/conflicts.page').then(
            (m) => m.CalendarConflictsPageComponent,
          ),
      },

      {
        path: 'notifications',
        canActivate: [featureGuard(AppFeature.Notifications)],
        loadComponent: () =>
          import('@pages/notifications/center/center.page').then(
            (m) => m.NotificationsCenterPageComponent,
          ),
      },
      {
        path: 'profile',
        canActivate: [featureGuard(AppFeature.Profile)],
        loadComponent: () =>
          import('@pages/profile/overview/overview.page').then(
            (m) => m.ProfileOverviewPageComponent,
          ),
      },

      {
        path: 'reports/overview',
        canActivate: [featureGuard(AppFeature.Reports)],
        loadComponent: () =>
          import('@pages/reports/overview/overview.page').then(
            (m) => m.ReportsOverviewPageComponent,
          ),
      },
      {
        path: 'reports/progress',
        canActivate: [featureGuard(AppFeature.Reports)],
        loadComponent: () =>
          import('@pages/reports/progress/progress.page').then(
            (m) => m.ReportsProgressPageComponent,
          ),
      },
      {
        path: 'reports/budget',
        canActivate: [featureGuard(AppFeature.Reports)],
        loadComponent: () =>
          import('@pages/reports/budget/budget.page').then((m) => m.ReportsBudgetPageComponent),
      },
      {
        path: 'reports/traceability',
        canActivate: [featureGuard(AppFeature.Reports)],
        loadComponent: () =>
          import('@pages/reports/traceability/traceability.page').then(
            (m) => m.ReportsTraceabilityPageComponent,
          ),
      },
      {
        path: 'reports/export',
        canActivate: [featureGuard(AppFeature.Reports)],
        loadComponent: () =>
          import('@pages/reports/export/export.page').then((m) => m.ReportsExportPageComponent),
      },

      {
        path: 'university/directions',
        canActivate: [featureGuard(AppFeature.University)],
        loadComponent: () =>
          import('@pages/university/directions/directions.page').then(
            (m) => m.UniversityDirectionsPageComponent,
          ),
      },
      {
        path: 'university/programs/:programId',
        canActivate: [featureGuard(AppFeature.University)],
        loadComponent: () =>
          import('@pages/university/programs/programs.page').then(
            (m) => m.UniversityProgramsPageComponent,
          ),
      },
      {
        path: 'university/groups',
        canActivate: [featureGuard(AppFeature.University)],
        loadComponent: () =>
          import('@pages/university/groups/groups.page').then(
            (m) => m.UniversityGroupsPageComponent,
          ),
      },
      {
        path: 'university/feedback/:courseId',
        canActivate: [featureGuard(AppFeature.University)],
        loadComponent: () =>
          import('@pages/university/feedback/feedback.page').then(
            (m) => m.UniversityFeedbackPageComponent,
          ),
      },

      {
        path: 'admin/users',
        canActivate: [
          roleGuard([UserRole.Administrator]),
          featureGuard(AppFeature.AdminUsers),
        ],
        loadComponent: () =>
          import('@pages/admin/users/users.page').then((m) => m.AdminUsersPageComponent),
      },
      {
        path: 'admin/roles',
        canActivate: [
          roleGuard([UserRole.Administrator]),
          featureGuard(AppFeature.AdminRoles),
        ],
        loadComponent: () =>
          import('@pages/admin/roles/roles.page').then((m) => m.AdminRolesPageComponent),
      },

      { path: '**', redirectTo: 'dashboard' },
    ],
  },

  { path: '**', redirectTo: 'dashboard' },
];
