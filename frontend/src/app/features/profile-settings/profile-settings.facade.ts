import { HttpErrorResponse } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import type {
  CreateDevelopmentTeamRequest,
  DevelopmentTeam,
  ProfileMeResponse,
  ProfileRole,
  UpdateUserProfileRequest,
  UserProfile,
} from '@core/api/contracts';
import { UsersApiService } from '@core/api/users-api.service';

@Injectable({ providedIn: 'root' })
export class ProfileSettingsFacade {
  private readonly usersApi = inject(UsersApiService);

  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly joiningTeam = signal(false);
  readonly leavingTeam = signal(false);
  readonly error = signal<string | null>(null);
  readonly notice = signal<string | null>(null);

  readonly profile = signal<UserProfile | null>(null);
  readonly availableRoles = signal<ProfileRole[]>([]);
  readonly teams = computed(() => this.profile()?.teams ?? []);
  readonly currentTeam = computed<DevelopmentTeam | null>(() => this.profile()?.teams?.[0] ?? null);

  load(): void {
    if (this.loading()) {
      return;
    }

    void this.refresh();
  }

  async refresh(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);

    try {
      const response = await firstValueFrom(this.usersApi.me());
      this.applyProfileResponse(response);
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось загрузить профиль.'));
    } finally {
      this.loading.set(false);
    }
  }

  clearMessage(): void {
    this.notice.set(null);
  }

  clearError(): void {
    this.error.set(null);
  }

  async saveProfile(payload: UpdateUserProfileRequest): Promise<boolean> {
    this.saving.set(true);
    this.error.set(null);
    this.notice.set(null);

    try {
      const response = await firstValueFrom(this.usersApi.updateMe(payload));
      this.applyProfileResponse(response);
      this.notice.set('Настройки профиля сохранены.');
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось сохранить профиль.'));
      return false;
    } finally {
      this.saving.set(false);
    }
  }

  async uploadAvatar(file: File): Promise<boolean> {
    this.saving.set(true);
    this.error.set(null);
    this.notice.set(null);

    try {
      const response = await firstValueFrom(this.usersApi.uploadAvatar(file));
      this.applyProfileResponse(response);
      this.notice.set('Фото профиля обновлено.');
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось загрузить фото.'));
      return false;
    } finally {
      this.saving.set(false);
    }
  }

  async createDevelopmentTeam(payload: CreateDevelopmentTeamRequest): Promise<boolean> {
    this.joiningTeam.set(true);
    this.error.set(null);
    this.notice.set(null);

    try {
      const response = await firstValueFrom(this.usersApi.createDevelopmentTeam(payload));
      this.applyProfileResponse(response);
      this.notice.set('Команда разработки создана.');
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось создать команду.'));
      return false;
    } finally {
      this.joiningTeam.set(false);
    }
  }

  async joinDevelopmentTeam(teamId: string): Promise<boolean> {
    this.joiningTeam.set(true);
    this.error.set(null);
    this.notice.set(null);

    try {
      const response = await firstValueFrom(this.usersApi.joinDevelopmentTeam(teamId));
      this.applyProfileResponse(response);
      this.notice.set('Вы вступили в команду разработки.');
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось вступить в команду.'));
      return false;
    } finally {
      this.joiningTeam.set(false);
    }
  }

  async leaveCurrentDevelopmentTeam(): Promise<boolean> {
    this.leavingTeam.set(true);
    this.error.set(null);
    this.notice.set(null);

    try {
      const response = await firstValueFrom(this.usersApi.leaveCurrentDevelopmentTeam());
      this.applyProfileResponse(response);
      this.notice.set('Вы вышли из команды разработки.');
      return true;
    } catch (error) {
      this.error.set(this.extractErrorMessage(error, 'Не удалось выйти из команды.'));
      return false;
    } finally {
      this.leavingTeam.set(false);
    }
  }

  private applyProfileResponse(response: ProfileMeResponse): void {
    this.profile.set(response.profile);
    this.availableRoles.set(response.available_profile_roles ?? []);
  }

  private extractErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof HttpErrorResponse) {
      const apiMessage = error.error?.error?.message;
      if (typeof apiMessage === 'string' && apiMessage.trim()) {
        return apiMessage.trim();
      }
    }

    return fallback;
  }
}
