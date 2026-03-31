import {
  APP_INITIALIZER,
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
} from '@angular/core';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideRouter } from '@angular/router';

import { AuthBootstrapService } from '@core/auth/auth-bootstrap.service';
import { authHttpInterceptor } from '@core/auth/auth-http.interceptor';
import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideHttpClient(withInterceptors([authHttpInterceptor])),
    provideRouter(routes),
    {
      provide: APP_INITIALIZER,
      multi: true,
      deps: [AuthBootstrapService],
      useFactory: (bootstrap: AuthBootstrapService) => () => bootstrap.bootstrap(),
    },
  ],
};
