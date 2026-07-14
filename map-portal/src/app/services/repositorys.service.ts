import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import {
  CreateImageRegistry,
  UpdateImageRegistry,
  DeleteImageRegistry
} from '../classes/';

@Injectable({
  providedIn: 'root'
})
export class RepositorysService {
  constructor (private apollo: Apollo) { }

  getImageRegistrys (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ImageRegistries(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                endpoint
                caCert
                tlsCertChain {
                  notAfter
                  isCA
                }
                authUser
                type
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

  getRegistryProject (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          ImageRegistry(id: "${id}") {
            endpoint
            projects
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getAllImagesByProject (id: string, project: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          ImageRepos(imageRegistry: "${id}", project: "${project}") {
            name
            tags
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getImageRegistrysNames (orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ImageRegistries(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
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

  deleteImageRegistry (id: string): Observable<any> {
    const input: DeleteImageRegistry = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteImageRegistry!){
          DeleteImageRegistry(input: $input){
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

  validateCreateImageRegistry (input: CreateImageRegistry): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateImageRegistry!){
          CreateImageRegistry(input: $input){
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

  validateUpdateImageRegistry (input: UpdateImageRegistry): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateImageRegistry!){
          UpdateImageRegistry(input: $input){
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

  createImageRegistry (input: CreateImageRegistry): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateImageRegistry!){
          CreateImageRegistry(input: $input){
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

  updateImageRegistry (input: UpdateImageRegistry): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateImageRegistry!){
          UpdateImageRegistry(input: $input){
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

  // used for catalogs
  getAllRegistrys (orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ImageRegistries(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                projects
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

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  getAllChartRepos (registryId: string, project: string, filters?: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ChartRepos(imageRegistry: "${registryId}", project: "${project}", filters: $filters) {
            name
            latestVersion
            avatar
            deprecated
            versions {
              _repo
              _project
              _imageRegistry
              version
            }
          }
        }`,
      errorPolicy: 'all',
      variables: {
        filters: filters
      }
    });
  }

  getChartDetails (registryId: string, project: string, name: string, version: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          ChartDetails(imageRegistry: "${registryId}", project: "${project}", name: "${name}", version: "${version}") {
            name
            description
            avatar
            sources
            homepage
            download
            keywords
            dependencies {
              name
              repository
              version
            }
            files
            values
          }
        }`,
      errorPolicy: 'all'
    });
  }

  // used for component update
  getAllChartRepoVersion (registryId: string, project: string, name: string): Observable<any> {
    const filters = [{ key: 'name', value: [name] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          ChartRepos(imageRegistry: "${registryId}", project: "${project}", filters: $filters) {
            versions {
              version
            }
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  getChartDetailValues (registryId: string, project: string, name: string, version: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          ChartDetails(imageRegistry: "${registryId}", project: "${project}", name: "${name}", version: "${version}") {
            download
            files
          }
        }`,
      errorPolicy: 'all'
    });
  }
}
