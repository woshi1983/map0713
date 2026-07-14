import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateBackupStorageLocation, DeleteBackupStorageLocation, BackupStorageLocationAccessMode } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class BackupLocationService {
  constructor (private apollo: Apollo) { }

  validateMinio (input: CreateBackupStorageLocation): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateBackupStorageLocation!){
        CreateBackupStorageLocation(input: $input){
          name
        }
      }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  getBackupLocationsOfCluster (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          BackupStorageLocations(filters: $filters) {
            name
            status
            spec { provider accessMode config {region s3Url} objectStorage {bucket prefix} }
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

  getAvailableBackupLocatons (clusterId: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          BackupStorageLocations(filters: $filters) {
            name
            status
            spec {accessMode}
            cluster {
              namespaces {name}
            }
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  createBackupLocation (input: CreateBackupStorageLocation): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateBackupStorageLocation!){
        CreateBackupStorageLocation(input: $input){
          name
        }
      }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteBackupLocation (name: string, id: string): Observable<any> {
    const input: DeleteBackupStorageLocation = {
      name: name,
      cluster: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteBackupStorageLocation!){
        DeleteBackupStorageLocation(input: $input){
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
