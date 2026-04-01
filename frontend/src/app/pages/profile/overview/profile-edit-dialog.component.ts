import { CommonModule } from '@angular/common';
import { Component, OnDestroy, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import { resolveApiUrl } from '@core/api/url.util';
import type { ProfileRole, UserProfile } from '@core/api/contracts';

export interface ProfileEditDialogData {
  profile: UserProfile;
  availableRoles: ProfileRole[];
}

export interface ProfileEditDialogResult {
  avatarFile: File | null;
  displayName: string;
  roleCodes: string[];
}

@Component({
  selector: 'app-profile-edit-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
  ],
  templateUrl: './profile-edit-dialog.component.html',
  styleUrl: './profile-edit-dialog.component.scss',
})
export class ProfileEditDialogComponent implements OnDestroy {
  private readonly fb = inject(FormBuilder);
  private readonly dialogRef = inject(
    MatDialogRef<ProfileEditDialogComponent, ProfileEditDialogResult | null>,
  );
  protected readonly data = inject<ProfileEditDialogData>(MAT_DIALOG_DATA);

  protected readonly selectedRoleCodes = signal<string[]>(
    (this.data.profile.profile_roles ?? []).map((role) => role.code),
  );
  protected readonly selectedAvatarFile = signal<File | null>(null);
  protected readonly previewUrl = signal<string | null>(resolveApiUrl(this.data.profile.avatar_url));

  protected readonly form = this.fb.nonNullable.group({
    display_name: [this.data.profile.display_name ?? ''],
  });

  protected readonly avatarInitials = computed(() => {
    const value =
      this.form.controls.display_name.value.trim() ||
      this.data.profile.display_name ||
      this.data.profile.email;

    return value
      .split(/\s+/)
      .filter((part) => !!part)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? '')
      .join('');
  });

  ngOnDestroy(): void {
    const preview = this.previewUrl();
    if (preview?.startsWith('blob:')) {
      URL.revokeObjectURL(preview);
    }
  }

  protected isRoleSelected(roleCode: string): boolean {
    return this.selectedRoleCodes().includes(roleCode);
  }

  protected toggleRole(roleCode: string): void {
    this.selectedRoleCodes.update((current) =>
      current.includes(roleCode)
        ? current.filter((item) => item !== roleCode)
        : [...current, roleCode],
    );
  }

  protected onAvatarSelected(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    const file = input?.files?.[0] ?? null;
    if (!file) {
      return;
    }

    const currentPreview = this.previewUrl();
    if (currentPreview?.startsWith('blob:')) {
      URL.revokeObjectURL(currentPreview);
    }

    this.selectedAvatarFile.set(file);
    this.previewUrl.set(URL.createObjectURL(file));
  }

  protected save(): void {
    this.dialogRef.close({
      avatarFile: this.selectedAvatarFile(),
      displayName: this.form.controls.display_name.value,
      roleCodes: this.selectedRoleCodes(),
    });
  }

  protected close(): void {
    this.dialogRef.close(null);
  }
}
