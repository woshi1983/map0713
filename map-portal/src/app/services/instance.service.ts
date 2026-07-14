import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';

@Injectable({
  providedIn: 'root'
})
export class InstanceService {
  constructor (private apollo: Apollo) { }

  getInstanceById (instanceId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Instance(id: "${instanceId}") {
            manifest
            release {name num}
            cluster {name}
            namespace
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getInstancesByCluster (num: number, limit: number, filters?: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Instances(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                namespace
                status {
                  state
                }
                release {id name num}
                application {id name}
                component {id name}
                federation {id name}
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

  // used in route
  getInstancesByComponentId (componentId: string): Observable<any> {
    const filters = [{ key: 'component', value: [componentId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Instances(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                status {
                  state
                }
                release {name num}
                services {
                  name
                  status {ingress {ip}}
                  spec {
                    type
                    ports {
                      name
                      port
                      protocol
                    }
                  }
                }
                namespace
                cluster {
                  name
                  loadBalancer {
                    id
                    type
                  }
                }
                federation {id}
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

  // used in backup
  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getInstancesByNamespace (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Instances(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                namespace
                status {
                  state
                }
                component {id name}
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

  getInstanceTaskStatus (instanceId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query{
          Instance(id: "${instanceId}") {
            status {
              tasks {
                name
                kind
                state
                message
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getInstanceService (instanceId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query{
          Instance(id: "${instanceId}") {
            services {
              name
              spec {
                type
                clusterIP
                ports {
                  port
                  nodePort
                  protocol
                }
              }
              status {
                ingress {
                  ip
                }
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getInstanceHealth (instanceId: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query{
          Instance(id: "${instanceId}") {
            health {
              workloads {
                name
                namespace
                kind
                readyReplicas
                replicas
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }
}
