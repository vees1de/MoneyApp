import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { MatGridListModule } from '@angular/material/grid-list';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-jira-summary-widget',
  standalone: true,
  imports: [CommonModule, MatGridListModule, WidgetShellComponent],
  templateUrl: './jira-summary-widget.component.html',
  styleUrl: './jira-summary-widget.component.scss',
})
export class JiraSummaryWidgetComponent {}

