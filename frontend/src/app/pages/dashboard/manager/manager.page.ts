import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { CourseRequestsWidgetComponent } from '@app/widgets/course-requests-widget/course-requests-widget.component';
import { CurrentLearningWidgetComponent } from '@app/widgets/current-learning-widget/current-learning-widget.component';
import { EmployeeQuotasWidgetComponent } from '@app/widgets/employee-quotas-widget/employee-quotas-widget.component';
import { QuickActionsWidgetComponent } from '@app/widgets/quick-actions-widget/quick-actions-widget.component';
import { TeamOverviewWidgetComponent } from '@app/widgets/team-overview-widget/team-overview-widget.component';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';
import { YougilePlaygroundWidgetComponent } from '@app/widgets/yougile-playground-widget/yougile-playground-widget.component';

@Component({
  selector: 'app-page-dashboard-manager',
  standalone: true,
  imports: [
    CommonModule,
    TeamOverviewWidgetComponent,
    UpcomingEventsWidgetComponent,
    CourseRequestsWidgetComponent,
    CurrentLearningWidgetComponent,
    EmployeeQuotasWidgetComponent,
    QuickActionsWidgetComponent,
    YougilePlaygroundWidgetComponent,
  ],
  templateUrl: './manager.page.html',
  styleUrl: './manager.page.scss',
})
export class DashboardManagerPageComponent {}
