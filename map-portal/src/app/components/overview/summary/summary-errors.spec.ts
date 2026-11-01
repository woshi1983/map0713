import { splitSummaryErrors } from './summary-errors';

describe('splitSummaryErrors', () => {
  it('deduplicates cluster failures and preserves unrelated errors', () => {
    const result = splitSummaryErrors([
      { message: 'gateway error', extensions: { cluster: 'vks-test03' } },
      { message: 'gateway error', extensions: { cluster: 'vks-test03' } },
      { message: 'gateway error', extensions: { cluster: 'tkgmw08' } },
      { message: 'insufficient privileges' }
    ]);
    expect(result.clusters).toEqual(['vks-test03', 'tkgmw08']);
    expect(result.otherErrors).toEqual([{ message: 'insufficient privileges' }]);
  });
});
