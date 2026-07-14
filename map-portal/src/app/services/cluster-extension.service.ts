import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { DisableClusterExtension, EnableClusterExtension, ModifyClusterExtension, CreateShortcut, UpdateShortcut, DeleteShortcut } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class ClusterExtensionService {
  constructor (private apollo: Apollo) { }

  getAllClusterExtension (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ClusterExtensions(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                endpoints
                state
                status
                spec
                related {edge { nodes {id name}}}
                cluster {id name}
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

  getAllCSIControllerExtension (num: number, limit: number, orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }, { key: 'name', value: ['CSI_DRIVER_NFS'] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ClusterExtensions(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                state
                status
                cluster { name }
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

  // used in cluster details octant link
  getClusterExtensionRelated (cluster: string, name: string): Observable<any> {
    const filters = [{ key: 'related', value: [cluster] }, { key: 'name', value: [name] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ClusterExtensions(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                endpoints
                state
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

  enableClusterExtensions (input: EnableClusterExtension): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: EnableClusterExtension!){
          EnableClusterExtension(input: $input){
            cluster {name}
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  modifyClusterExtensions (input: ModifyClusterExtension): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: ModifyClusterExtension!){
          ModifyClusterExtension(input: $input){
            cluster {name}
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  disableClusterExtensions (id: string): Observable<any> {
    const input: DisableClusterExtension = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DisableClusterExtension!){
          DisableClusterExtension(input: $input){
            cluster {name}
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  getAllShotcuts (orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Shortcuts(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                description
                endpoint
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

  createShortcut (input: CreateShortcut): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateShortcut!){
          CreateShortcut(input: $input){
            name
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  updateShortcut (input: UpdateShortcut): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateShortcut!){
          UpdateShortcut(input: $input){
            name
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteShortcut (id: string): Observable<any> {
    const input: DeleteShortcut = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteShortcut!){
          DeleteShortcut(input: $input){
            name
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }
}
