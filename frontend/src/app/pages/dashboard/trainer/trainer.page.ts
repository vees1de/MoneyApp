import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { CourseRequestsWidgetComponent } from '@app/widgets/course-requests-widget/course-requests-widget.component';
import { CurrentLearningWidgetComponent } from '@app/widgets/current-learning-widget/current-learning-widget.component';
import { JiraSummaryWidgetComponent } from '@app/widgets/jira-summary-widget/jira-summary-widget.component';
import { QuickActionsWidgetComponent } from '@app/widgets/quick-actions-widget/quick-actions-widget.component';
import { RecommendationsWidgetComponent } from '@app/widgets/recommendations-widget/recommendations-widget.component';
import { TeamOverviewWidgetComponent } from '@app/widgets/team-overview-widget/team-overview-widget.component';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';

@Component({
  selector: 'app-page-dashboard-trainer',
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
  ],
  templateUrl: './trainer.page.html',
  styleUrl: './trainer.page.scss',
})
export class DashboardTrainerPageComponent {}