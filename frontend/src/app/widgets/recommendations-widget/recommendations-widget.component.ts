import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-recommendations-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, WidgetShellComponent],
  templateUrl: './recommendations-widget.component.html',
})
export class RecommendationsWidgetComponent {}
