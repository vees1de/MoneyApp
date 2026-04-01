import { Injectable, inject, signal, computed } from '@angular/core';
import { ReportsApiService } from '@core/api/reports-api.service';
import type { SourceMeta, SmartExportRequest } from '@entities/smart-export';

@Injectable({ providedIn: 'root' })
export class ReportsAnalyticsFacade {
  private readonly api = inject(ReportsApiService);

  readonly sources = signal<SourceMeta[]>([]);
  readonly sourcesLoading = signal(false);
  readonly sourcesError = signal<string | null>(null);
  readonly exporting = signal(false);
  readonly exportError = signal<string | null>(null);

  readonly sourcesLoaded = computed(() => this.sources().length > 0);

  load(): void {
    if (this.sourcesLoaded()) return;

    this.sourcesLoading.set(true);
    this.sourcesError.set(null);

    this.api.getSources().subscribe({
      next: (res) => {
        this.sources.set(res.sources);
        this.sourcesLoading.set(false);
      },
      error: () => {
        this.sourcesError.set('Не удалось загрузить источники данных');
        this.sourcesLoading.set(false);
      },
    });
  }

  exportToExcel(request: SmartExportRequest): void {
    this.exporting.set(true);
    this.exportError.set(null);

    this.api.smartExport(request).subscribe({
      next: (blob) => {
        this.exporting.set(false);
        this.downloadBlob(blob, request.source);
      },
      error: () => {
        this.exportError.set('Ошибка при выгрузке. Попробуйте ещё раз.');
        this.exporting.set(false);
      },
    });
  }

  private downloadBlob(blob: Blob, source: string): void {
    const now = new Date();
    const timestamp = now.toISOString().slice(0, 10).replace(/-/g, '');
    const filename = `${source}-export-${timestamp}.xlsx`;

    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  }
}
