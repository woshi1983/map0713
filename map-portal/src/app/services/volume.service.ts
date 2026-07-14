import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import {
  CreateConfigMap,
  UpdateConfigMap,
  DeleteConfigMap,
  CreatePersistentVolumeClaim,
  DeletePersistentVolumeClaim,
  CreateSecret,
  DeleteSecret,
  UpdateSecret,
  UpdatePersistentVolumeClaim,
  CreateStorageClass,
  UpdateStorageClass,
  DeleteStorageClass
} from '../classes';

@Injectable({
  providedIn: 'root'
})
export class VolumeService {
  constructor (private apollo: Apollo) { }

  getAllConfigmaps (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ConfigMaps(filters: $filters) {
            name
            namespace
            labels
            data
            component {id name}
            application {id name}
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

  createConfigMap (param: CreateConfigMap): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateConfigMap!){
          CreateConfigMap(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  updateConfigMap (param: UpdateConfigMap): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateConfigMap!){
          UpdateConfigMap(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  deleteConfigMap (name: string, cluster: string, namespace: string): Observable<any> {
    const input: DeleteConfigMap = {
      name: name,
      cluster: cluster,
      namespace: namespace
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteConfigMap!){
          DeleteConfigMap(input: $input){
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

  getAllPVCs (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          PersistentVolumeClaims(filters: $filters) {
            name
            labels
            component {id name}
            application {id name}
            cluster {id name}
            namespace
            status
            spec {
              accessModes
              resources {
                requests {
                  storage
                }
              }
              storageClass
            }
            component {name}
            application {name}
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

  createPVC (param: CreatePersistentVolumeClaim): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreatePersistentVolumeClaim!){
          CreatePersistentVolumeClaim(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  updatePVC (param: UpdatePersistentVolumeClaim): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdatePersistentVolumeClaim!){
          UpdatePersistentVolumeClaim(input: $input){
            name
            createdAt
            lastModified
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  deletePVC (name: string, cluster: string, namespace: string): Observable<any> {
    const input: DeletePersistentVolumeClaim = {
      name: name,
      cluster: cluster,
      namespace: namespace
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeletePersistentVolumeClaim!){
          DeletePersistentVolumeClaim(input: $input){
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

  getAllSecrets (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Secrets(filters: $filters) {
            name
            labels
            component {id name}
            application {id name}
            cluster {id name}
            namespace
            data
            type
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

  // used by ingress https
  getTLSSecretsByNamespace (clusterId: string, namespace: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }, { key: 'namespace', value: [namespace] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Secrets(filters: $filters) {
            name
            type
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  createSecret (param: CreateSecret): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateSecret!){
          CreateSecret(input: $input){
             name
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  updateSecret (param: UpdateSecret): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateSecret!){
          UpdateSecret(input: $input){
             name
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  deleteSecret (name: string, cluster: string, namespace: string): Observable<any> {
    const input: DeleteSecret = {
      name: name,
      cluster: cluster,
      namespace: namespace
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteSecret!){
          DeleteSecret(input: $input){
             name
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  getAllStorageClasses (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          StorageClasses(filters: $filters) {
            name
            labels
            annotations
            provisioner
            mountOptions
            reclaimPolicy
            volumeBindingMode
            allowVolumeExpansion
            cluster {id name}
            component {id name}
            application { id name}
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

  getAllStorageClassesByCluster (clusterId: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          StorageClasses(filters: $filters) {
            name
            allowVolumeExpansion
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  createStorageClass (param: CreateStorageClass): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateStorageClass!){
          CreateStorageClass(input: $input){
            name
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  updateStorageClass (param: UpdateStorageClass): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateStorageClass!){
          UpdateStorageClass(input: $input){
            name
          }
        }`,
      variables: {
        input: param
      },
      errorPolicy: 'all'
    });
  }

  deleteStorageClass (name: string, cluster: string): Observable<any> {
    const input: DeleteStorageClass = {
      name: name,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteStorageClass!){
          DeleteStorageClass(input: $input){
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
