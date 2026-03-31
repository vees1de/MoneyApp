import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

export interface CreateTeamDialogResult {
  description?: string | null;
  name: string;
}

@Component({
  selector: 'app-create-team-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
  ],
  templateUrl: './create-team-dialog.component.html',
  styleUrl: './create-team-dialog.component.scss',
})
export class CreateTeamDialogComponent {
  private readonly fb = inject(FormBuilder);
  private readonly dialogRef = inject(
    MatDialogRef<CreateTeamDialogComponent, CreateTeamDialogResult | null>,
  );

  protected readonly form = this.fb.nonNullable.group({
    name: [''],
    description: [''],
  });

  protected save(): void {
    const values = this.form.getRawValue();
    const name = values.name.trim();
    if (!name) {
      this.form.controls.name.markAsTouched();
      return;
    }

    this.dialogRef.close({
      name,
      description: values.description.trim() || null,
    });
  }

  protected close(): void {
    this.dialogRef.close(null);
  }
}
