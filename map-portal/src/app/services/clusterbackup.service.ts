import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { DeleteBackup, CreateBackup } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class ClusterBackupService {
  constructor (private apollo: Apollo) { }

  getClusterBackups (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]){
          Backups(filters: $filters) {
            name
            status
            spec {
              useRestic
              includedNamespaces
              ttl
            }
            instances {name namespace }
            location {name spec {accessMode}}
            cluster {id name}
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

  // get backup namespace when resotre
  getBackupNamespaces (cluster: string, backup: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
           Backup(cluster: "${cluster}", name: "${backup}") {
             spec {includedNamespaces}
           }
        }`,
      errorPolicy: 'all'
    });
  }

  createClusterBackup (input: CreateBackup): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateBackup!){
        CreateBackup(input: $input){
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

  deleteClusterBackup (name: string, cluster: string): Observable<any> {
    const input: DeleteBackup = {
      name: name,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteBackup!){
          DeleteBackup(input: $input){
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
