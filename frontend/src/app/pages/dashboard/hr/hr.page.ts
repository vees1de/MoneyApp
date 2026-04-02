import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { CourseRequestsWidgetComponent } from '@app/widgets/course-requests-widget/course-requests-widget.component';
import { CurrentLearningWidgetComponent } from '@app/widgets/current-learning-widget/current-learning-widget.component';
import { DepartmentBudgetWidgetComponent } from '@app/widgets/department-budget-widget/department-budget-widget.component';
import { HrSuggestionsWidgetComponent } from '@app/widgets/hr-suggestions-widget/hr-suggestions-widget.component';
import { IntakeHrApprovalsWidgetComponent } from '@app/widgets/intake-hr-approvals-widget/intake-hr-approvals-widget.component';
import { QuickActionsWidgetComponent } from '@app/widgets/quick-actions-widget/quick-actions-widget.component';
import { RisksWidgetComponent } from '@app/widgets/risks-widget/risks-widget.component';
import { SpendingHistoryWidgetComponent } from '@app/widgets/spending-history-widget/spending-history-widget.component';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';

@Component({
  selector: 'app-page-dashboard-hr',
  standalone: true,
  imports: [
    CommonModule,
    QuickActionsWidgetComponent,
    UpcomingEventsWidgetComponent,
    CourseRequestsWidgetComponent,
    DepartmentBudgetWidgetComponent,
    HrSuggestionsWidgetComponent,
    IntakeHrApprovalsWidgetComponent,
    CurrentLearningWidgetComponent,
    RisksWidgetComponent,
    SpendingHistoryWidgetComponent,
  ],
  templateUrl: './hr.page.html',
  styleUrl: './hr.page.scss',
})
export class DashboardHrPageComponent {}
