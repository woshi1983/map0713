export interface VKSOSImage {
  name: string;
  osName: string;
  osVersion: string;
  kubernetesVersion: string;
  releases: string[];
}

function releaseBase(version: string): string {
  return (version || '').replace(/^v/, '').replace(/---/g, '+').replace(/-vkr\.\d+$/, '');
}

export function availableVKSOSImages(images: VKSOSImage[], release: string): VKSOSImage[] {
  if (!release) return [];
  const base = releaseBase(release);
  const exact = images.filter(image => (image.releases || []).some(value => releaseBase(value) === base &&
    value.replace(/^v/, '').replace(/---/g, '+') === release.replace(/^v/, '').replace(/---/g, '+')));
  const hasReleaseReferences = images.some(image => (image.releases || []).some(value => releaseBase(value) === base));
  const candidates = hasReleaseReferences ? exact : images;
  return candidates.filter(image =>
    releaseBase(image.kubernetesVersion) === base &&
    (image.osName === 'photon' || image.osName === 'ubuntu') &&
    !!image.osVersion
  );
}

export function availableVKSOSNames(images: VKSOSImage[], release: string): string[] {
  const names = new Set(availableVKSOSImages(images, release).map(image => image.osName));
  return Array.from(names).sort();
}

export function availableVKSOSVersions(images: VKSOSImage[], release: string, osName: string): string[] {
  const versions = new Set(availableVKSOSImages(images, release)
    .filter(image => image.osName === osName)
    .map(image => image.osVersion));
  return Array.from(versions).sort((a, b) => a.localeCompare(b, undefined, {numeric: true}));
}
