import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';

import { AuthSessionService } from '@core/auth/auth-session.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { ProfileSettingsFacade } from '@features/profile-settings';
import type { User } from '@entities/user';

@Component({
  selector: 'app-page-profile-overview',
  standalone: true,
  imports: [CommonModule, MatButtonModule],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class ProfileOverviewPageComponent {
  private readonly facade = inject(ProfileSettingsFacade);
  private readonly authSession = inject(AuthSessionService);
  private readonly authState = inject(AuthStateService);
  private readonly router = inject(Router);

  protected readonly routePath = '/profile/overview';
  protected readonly entitySample: User[] = [];

  protected loadPage(): void {
    this.facade.load();
  }

  protected async logout(): Promise<void> {
    this.authSession.clear();
    this.authState.setCurrentUser(null);
    await this.router.navigateByUrl('/login');
  }
}
