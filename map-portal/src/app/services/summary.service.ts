import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class SummaryService {
  constructor (private apollo: Apollo) { }

  getBasicMetrics (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Organizations(page: {limit: -1}) {
            totalCount
          }
          Metrics {
            activeUsers
          }
          Components(page: {limit: -1}) {
            totalCount
          }
          Instances(page: {limit: -1}) {
            totalCount
          }
          BinaryImages {
            name
          }
          DockerfileImages {
            name
          }
          SourceImages {
            name
          }
          Ingresses {
            name
          }
          Routes(page: {limit: -1}) {
            totalCount
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getAllClustersStatus (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Clusters(page: {limit: -1}) {
            edge {
              nodes {
                name
                status
                organization {name}
                extensions(page: {limit: -1}) {
                  edge {
                    nodes {
                      name
                      state
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

  getOrgQuota (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        Organizations(page: {limit: -1}) {
          edge {
            nodes {
              name
              quota {
                used {
                  cpus
                  memoryMi
                  storageGi
                }
                total {
                  cpus
                  memoryMi
                  storageGi
                }
              }
            }
          }
        }
      }`,
      errorPolicy: 'all'
    });
  }
}
