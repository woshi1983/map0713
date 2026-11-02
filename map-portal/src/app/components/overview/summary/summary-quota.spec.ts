import { of } from 'rxjs';

import { SummaryComponent } from './summary.component';

describe('Summary quota metrics', () => {
  it('includes resource usage from unlimited projects without treating finite quotas as a global limit', () => {
    const data = { data: { Organizations: { edge: { nodes: [
      { name: 'default', quota: { used: { cpus: 360, memoryMi: 1226744, storageGi: 2790 }, total: { cpus: -1, memoryMi: -1, storageGi: -1 } } },
      { name: 'unlimited', quota: { used: { cpus: 28, memoryMi: 106496, storageGi: 280 }, total: { cpus: 0, memoryMi: 0, storageGi: 0 } } },
      { name: 'limited', quota: { used: { cpus: 2, memoryMi: 4096, storageGi: 25 }, total: { cpus: 4, memoryMi: 8192, storageGi: 100 } } }
    ] } } } };
    const component = new SummaryComponent(
      {} as any,
      { getOrgQuota: () => of(data) } as any,
      { show: () => undefined } as any,
      { instant: (key: string) => key } as any
    );

    component.refreshOrgResourceMerics();

    expect(component.allOrgResourceMetrics.totalResource).toEqual({
      cpu: { used: 390, total: 4 },
      memory: { used: 1305.992, total: 8 },
      storage: { used: 3095, total: 100 }
    });
    expect(component.unlimitedQuota).toEqual({ cpu: true, memory: true, storage: true });
    expect(component.allOrgResourceMetrics.orgMetrics.name).toEqual(['default', 'unlimited', 'limited']);
    expect(component.allOrgResourceMetrics.orgMetrics.cpuFree).toEqual([0, 0, 2]);
    expect(component.allOrgResourceMetrics.orgMetrics.memoryFree).toEqual([0, 0, 4]);
    expect(component.allOrgResourceMetrics.orgMetrics.storageFree).toEqual([0, 0, 75]);
    expect((component.allTenantCPUOption.series[0] as any).type).toBe('pie');
    expect((component.allTenantCPUOption.series[0] as any).data).toEqual([
      { name: 'default', value: 360 },
      { name: 'unlimited', value: 28 },
      { name: 'limited', value: 2 }
    ]);
    expect((component.allTenantMemoryOption.series[0] as any).type).toBe('pie');
    expect((component.allTenantStoageOption.series[0] as any).type).toBe('pie');
  });
});
