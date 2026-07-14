import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateCluster, DeleteCluster, UpdateCluster, AttachCluster } from '../classes/';

@Injectable({
  providedIn: 'root'
})
export class ClustersService {
  constructor (private apollo: Apollo) { }

  getAllClusters (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Clusters(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                endpoint
                version
                type
                class
                status
                state
                tlsCertChain {
                  notAfter
                  isCA
                }
                spec {
                  __typename
                  ... on TKGSClusterSpec {
                    namespace
                    ctlPlaneClass
                    ctlPlaneCount
                    dataPlaneClass
                    dataPlaneCount
                  }
                  ... on TKGMClusterSpec {
                    namespace
                    ctlPlaneCount
                    dataPlaneCount
                    ctlPlaneCPUs
                    ctlPlaneMemMi
                    ctlPlaneStorageCapGi
                    dataPlaneCPUs
                    dataPlaneMemMi
                    dataPlaneStorageCapGi
                  }
                  ... on TKG2MClusterSpec {
                    namespace
                    ctlPlaneReplicas
                    workerNodePool0Replicas
                    ctlPlaneNumCpus
                    ctlPlaneMemMiB
                    ctlPlaneDiskGiB
                    workerNumCpus
                    workerMemMiB
                    workerDiskGiB
                  }
                }
                provider {
                  id
                  name
                  spec {
                    __typename
                    ... on TKGSClusterProviderSpec {
                      vmClasses {name spec}
                    }
                  }
                }
                createdAt
                lastModified
              }
            }
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  getClusterInfo (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Cluster(id: "${id}") {
              name
              description
              type
              class
              endpoint
              version
              status
              state
              spec {
                  __typename
                  ... on TKGSClusterSpec {
                    namespace
                    cni
                    podCIDR
                    svcCIDR
                    svcDomain
                    defStorage
                    ctlPlaneStorageCapGi
                    dataPlaneStorageCapGi
                    ctlPlaneClass
                    ctlPlaneCount
                    dataPlaneClass
                    dataPlaneCount
                  }
                  ... on TKGMClusterSpec {
                    namespace
                    cni
                    podCIDR
                    svcCIDR
                    ctlPlaneCount
                    dataPlaneCount
                    ctlPlaneCPUs
                    ctlPlaneMemMi
                    ctlPlaneStorageCapGi
                    dataPlaneCPUs
                    dataPlaneMemMi
                    dataPlaneStorageCapGi
                  }
                  ... on TKG2MClusterSpec {
                    namespace
                    cni
                    clusterCIDR
                    serviceCIDR
                    ctlPlaneReplicas
                    workerNodePool0Replicas
                    ctlPlaneNumCpus
                    ctlPlaneMemMiB
                    ctlPlaneDiskGiB
                    workerNumCpus
                    workerMemMiB
                    workerDiskGiB
                  }
              }
              extensions {
                edge {
                  nodes {
                    id
                    name
                    endpoints
                    state
                    status
                  }
                }
              }
              kubeConfig
              sshPrivateKey
              imageRegistries {
                edge {
                  nodes {
                    imageRegistry {id name endpoint}
                    default
                  }
                }
              }
              loadBalancer {id name endpoint}
              provider {
                  id
                  name
                  spec {
                    __typename
                    ... on TKGSClusterProviderSpec {
                      vmClasses {name spec}
                      clusterReleases
                    }
                    ... on TKGMClusterProviderSpec {
                      clusterReleases
                    }
                    ... on TKG2MClusterProviderSpec {
                      clusterReleases: tanzuKubernetesReleases
                    }
                  }
              }
              createdAt
              lastModified
          }
        }`,
      errorPolicy: 'all'
    });
  }

  checkAttachClusterToken (server: string, token: string, registry: string): Observable<any> {
    const input: AttachCluster = {
      name: 'check-tsjdhfksi',
      endpoint: server,
      authToken: token,
      imageRegistries: [{
        id: registry,
        default: true
      }],
      validate: true,
      dryrun: true
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: AttachCluster!){
          AttachCluster(input: $input){
            id
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  getAllHealthyClusterId (orgId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}, filters: [{key: "status", value: ["HEALTHY"]}, {key: "organization", value: ["${orgId}"]}]) {
            edge {
              nodes {
                id
                name
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getClusterAndRegistry (orgId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}, filters: [{key: "status", value: ["HEALTHY"]}, {key: "organization", value: ["${orgId}"]}]) {
            edge {
              nodes {
                id
                name
                imageRegistries {
                  edge {
                    nodes {
                      imageRegistry {
                       id
                       name
                       authUser
                      }
                    }
                  }
                }
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getTbsEnabledClusterId (orgId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}, filters: [{key: "status", value: ["HEALTHY"]}, {key: "organization", value: ["${orgId}"]}, {key: "extension", value: ["tbs"]}]) {
            totalCount
            edge {
              nodes {
                id
                name
                imageRegistries {
                  edge {
                    nodes {
                      imageRegistry {
                       id
                       name
                       authUser
                      }
                    }
                  }
                }
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getRegistryOfCluster (clusterId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Cluster(id: "${clusterId}") {
            imageRegistries {
              edge {
                nodes {
                  imageRegistry {
                   id
                   name
                   authUser
                  }
                }
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getAllDeployments (clusterId: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Deployments(filters: $filters) {
            name
            namespace
            status { readyReplicas replicas }
            createdAt
            lastModified
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  getExtensionDisabledClusterId (orgId: string, extensionName: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}, filters: [{key: "status", value: ["HEALTHY"]}, {key: "organization", value: ["${orgId}"]}, {key: "noExtension", value: ["${extensionName}"]}]) {
            edge {
              nodes {
                id
                name
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getExtensionEnabledClusterId (orgId: string, extensionName: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}, filters: [{key: "status", value: ["HEALTHY"]}, {key: "organization", value: ["${orgId}"]}, {key: "extension", value: ["${extensionName}"]}]) {
            totalCount
            edge {
              nodes {
                id
                name
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getHealthyClusterIdAndNamespace (orgId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}, filters: [{key: "status", value: ["HEALTHY"]}, {key: "organization", value: ["${orgId}"]}]) {
            edge {
              nodes {
                id
                name
                namespaces {name}
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getClusterNamespaces (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Cluster(id: "${id}") {
            namespaces {name}
          }
        }`,
      errorPolicy: 'all'
    });
  }

  attachCluster (input: AttachCluster): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: AttachCluster!){
          AttachCluster(input: $input){
            id
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  createCluster (input: CreateCluster): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateCluster!){
          CreateCluster(input: $input){
            name
            description
            endpoint
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  ReinitCluster (id: string): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      recover: true
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateClusterDesc (id: string, desc: string): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      description: desc
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKGSClusterWorkerCount (id: string, count: number): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      spec: {
        tkgs: { dataPlaneCount: count }
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKGMClusterWorkerCount (id: string, count: number): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      spec: {
        tkgm: { dataPlaneCount: count }
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKG2MClusterWorkerCount (id: string, count: number): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      spec: {
        tkg2m: { workerNodePool0Replicas: count }
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKGSClusterMasterCount (id: string): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      spec: {
        tkgs: {
          ctlPlaneCount: 3
        }
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKGMClusterMasterCount (id: string, masterNum: number): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      spec: {
        tkgm: {
          ctlPlaneCount: masterNum
        }
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKG2MClusterMasterCount (id: string, masterNum: number): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      spec: {
        tkg2m: {
          ctlPlaneReplicas: masterNum
        }
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateClusterVMClass(input: UpdateCluster): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateClusterVersion (id: string, version: string): Observable<any> {
    const input: UpdateCluster = {
      id: id,
      version: version
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateCluster!){
          UpdateCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteCluster (id: string): Observable<any> {
    const input: DeleteCluster = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteCluster!){
          DeleteCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  forceDeleteCluster (id: string): Observable<any> {
    const input: DeleteCluster = {
      id: id,
      force: true
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteCluster!){
          DeleteCluster(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }
}
