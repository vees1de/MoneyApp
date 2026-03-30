import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-quick-actions-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatButtonModule, WidgetShellComponent],
  templateUrl: './quick-actions-widget.component.html',
  styleUrl: './quick-actions-widget.component.scss',
})
export class QuickActionsWidgetComponent {}
