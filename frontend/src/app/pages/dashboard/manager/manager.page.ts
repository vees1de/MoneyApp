import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-dashboard-manager',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './manager.page.html',
  styleUrl: './manager.page.scss',
})
export class DashboardManagerPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/dashboard/manager';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
