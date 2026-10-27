import {availableVKSUpgradeVersions, vksVersionForMutation} from './vks-upgrade-versions';

describe('VKS upgrade versions', () => {
  it('offers later VKS releases within the current or next Kubernetes minor', () => {
    const versions = [
      'v1.34.2---vmware.1-vkr.2',
      'v1.35.2---vmware.1-vkr.1',
      'v1.35.2---vmware.1-vkr.3',
      'v1.35.3---vmware.1-vkr.1',
      'v1.36.0---vmware.1-vkr.1',
      'v1.37.0---vmware.1-vkr.1'
    ];
    expect(availableVKSUpgradeVersions(versions, 'v1.35.2---vmware.1-vkr.1')).toEqual([
      'v1.35.3---vmware.1-vkr.1',
      'v1.36.0---vmware.1-vkr.1'
    ]);
  });

  it('does not treat an equivalent release spelling as an upgrade', () => {
    expect(availableVKSUpgradeVersions(['v1.35.2+vmware.1-vkr.3'],
      'v1.35.2---vmware.1-vkr.3')).toEqual([]);
  });

  it('sends VKS build metadata in GraphQL version form', () => {
    expect(vksVersionForMutation('v1.36.0---vmware.1-vkr.1'))
      .toBe('v1.36.0+vmware.1-vkr.1');
  });
});
