import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-my-requests-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, WidgetShellComponent],
  templateUrl: './my-requests-widget.component.html',
  styleUrl: './my-requests-widget.component.scss',
})
export class MyRequestsWidgetComponent {}

