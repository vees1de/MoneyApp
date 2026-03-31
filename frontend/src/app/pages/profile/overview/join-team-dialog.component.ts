import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';

import type { DevelopmentTeam } from '@core/api/contracts';
import { UsersApiService } from '@core/api/users-api.service';

@Component({
  selector: 'app-join-team-dialog',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
  ],
  templateUrl: './join-team-dialog.component.html',
  styleUrl: './join-team-dialog.component.scss',
})
export class JoinTeamDialogComponent implements OnInit {
  private readonly usersApi = inject(UsersApiService);
  private readonly dialogRef = inject(MatDialogRef<JoinTeamDialogComponent, string | null>);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly teams = signal<DevelopmentTeam[]>([]);
  protected readonly query = signal('');

  protected readonly filteredTeams = computed(() => {
    const normalizedQuery = this.query().trim().toLowerCase();
    if (!normalizedQuery) {
      return this.teams();
    }

    return this.teams().filter((team) =>
      [team.name, team.description ?? '', this.teamLeadLabel(team)]
        .join(' ')
        .toLowerCase()
        .includes(normalizedQuery),
    );
  });

  ngOnInit(): void {
    this.load();
  }

  protected teamLeadLabel(team: DevelopmentTeam): string {
    return team.members.find((member) => member.is_lead)?.display_name ?? 'Лид не назначен';
  }

  protected onQueryChange(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    this.query.set(input?.value ?? '');
  }

  protected selectTeam(teamId: string): void {
    this.dialogRef.close(teamId);
  }

  protected close(): void {
    this.dialogRef.close(null);
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);

    this.usersApi.listAvailableDevelopmentTeams().subscribe({
      next: (teams) => {
        this.teams.set(teams ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Не удалось загрузить список команд.');
        this.loading.set(false);
      },
    });
  }
}
