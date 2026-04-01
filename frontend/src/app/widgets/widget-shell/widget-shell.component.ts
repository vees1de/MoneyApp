import { CommonModule } from '@angular/common';
import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';

@Component({
  selector: 'app-widget-shell',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, MatButtonModule],
  templateUrl: './widget-shell.component.html',
  styleUrl: './widget-shell.component.scss',
})
export class WidgetShellComponent {
  readonly title = input.required<string>();
  readonly listRoute = input<string | null>(null);
  readonly listActionLabel = input<string>('Посмотреть еще');
  readonly actionLabel = input<string | null>(null);
  readonly actionRoute = input<string | null>(null);
  readonly loading = input(false);
  readonly error = input<string | null>(null);
  readonly empty = input(false);
}
