import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CookieService } from 'ngx-cookie';
import { BrandingService } from './branding.service';

/** Branding is public to read; only writes carry the current user token. */
describe('BrandingService', () => {
  let service: BrandingService;
  let http: HttpTestingController;
  let translate: TranslateService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule, TranslateModule.forRoot()],
      providers: [{ provide: CookieService, useValue: { get: () => `${btoa('test-token')}.signature` } }]
    });
    service = TestBed.inject(BrandingService);
    http = TestBed.inject(HttpTestingController);
    translate = TestBed.inject(TranslateService);
  });

  afterEach(() => http.verify());

  it('loads public branding and updates the browser title for the selected language', () => {
    translate.use('zh-cn');
    service.load().subscribe();
    const request = http.expectOne('/api/branding');
    expect(request.request.headers.has('Authorization')).toBeFalse();
    request.flush({ nameZh: '测试平台', nameEn: 'Test Platform', logoUrl: '/api/branding/logo' });
    expect(service.name).toBe('测试平台');
    expect(document.title).toBe('测试平台');
    expect(service.logoUrl).toBe('/api/branding/logo');
    translate.use('en-us');
    expect(service.name).toBe('Test Platform');
  });

  it('sends an authenticated update and refreshes the displayed values', () => {
    service.save({ nameZh: '新名称', nameEn: 'New name', resetLogo: true }).subscribe();
    const request = http.expectOne('/api/branding');
    expect(request.request.method).toBe('PUT');
    expect(request.request.headers.get('Authorization')).toBe('Bearer test-token');
    expect(request.request.body.resetLogo).toBeTrue();
    request.flush({ nameZh: '新名称', nameEn: 'New name', logoUrl: '' });
    expect(service.current.nameEn).toBe('New name');
  });
});
