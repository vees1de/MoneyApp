import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { MatListModule } from '@angular/material/list';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-team-overview-widget',
  standalone: true,
  imports: [CommonModule, MatListModule, WidgetShellComponent],
  templateUrl: './team-overview-widget.component.html',
})
export class TeamOverviewWidgetComponent {}
