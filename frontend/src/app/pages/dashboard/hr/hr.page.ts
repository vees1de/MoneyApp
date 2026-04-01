import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { CourseRequestsWidgetComponent } from '@app/widgets/course-requests-widget/course-requests-widget.component';
import { CurrentLearningWidgetComponent } from '@app/widgets/current-learning-widget/current-learning-widget.component';
import { QuickActionsWidgetComponent } from '@app/widgets/quick-actions-widget/quick-actions-widget.component';
import { RisksWidgetComponent } from '@app/widgets/risks-widget/risks-widget.component';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';

@Component({
  selector: 'app-page-dashboard-hr',
  standalone: true,
  imports: [
    CommonModule,
    QuickActionsWidgetComponent,
    UpcomingEventsWidgetComponent,
    CourseRequestsWidgetComponent,
    CurrentLearningWidgetComponent,
    RisksWidgetComponent,
  ],
  templateUrl: './hr.page.html',
  styleUrl: './hr.page.scss',
})
export class DashboardHrPageComponent {}
