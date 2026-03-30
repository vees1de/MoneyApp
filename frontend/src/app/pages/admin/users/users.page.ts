import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AdminFacade } from '@features/admin';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-admin-users',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './users.page.html',
  styleUrl: './users.page.scss',
})
export class AdminUsersPageComponent {
  private readonly facade = inject(AdminFacade);
  protected readonly routePath = '/admin/users';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
