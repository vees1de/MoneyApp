import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-course-requests-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, WidgetShellComponent],
  templateUrl: './course-requests-widget.component.html',
  styleUrl: './course-requests-widget.component.scss',
})
export class CourseRequestsWidgetComponent {}
