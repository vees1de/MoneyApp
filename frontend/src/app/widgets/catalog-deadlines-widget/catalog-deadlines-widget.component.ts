import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

type IntakePoint = {
  title: string;
  dateLabel: string;
  placesLeft: string;
};

@Component({
  selector: 'app-catalog-deadlines-widget',
  standalone: true,
  imports: [CommonModule, MatIconModule, WidgetShellComponent],
  templateUrl: './catalog-deadlines-widget.component.html',
  styleUrl: './catalog-deadlines-widget.component.scss',
})
export class CatalogDeadlinesWidgetComponent {
  protected readonly intakePoints: IntakePoint[] = [
    { title: 'Product Discovery', dateLabel: '15 апреля', placesLeft: '8 мест' },
    { title: 'SQL для аналитиков', dateLabel: '19 апреля', placesLeft: '12 мест' },
    { title: 'Системный дизайн', dateLabel: '23 апреля', placesLeft: '5 мест' },
  ];
}
