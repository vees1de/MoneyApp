import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { CourseRequestsWidgetComponent } from '@app/widgets/course-requests-widget/course-requests-widget.component';
import { QuickActionsWidgetComponent } from '@app/widgets/quick-actions-widget/quick-actions-widget.component';
import { UpcomingEventsWidgetComponent } from '@app/widgets/upcoming-events-widget/upcoming-events-widget.component';

@Component({
  selector: 'app-page-dashboard-hr',
  standalone: true,
  imports: [
    CommonModule,
    UpcomingEventsWidgetComponent,
    CourseRequestsWidgetComponent,
    QuickActionsWidgetComponent,
  ],
  templateUrl: './hr.page.html',
  styleUrl: './hr.page.scss',
})
export class DashboardHrPageComponent {}