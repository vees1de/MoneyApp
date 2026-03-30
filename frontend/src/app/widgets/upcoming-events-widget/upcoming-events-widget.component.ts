import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-upcoming-events-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, WidgetShellComponent],
  templateUrl: './upcoming-events-widget.component.html',
})
export class UpcomingEventsWidgetComponent {}
