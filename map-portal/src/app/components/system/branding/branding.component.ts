import { Component, OnInit } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { AuthService } from '../../../services/auth.service';
import { BrandingService, BrandingUpdate } from '../../../services/branding.service';

@Component({
  selector: 'app-branding',
  templateUrl: './branding.component.html',
  styleUrls: ['./branding.component.css']
})
export class BrandingComponent implements OnInit {
  nameZh = '';
  nameEn = '';
  logoDataURL = '';
  resetLogo = false;
  loading = false;
  saving = false;
  error = '';
  success = '';
  readonly maxLogoSize = 256 * 1024;

  constructor(public branding: BrandingService, public authService: AuthService) {}

  ngOnInit(): void {
    if (!this.authService.getUserInfo()?.admin) { return; }
    this.loading = true;
    this.branding.load().subscribe({
      next: config => { this.nameZh = config.nameZh || ''; this.nameEn = config.nameEn || ''; this.loading = false; },
      error: error => { this.error = this.errorMessage(error, '读取品牌设置失败'); this.loading = false; }
    });
  }

  get previewLogo(): string { return this.resetLogo ? this.branding.defaultLogoUrl : this.logoDataURL || this.branding.logoUrl; }

  selectLogo(event: Event): void {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    this.error = '';
    this.success = '';
    if (!file) { return; }
    if (!['image/png', 'image/jpeg'].includes(file.type)) {
      this.error = '请选择 PNG 或 JPEG 格式的图片。'; input.value = ''; return;
    }
    if (file.size > this.maxLogoSize) {
      this.error = '图片大小不能超过 256 KiB。'; input.value = ''; return;
    }
    const reader = new FileReader();
    reader.onerror = () => { this.error = '读取图片失败，请重试。'; input.value = ''; };
    reader.onload = () => {
      const data = reader.result;
      if (typeof data !== 'string' || !data.startsWith(`data:${file.type};base64,`)) {
        this.error = '图片内容无效。'; input.value = ''; return;
      }
      const image = new Image();
      image.onerror = () => { this.error = '无法识别图片内容。'; input.value = ''; };
      image.onload = () => {
        if (image.width > 2048 || image.height > 1024) {
          this.error = '图片尺寸不能超过 2048 × 1024 像素。'; input.value = ''; return;
        }
        this.logoDataURL = data;
        this.resetLogo = false;
      };
      image.src = data;
    };
    reader.readAsDataURL(file);
  }

  useDefaultLogo(): void {
    this.logoDataURL = '';
    this.resetLogo = true;
    this.error = '';
    this.success = '';
  }

  save(): void {
    if (!this.authService.getUserInfo()?.admin) { this.error = '仅超级管理员可以修改品牌设置。'; return; }
    const nameZh = this.nameZh.trim();
    const nameEn = this.nameEn.trim();
    if (!nameZh || !nameEn || nameZh.length > 60 || nameEn.length > 60) {
      this.error = '中英文名称均为必填，且不能超过 60 个字符。'; return;
    }
    const update: BrandingUpdate = { nameZh, nameEn };
    if (this.logoDataURL) { update.logoDataURL = this.logoDataURL; }
    if (this.resetLogo) { update.resetLogo = true; }
    this.saving = true;
    this.error = '';
    this.success = '';
    this.branding.save(update).subscribe({
      next: () => { this.saving = false; this.logoDataURL = ''; this.resetLogo = false; this.success = '品牌设置已保存。'; },
      error: error => { this.saving = false; this.error = this.errorMessage(error, '保存品牌设置失败'); }
    });
  }

  private errorMessage(error: unknown, fallback: string): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 401 || error.status === 403) { return '没有修改权限或登录已过期，请重新登录。'; }
      const detail = error.error?.error || error.error?.message;
      return detail ? `${fallback}：${detail}` : `${fallback}（HTTP ${error.status}）。`;
    }
    return error instanceof Error ? error.message : fallback;
  }
}
