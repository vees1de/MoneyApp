import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatListModule, MatSelectionListChange } from '@angular/material/list';

export interface ApplicationFocusOption {
  id: string;
  applicantLabel: string;
  statusLabel: string;
}

export interface ApplicationFocusDialogData {
  applications: ApplicationFocusOption[];
  selectedApplicationId: string | null;
}

@Component({
  selector: 'app-application-focus-dialog',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, MatListModule],
  templateUrl: './application-focus-dialog.component.html',
  styleUrl: './application-focus-dialog.component.scss',
})
export class ApplicationFocusDialogComponent {
  protected readonly data = inject<ApplicationFocusDialogData>(MAT_DIALOG_DATA);
  private readonly dialogRef = inject<MatDialogRef<ApplicationFocusDialogComponent, string | null>>(
    MatDialogRef,
  );

  protected readonly selectedApplicationId = signal<string | null>(
    this.data.selectedApplicationId ?? this.data.applications[0]?.id ?? null,
  );

  protected select(event: MatSelectionListChange): void {
    const option = event.options[0];
    this.selectedApplicationId.set((option?.value as string | undefined) ?? null);
  }

  protected cancel(): void {
    this.dialogRef.close(null);
  }

  protected confirm(): void {
    this.dialogRef.close(this.selectedApplicationId());
  }
}
