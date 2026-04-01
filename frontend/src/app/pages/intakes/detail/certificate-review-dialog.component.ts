import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

import type { CourseApplication } from '@core/api/contracts';

export interface CertificateReviewDialogResult {
  action: 'approve' | 'reject';
  comment: string;
}

@Component({
  selector: 'app-certificate-review-dialog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatDialogModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
  ],
  template: `
    <div class="cert-review">
      <header class="cert-review__header">
        <div class="cert-review__header-info">
          <h2>Проверка сертификата</h2>
          <p>{{ data.userName }}</p>
        </div>
        <button mat-icon-button (click)="close()">
          <mat-icon>close</mat-icon>
        </button>
      </header>

      <div class="cert-review__body">
        <div class="cert-review__preview">
          @if (data.fileUrl) {
            @if (canPreviewAsImage() && !imageError) {
              <img
                [src]="data.fileUrl"
                alt="Сертификат"
                class="cert-review__image"
                (error)="imageError = true"
              />
            } @else {
              <div class="cert-review__fallback">
                <mat-icon>description</mat-icon>
                <p>
                  {{
                    imageError
                      ? 'Не удалось загрузить изображение'
                      : 'Предпросмотр недоступен для этого типа файла'
                  }}
                </p>
                <a [href]="data.fileUrl" target="_blank" mat-stroked-button>
                  <mat-icon>open_in_new</mat-icon>
                  Открыть файл
                </a>
              </div>
            }
          } @else {
            <div class="cert-review__fallback">
              <mat-icon>hide_image</mat-icon>
              <p>Файл сертификата недоступен</p>
            </div>
          }
        </div>

        <div class="cert-review__actions-panel">
          <h3>Решение</h3>

          <mat-form-field appearance="outline" class="cert-review__comment-field">
            <mat-label>Комментарий</mat-label>
            <textarea matInput rows="4" [(ngModel)]="comment"></textarea>
          </mat-form-field>

          <div class="cert-review__buttons">
            <button
              mat-flat-button
              color="primary"
              (click)="approve()"
            >
              <mat-icon>check_circle</mat-icon>
              Подтвердить сертификат
            </button>
            <button
              mat-stroked-button
              (click)="reject()"
            >
              <mat-icon>cancel</mat-icon>
              Отклонить
            </button>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .cert-review {
      display: flex;
      flex-direction: column;
      height: 100%;
      background: var(--app-surface, #fff);
    }

    .cert-review__header {
      align-items: center;
      border-bottom: 1px solid var(--app-border, #d9dee8);
      display: flex;
      justify-content: space-between;
      padding: 16px 24px;
    }

    .cert-review__header h2 {
      font-size: 1.125rem;
      font-weight: 700;
      margin: 0;
    }

    .cert-review__header p {
      color: var(--app-muted, #969fa8);
      font-size: 0.875rem;
      margin: 4px 0 0;
    }

    .cert-review__body {
      display: grid;
      flex: 1;
      grid-template-columns: 1fr 360px;
      min-height: 0;
    }

    .cert-review__preview {
      align-items: center;
      background: var(--app-surface-soft, #f2f3f7);
      display: flex;
      justify-content: center;
      overflow: auto;
      padding: 24px;
    }

    .cert-review__image {
      border-radius: 8px;
      box-shadow: var(--shadow-card, 0 1px 2px rgb(15 23 42 / 8%));
      max-height: 100%;
      max-width: 100%;
      object-fit: contain;
    }

    .cert-review__fallback {
      align-items: center;
      color: var(--app-muted, #969fa8);
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .cert-review__fallback mat-icon {
      font-size: 48px;
      height: 48px;
      width: 48px;
    }

    .cert-review__actions-panel {
      border-left: 1px solid var(--app-border, #d9dee8);
      display: flex;
      flex-direction: column;
      gap: 16px;
      padding: 24px;
    }

    .cert-review__actions-panel h3 {
      font-size: 1rem;
      font-weight: 700;
      margin: 0;
    }

    .cert-review__comment-field {
      width: 100%;
    }

    .cert-review__buttons {
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    @media (max-width: 768px) {
      .cert-review__body {
        grid-template-columns: 1fr;
        grid-template-rows: 1fr auto;
      }

      .cert-review__actions-panel {
        border-left: none;
        border-top: 1px solid var(--app-border, #d9dee8);
      }
    }
  `],
})
export class CertificateReviewDialogComponent {
  comment = '';
  imageError = false;

  constructor(
    @Inject(MAT_DIALOG_DATA)
    public data: {
      application: CourseApplication;
      userName: string;
      fileUrl: string | null;
      fileName: string | null;
    },
    private dialogRef: MatDialogRef<CertificateReviewDialogComponent, CertificateReviewDialogResult | null>,
  ) {}

  canPreviewAsImage(): boolean {
    const fileName = this.data.fileName?.trim();
    if (!fileName) {
      return false;
    }

    return /\.(png|jpe?g|gif|webp|bmp|svg)$/i.test(fileName);
  }

  close(): void {
    this.dialogRef.close(null);
  }

  approve(): void {
    this.dialogRef.close({ action: 'approve', comment: this.comment.trim() });
  }

  reject(): void {
    this.dialogRef.close({ action: 'reject', comment: this.comment.trim() });
  }
}
