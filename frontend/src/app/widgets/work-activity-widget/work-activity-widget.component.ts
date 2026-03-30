import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-work-activity-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, WidgetShellComponent],
  templateUrl: './work-activity-widget.component.html',
  styleUrl: './work-activity-widget.component.scss',
})
export class WorkActivityWidgetComponent {}

