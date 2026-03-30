import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthFacade } from '@features/auth';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-dashboard-admin',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './admin.page.html',
  styleUrl: './admin.page.scss',
})
export class DashboardAdminPageComponent {
  private readonly facade = inject(AuthFacade);
  protected readonly routePath = '/dashboard/admin';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
