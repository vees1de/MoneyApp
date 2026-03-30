import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ProfileSettingsFacade } from '@features/profile-settings';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-profile-overview',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class ProfileOverviewPageComponent {
  private readonly facade = inject(ProfileSettingsFacade);
  protected readonly routePath = '/profile/overview';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }
}
