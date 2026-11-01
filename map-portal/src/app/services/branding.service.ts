import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { CookieService } from 'ngx-cookie';
import { TranslateService } from '@ngx-translate/core';
import { Observable, throwError } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { TOKEN_COOKIE_KEY, LANG_LOCALSTORAGE_KEY, DeFaultLang } from '../utils/static-entries/shared.const';

export interface BrandingConfig {
  nameZh: string;
  nameEn: string;
  logoUrl: string;
}

export interface BrandingUpdate {
  nameZh: string;
  nameEn: string;
  logoDataURL?: string;
  resetLogo?: boolean;
}

@Injectable({ providedIn: 'root' })
export class BrandingService {
  private config: BrandingConfig = { nameZh: '', nameEn: '', logoUrl: '' };
  private logoRevision = 0;

  constructor(private http: HttpClient, private cookie: CookieService, private translate: TranslateService) {
    this.translate.onLangChange.subscribe(() => this.updateTitle());
    this.updateTitle();
  }

  get name(): string {
    const language = this.translate.currentLang || localStorage.getItem(LANG_LOCALSTORAGE_KEY) || DeFaultLang;
    const configured = language.toLowerCase().startsWith('zh') ? this.config.nameZh : this.config.nameEn;
    const fallback = this.translate.instant('LOGINPAGE.TITLE');
    return configured || (fallback === 'LOGINPAGE.TITLE' ? (language.startsWith('zh') ? '应用现代化平台' : 'Application Modernization Platform') : fallback);
  }

  get logoUrl(): string {
    const logo = this.config.logoUrl;
    if (!logo) { return environment.assetsPath + 'images/logo.png'; }
    if (!this.logoRevision) { return logo; }
    return logo + (logo.includes('?') ? '&' : '?') + 'v=' + this.logoRevision;
  }

  get defaultLogoUrl(): string { return environment.assetsPath + 'images/logo.png'; }

  get current(): BrandingConfig { return { ...this.config }; }

  load(): Observable<BrandingConfig> {
    return this.http.get<BrandingConfig>('/api/branding').pipe(tap(config => this.apply(config)));
  }

  save(update: BrandingUpdate): Observable<BrandingConfig> {
    const raw = this.cookie.get(TOKEN_COOKIE_KEY)?.trim();
    const parts = raw?.split('.');
    if (!parts || parts.length < 2) { return throwError(new Error('登录已过期，请重新登录。')); }
    let token: string;
    try { token = atob(parts[0]); } catch (_) { return throwError(new Error('登录凭据无效，请重新登录。')); }
    if (!token) { return throwError(new Error('登录凭据无效，请重新登录。')); }
    const headers = new HttpHeaders({ Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' });
    return this.http.put<BrandingConfig>('/api/branding', update, { headers })
      .pipe(tap(config => this.apply(config, !!(update.logoDataURL || update.resetLogo))));
  }

  private apply(config: BrandingConfig, logoChanged = false): void {
    this.config = {
      nameZh: typeof config?.nameZh === 'string' ? config.nameZh : '',
      nameEn: typeof config?.nameEn === 'string' ? config.nameEn : '',
      logoUrl: typeof config?.logoUrl === 'string' ? config.logoUrl : ''
    };
    if (logoChanged) { this.logoRevision = Date.now(); }
    this.updateTitle();
  }

  private updateTitle(): void { document.title = this.name; }
}
