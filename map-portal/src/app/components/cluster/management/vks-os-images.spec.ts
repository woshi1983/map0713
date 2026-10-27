import {availableVKSOSImages, availableVKSOSNames, availableVKSOSVersions, VKSOSImage} from './vks-os-images';

describe('VKS OS image choices', () => {
  const images: VKSOSImage[] = [
    {name: 'old', osName: 'photon', osVersion: '4.0', kubernetesVersion: 'v1.33.0+vmware.1', releases: []},
    {name: 'photon-5', osName: 'photon', osVersion: '5.0', kubernetesVersion: 'v1.34.2+vmware.2', releases: ['v1.34.2+vmware.2-vkr.2']},
    {name: 'photon-6', osName: 'photon', osVersion: '6.0', kubernetesVersion: 'v1.34.2+vmware.2', releases: ['v1.34.2+vmware.2-vkr.1']},
    {name: 'ubuntu', osName: 'ubuntu', osVersion: '24.04', kubernetesVersion: 'v1.34.2+vmware.2', releases: ['v1.34.2+vmware.2-vkr.2']},
    {name: 'worker-only', osName: 'windows', osVersion: '2025', kubernetesVersion: 'v1.34.2+vmware.2', releases: ['v1.34.2+vmware.2-vkr.2']},
  ];

  it('limits choices to the selected release and control-plane OS types', () => {
    const release = 'v1.34.2---vmware.2-vkr.2';
    expect(availableVKSOSNames(images, release)).toEqual(['photon', 'ubuntu']);
    expect(availableVKSOSVersions(images, release, 'photon')).toEqual(['5.0']);
    expect(availableVKSOSVersions(images, release, 'ubuntu')).toEqual(['24.04']);
    expect(availableVKSOSNames(images, '1.34.2+vmware.2-vkr.2')).toEqual(['photon', 'ubuntu']);
    expect(availableVKSOSVersions(images, 'v1.34.2+vmware.2-vkr.1', 'photon')).toEqual(['6.0']);
    expect(availableVKSOSNames(images, 'v1.34.2+vmware.2-vkr.3')).toEqual([]);
    expect(availableVKSOSImages(images, 'v1.35.0+vmware.1')).toEqual([]);
  });
});
