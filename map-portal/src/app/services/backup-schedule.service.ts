import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateClusterBackupSchedule, DeleteClusterBackupSchedule, UpdateClusterBackupSchedule } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class BackupScheduleService {
  constructor (private apollo: Apollo) { }

  getClusterBackupSchedule (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]){
          ClusterBackupSchedules(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                spec {
                  includedNamespaces
                  useRestic
                  ttl
                }
                location
                schedule
                lastFire
                nextFire
                components {edge {nodes {name}}}
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

  createClusterBackupSchedule (input: CreateClusterBackupSchedule): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateClusterBackupSchedule!){
        CreateClusterBackupSchedule(input: $input){
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

  updateClusterBackupSchedule (input: UpdateClusterBackupSchedule): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateClusterBackupSchedule!){
        UpdateClusterBackupSchedule(input: $input){
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

  deleteClusterBackupSchedule (id: string): Observable<any> {
    const input: DeleteClusterBackupSchedule = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteClusterBackupSchedule!){
          DeleteClusterBackupSchedule(input: $input){
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
