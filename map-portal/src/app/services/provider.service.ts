import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import {
  AddClusterProviderTrustedCA,
  RemoveClusterProviderTrustedCA,
  DeleteClusterProvider,
  CreateTKGSClusterProvider,
  UpdateTKGSClusterProvider,
  CreateTKGMClusterProvider,
  UpdateTKGMClusterProvider,
  CreateTKG2MClusterProvider,
  UpdateTKG2MClusterProvider
} from '../classes';

@Injectable({
  providedIn: 'root'
})
export class ProviderService {
  constructor (private apollo: Apollo) {
  }

  getAllProviders (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ClusterProviders(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                class
                spec {
                __typename
                  ... on TKGSClusterProviderSpec {
                    endpoint
                    caCert
                    insecureSkipTLSVerify
                    authUser
                    defaultNamespace
                  }
                  ... on TKGMClusterProviderSpec {
                    endpoint
                    defaultNamespace
                    defaultEnv
                  }
                  ... on TKG2MClusterProviderSpec {
                    endpoint
                    defaultNamespace
                  }
                }
                tlsCertChain {
                  notAfter
                  isCA
                }
                imageRegistries {
                  edge {
                    nodes {
                      imageRegistry {id name}
                      default
                    }
                  }
                }
                loadBalancer {id name}
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

  getProviderNames (orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ClusterProviders(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                class
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

  getProviderById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        ClusterProvider(id: "${id}") {
          name
          description
          class
          spec {
            __typename
            ... on TKGSClusterProviderSpec {
                defaultNamespace
                endpoint
                caCert
                insecureSkipTLSVerify
                authUser
             }
            ... on TKGMClusterProviderSpec {
                defaultNamespace
                defaultEnv
            }
            ... on TKG2MClusterProviderSpec {
                defaultNamespace
            }
          }
          imageRegistries {
            edge {
              nodes {
                imageRegistry {id name}
                default
              }
            }
          }
          loadBalancer {id name}
          createdAt
          lastModified
        }
      }`,
      errorPolicy: 'all'
    });
  }

  getTKGSClusterProfileByProvider (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        ClusterProvider(id: "${id}") {
          spec {
            __typename
            ... on TKGSClusterProviderSpec {
                namespaces {
                  name
                  storages
                }
                vmClasses { name spec }
                clusterReleases
                defaultNamespace
             }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  getTKGMClusterProfileByProvider (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        ClusterProvider(id: "${id}") {
          spec {
            __typename
            ... on TKGMClusterProviderSpec {
                clusterReleases
                defaultNamespace
            }
            ... on TKG2MClusterProviderSpec {
                clusterReleases: tanzuKubernetesReleases
                defaultNamespace
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  getAllNamespaceName (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        ClusterProvider(id: "${id}") {
          spec {
            ... on TKGSClusterProviderSpec {
                namespaces {
                  name
                }
              }
            ... on TKGMClusterProviderSpec {
                namespaces {
                  name
                }
            }
            ... on TKG2MClusterProviderSpec {
                namespaces {
                  name
                }
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  getNamespaceByProvider (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        ClusterProvider(id: "${id}") {
          class
          spec {
            __typename
            ... on TKGSClusterProviderSpec {
                namespaces {
                  name
                  storages
                  createdAt
                  lastModified
                }
              }
            ... on TKGMClusterProviderSpec {
                namespaces {
                  name
                  createdAt
                  lastModified
                }
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  checkCreateTKGSProviderConnection (input: CreateTKGSClusterProvider): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateTKGSClusterProvider!){
          CreateTKGSClusterProvider(input: $input){
            id
            name
            spec {
              __typename
              ... on TKGSClusterProviderSpec {
                  namespaces {
                    name
                  }
              }
            }
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  checkCreateTKGMProviderConnection (kubeconfig: string, registry: string): Observable<any> {
    const input: CreateTKG2MClusterProvider = {
      name: 'check-wsdertt',
      spec: {
        kubeConfig: kubeconfig,
        defaultNamespace: 'map-check'
      },
      imageRegistries: [{
        id: registry,
        default: true
      }],
      validate: true,
      dryrun: true
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateTKG2MClusterProvider!){
          CreateTKG2MClusterProvider(input: $input){
            id
            name
            spec {
              __typename
              ... on TKG2MClusterProviderSpec {
                  namespaces {
                    name
                  }
              }
            }
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  checkUpdateTKGSProviderConnection (input: UpdateTKGSClusterProvider): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKGSClusterProvider!){
          UpdateTKGSClusterProvider(input: $input){
            id
            name
            spec {
              __typename
              ... on TKGSClusterProviderSpec {
                  namespaces {
                    name
                  }
              }
            }
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  checkUpdateTKGMProviderConnection (id: string, kubeconfig: string): Observable<any> {
    const input: UpdateTKG2MClusterProvider = {
      id: id,
      spec: {
        kubeConfig: kubeconfig
      },
      validate: true,
      dryrun: true
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKG2MClusterProvider!){
          UpdateTKG2MClusterProvider(input: $input){
            id
            name
            spec {
              __typename
              ... on TKG2MClusterProviderSpec {
                  namespaces {
                    name
                  }
              }
            }
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  createTKGSProvider (input: CreateTKGSClusterProvider): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateTKGSClusterProvider!){
          CreateTKGSClusterProvider(input: $input){
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

  createTKGMProvider (input: CreateTKGMClusterProvider): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateTKG2MClusterProvider!){
          CreateTKG2MClusterProvider(input: $input){
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

  deleteProvider (id: string): Observable<any> {
    const input: DeleteClusterProvider = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteClusterProvider!){
          DeleteClusterProvider(input: $input){
            id
            name
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateTKGSProvider (input: UpdateTKGSClusterProvider): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKGSClusterProvider!){
          UpdateTKGSClusterProvider(input: $input){
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

  updateTKGMProvider (input: UpdateTKGMClusterProvider): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKG2MClusterProvider!){
          UpdateTKG2MClusterProvider(input: $input){
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

  getCaTrustByProvider (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        ClusterProvider(id: "${id}") {
          class
          spec {
            __typename
            ... on TKGSClusterProviderSpec {
                trustedCAs {
                  name
                  cert
                }
            }
            ... on TKGMClusterProviderSpec {
                trustedCAs {
                  name
                  cert
                }
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }

  addTKGSProviderCaTrust (id: string, caList: AddClusterProviderTrustedCA[]): Observable<any> {
    const input: UpdateTKGSClusterProvider = {
      id: id,
      spec: {
        addTrustedCAs: caList
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKGSClusterProvider!){
          UpdateTKGSClusterProvider(input: $input){
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

  addTKGMProviderCaTrust (id: string, caList: AddClusterProviderTrustedCA[]): Observable<any> {
    const input: UpdateTKGMClusterProvider = {
      id: id,
      spec: {
        addTrustedCAs: caList
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKGMClusterProvider!){
          UpdateTKGMClusterProvider(input: $input){
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

  deleteTKGSProviderCaTrust (id: string, caList: RemoveClusterProviderTrustedCA[]): Observable<any> {
    const input: UpdateTKGSClusterProvider = {
      id: id,
      spec: {
        removeTrustedCAs: caList
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKGSClusterProvider!){
          UpdateTKGSClusterProvider(input: $input){
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

  deleteTKGMProviderCaTrust (id: string, caList: RemoveClusterProviderTrustedCA[]): Observable<any> {
    const input: UpdateTKGMClusterProvider = {
      id: id,
      spec: {
        removeTrustedCAs: caList
      }
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateTKGMClusterProvider!){
          UpdateTKGMClusterProvider(input: $input){
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
}
