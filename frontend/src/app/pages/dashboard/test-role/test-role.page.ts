import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { TeamOverviewWidgetComponent } from '@app/widgets/team-overview-widget/team-overview-widget.component';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';
import { JiraSummaryWidgetComponent } from '@app/widgets/jira-summary-widget/jira-summary-widget.component';
import { CourseRequestsWidgetComponent } from '@app/widgets/course-requests-widget/course-requests-widget.component';
import { CurrentLearningWidgetComponent } from '@app/widgets/current-learning-widget/current-learning-widget.component';
import { RecommendationsWidgetComponent } from '@app/widgets/recommendations-widget/recommendations-widget.component';
import { QuickActionsWidgetComponent } from '@app/widgets/quick-actions-widget/quick-actions-widget.component';
import { MyRequestsWidgetComponent } from '@app/widgets/my-requests-widget/my-requests-widget.component';
import { WorkActivityWidgetComponent } from '@app/widgets/work-activity-widget/work-activity-widget.component';

@Component({
  selector: 'app-page-dashboard-test-role',
  standalone: true,
  imports: [
    CommonModule,
    TeamOverviewWidgetComponent,
    UpcomingEventsWidgetComponent,
    JiraSummaryWidgetComponent,
    CourseRequestsWidgetComponent,
    CurrentLearningWidgetComponent,
    RecommendationsWidgetComponent,
    QuickActionsWidgetComponent,
    MyRequestsWidgetComponent,
    WorkActivityWidgetComponent,
  ],
  templateUrl: './test-role.page.html',
  styleUrl: './test-role.page.scss',
})
export class DashboardTestRolePageComponent {}
