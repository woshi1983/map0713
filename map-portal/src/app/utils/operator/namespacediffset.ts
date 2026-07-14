import { Namespace, TKGMClusterProviderSpecNamespace } from '../../classes';

export class Namespacediffset {
  public static excludedNamespaceNmaeList = [
    'default', 'build-service', 'kpack', 'kube-node-lease', 'kube-public', 'kube-system', 'stacks-operator-system',
    'vmware-system-auth', 'vmware-system-cloud-provider', 'vmware-system-csi', 'vmware-system-map', 'avi-system',
    'tkg-system', 'tkg-system-public', 'tanzu-package-repo-global', 'cert-manager', 'tanzu-system-dashboards',
    'tanzu-system-ingress', 'tanzu-system-build-logging', 'tanzu-system-monitoring', 'vmware-system-elastic', 'vmware-system-octant', 'tanzu-system', 'tkg-system-telemetry', 'cert-injection-webhook'];

  public static excludedProviderNamespaceList = [
    'default', 'capi-kubeadm-bootstrap-system', 'capi-kubeadm-control-plane-system',
    'capi-system', 'capv-system', 'cert-manager', 'kube-node-lease', 'kube-public',
    'tanzu-system', 'tkg-system-telemetry', 'kube-system', 'tanzu-package-repo-global',
    'tkg-system', 'tkg-system-networking', 'tkg-system-public', 'tkr-system'];
}

export function namespacediffset (currentList: Namespace[]): Namespace[] {
  const excludeNameList: string[] = Namespacediffset.excludedNamespaceNmaeList;
  const outputNamespaceList: Namespace[] = [];
  for (let i = 0; i < currentList.length; i++) {
    if (!excludeNameList.find(x => x === currentList[i].name)) {
      outputNamespaceList.push(currentList[i]);
    }
  }
  return outputNamespaceList;
}

export function prividerNamespacediffset (currentList: TKGMClusterProviderSpecNamespace[]): TKGMClusterProviderSpecNamespace[] {
  const excludeNameList: string[] = Namespacediffset.excludedProviderNamespaceList;
  const outputNamespaceList: TKGMClusterProviderSpecNamespace[] = [];
  for (let i = 0; i < currentList.length; i++) {
    if (!excludeNameList.find(x => x === currentList[i].name)) {
      outputNamespaceList.push(currentList[i]);
    }
  }
  return outputNamespaceList;
}
