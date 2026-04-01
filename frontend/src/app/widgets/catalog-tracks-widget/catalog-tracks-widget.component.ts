import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

type TrackItem = {
  name: string;
  progress: number;
  courses: number;
};

@Component({
  selector: 'app-catalog-tracks-widget',
  standalone: true,
  imports: [CommonModule, WidgetShellComponent],
  templateUrl: './catalog-tracks-widget.component.html',
  styleUrl: './catalog-tracks-widget.component.scss',
})
export class CatalogTracksWidgetComponent {
  protected readonly tracks: TrackItem[] = [
    { name: 'Управление продуктом', progress: 78, courses: 6 },
    { name: 'Frontend architecture', progress: 51, courses: 4 },
    { name: 'Leadership и коммуникации', progress: 36, courses: 3 },
  ];
}
