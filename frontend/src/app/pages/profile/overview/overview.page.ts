import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject } from '@angular/core';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';

import { AuthSessionService } from '@core/auth/auth-session.service';
import { AuthStateService } from '@core/auth/auth-state.service';
import { identityUserDisplayName } from '@core/domain/identity.util';
import { ProfileSettingsFacade } from '@features/profile-settings';

import {
  CreateTeamDialogComponent,
  type CreateTeamDialogResult,
} from './create-team-dialog.component';
import { JoinTeamDialogComponent } from './join-team-dialog.component';
import {
  ProfileEditDialogComponent,
  type ProfileEditDialogResult,
} from './profile-edit-dialog.component';

@Component({
  selector: 'app-page-profile-overview',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatCardModule, MatIconModule],
  templateUrl: './overview.page.html',
  styleUrl: './overview.page.scss',
})
export class ProfileOverviewPageComponent implements OnInit {
  private readonly facade = inject(ProfileSettingsFacade);
  private readonly authSession = inject(AuthSessionService);
  private readonly authState = inject(AuthStateService);
  private readonly router = inject(Router);
  private readonly dialog = inject(MatDialog);

  protected readonly loading = this.facade.loading;
  protected readonly saving = this.facade.saving;
  protected readonly joiningTeam = this.facade.joiningTeam;
  protected readonly leavingTeam = this.facade.leavingTeam;
  protected readonly error = this.facade.error;
  protected readonly notice = this.facade.notice;
  protected readonly profile = this.facade.profile;
  protected readonly availableRoles = this.facade.availableRoles;
  protected readonly currentTeam = this.facade.currentTeam;
  protected readonly currentUser = this.authState.currentUser;

  protected readonly profileName = computed(() => {
    const profile = this.profile();
    if (profile?.display_name?.trim()) {
      return profile.display_name.trim();
    }

    const currentUser = this.currentUser();
    if (currentUser) {
      return identityUserDisplayName(currentUser);
    }

    return profile?.email ?? 'Профиль';
  });

  protected readonly avatarInitials = computed(() => {
    const label = this.profileName()
      .split(/\s+/)
      .filter((part) => !!part)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? '')
      .join('');

    return label || 'ME';
  });

  protected readonly systemRoles = computed(() => this.currentUser()?.roles ?? []);
  protected readonly employeeProfile = computed(() => this.currentUser()?.employee_profile ?? null);

  ngOnInit(): void {
    this.facade.load();
  }

  protected roleLabel(roleCode: string): string {
    return this.authState.roleLabel(roleCode);
  }

  protected async openEditDialog(): Promise<void> {
    const profile = this.profile();
    if (!profile) {
      return;
    }

    const dialogRef = this.dialog.open(ProfileEditDialogComponent, {
      width: '760px',
      maxWidth: '96vw',
      data: {
        profile,
        availableRoles: this.availableRoles(),
      },
    });

    dialogRef.afterClosed().subscribe((result: ProfileEditDialogResult | null) => {
      if (!result) {
        return;
      }

      void this.applyProfileChanges(result);
    });
  }

  protected openCreateTeamDialog(): void {
    const dialogRef = this.dialog.open(CreateTeamDialogComponent, {
      width: '560px',
      maxWidth: '96vw',
    });

    dialogRef.afterClosed().subscribe((result: CreateTeamDialogResult | null) => {
      if (!result) {
        return;
      }

      void this.facade.createDevelopmentTeam({
        name: result.name,
        description: result.description ?? null,
      });
    });
  }

  protected openJoinTeamDialog(): void {
    const dialogRef = this.dialog.open(JoinTeamDialogComponent, {
      width: '760px',
      maxWidth: '96vw',
    });

    dialogRef.afterClosed().subscribe((teamId: string | null) => {
      if (!teamId) {
        return;
      }

      void this.facade.joinDevelopmentTeam(teamId);
    });
  }

  protected async leaveCurrentTeam(): Promise<void> {
    await this.facade.leaveCurrentDevelopmentTeam();
  }

  protected async refresh(): Promise<void> {
    await this.facade.refresh();
  }

  protected clearError(): void {
    this.facade.clearError();
  }

  protected clearNotice(): void {
    this.facade.clearMessage();
  }

  protected async logout(): Promise<void> {
    this.authSession.clear();
    this.authState.setCurrentUser(null);
    await this.router.navigateByUrl('/login');
  }

  private async applyProfileChanges(result: ProfileEditDialogResult): Promise<void> {
    const profile = this.profile();
    if (!profile) {
      return;
    }

    if (result.avatarFile) {
      const uploaded = await this.facade.uploadAvatar(result.avatarFile);
      if (!uploaded) {
        return;
      }
    }

    const displayName = result.displayName.trim() || null;
    const currentDisplayName = profile.display_name?.trim() || null;
    const newRoles = [...result.roleCodes].sort();
    const currentRoles = (profile.profile_roles ?? []).map((role) => role.code).sort();

    if (displayName === currentDisplayName && arraysEqual(newRoles, currentRoles)) {
      return;
    }

    await this.facade.saveProfile({
      display_name: displayName,
      role_codes: newRoles,
    });
  }
}

function arraysEqual(left: string[], right: string[]): boolean {
  if (left.length !== right.length) {
    return false;
  }

  return left.every((item, index) => item === right[index]);
}
