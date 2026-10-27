interface VKSReleaseVersion {
  major: number;
  minor: number;
  patch: number;
  vmware: number;
  vkr: number;
}

function parseVKSRelease(value: string): VKSReleaseVersion | null {
  const match = /^v?(\d+)\.(\d+)\.(\d+)\+vmware\.(\d+)-vkr\.(\d+)$/.exec(
    (value || '').replace('---', '+'));
  if (!match) return null;
  const [major, minor, patch, vmware, vkr] = match.slice(1).map(Number);
  return {major, minor, patch, vmware, vkr};
}

function compareVKSRelease(a: VKSReleaseVersion, b: VKSReleaseVersion): number {
  for (const part of ['major', 'minor', 'patch', 'vmware', 'vkr'] as const) {
    const difference = a[part] - b[part];
    if (difference) return difference;
  }
  return 0;
}

export function availableVKSUpgradeVersions(releases: string[], currentVersion: string): string[] {
  const current = parseVKSRelease(currentVersion);
  if (!current) return [];
  return Array.from(new Set(releases || []))
    .filter(release => {
      const version = parseVKSRelease(release);
      return version && version.major === current.major &&
        ((version.minor === current.minor && version.patch > current.patch) ||
          version.minor === current.minor + 1);
    })
    .sort((a, b) => compareVKSRelease(parseVKSRelease(a), parseVKSRelease(b)));
}

export function vksVersionForMutation(release: string): string {
  return release.replace('---', '+');
}
