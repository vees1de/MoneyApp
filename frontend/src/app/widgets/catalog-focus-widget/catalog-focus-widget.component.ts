import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';

import { WidgetShellComponent } from '@app/widgets/widget-shell/widget-shell.component';

@Component({
  selector: 'app-catalog-focus-widget',
  standalone: true,
  imports: [CommonModule, MatIconModule, WidgetShellComponent],
  templateUrl: './catalog-focus-widget.component.html',
  styleUrl: './catalog-focus-widget.component.scss',
})
export class CatalogFocusWidgetComponent {}
