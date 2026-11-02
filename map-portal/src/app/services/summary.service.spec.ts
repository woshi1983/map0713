import { of } from 'rxjs';
import { print } from 'graphql';

import { SummaryService } from './summary.service';

describe('SummaryService basic metrics', () => {
  it('skips platform-wide active users for a project user', () => {
    const query = jasmine.createSpy('query').and.returnValue(of({ data: {} }));
    const service = new SummaryService({ query } as any);

    service.getBasicMetrics(false).subscribe();

    const options = query.calls.mostRecent().args[0] as any;
    expect(options.variables.includePlatformMetrics).toBe(false);
    expect(print(options.query)).toContain('Metrics @include(if: $includePlatformMetrics)');
  });

  it('includes platform-wide active users for a super administrator', () => {
    const query = jasmine.createSpy('query').and.returnValue(of({ data: {} }));
    const service = new SummaryService({ query } as any);

    service.getBasicMetrics(true).subscribe();

    expect((query.calls.mostRecent().args[0] as any).variables.includePlatformMetrics).toBe(true);
  });
});
