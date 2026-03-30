import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatListModule } from '@angular/material/list';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-current-learning-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, MatListModule, MatProgressBarModule, WidgetShellComponent],
  templateUrl: './current-learning-widget.component.html',
  styleUrl: './current-learning-widget.component.scss',
})
export class CurrentLearningWidgetComponent {}

