import { CommonModule, Location } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { AuthStateService } from '@core/auth/auth-state.service';
import { SidebarComponent } from '@core/layout/sidebar/sidebar.component';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [CommonModule, RouterOutlet, SidebarComponent],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.scss',
})
export class AppShellComponent {
  private readonly location = inject(Location);
  protected readonly authState = inject(AuthStateService);

  protected goBack(): void {
    this.location.back();
  }
}
