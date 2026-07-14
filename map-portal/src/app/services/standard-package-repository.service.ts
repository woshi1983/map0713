import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import {
  CreateStandardPackageRepository,
  DeleteStandardPackageRepository,
  SetDefaultStandardPackageRepository,
  UpdateStandardPackageRepository
} from '../classes';

@Injectable({
  providedIn: 'root'
})
export class StandardPackageRepositoryService {
  constructor (private apollo: Apollo) { }

  getStandardPackageRepositories (num: number, limit: number, filters: {key: string; value: string[]}[]): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          StandardPackageRepositories(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                image
                default
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

  validateCreateStandardPackageRepository (input: CreateStandardPackageRepository): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateStandardPackageRepository!){
        CreateStandardPackageRepository(input: $input){
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

  validateUpdateStandardPackageRepository (input: UpdateStandardPackageRepository): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateStandardPackageRepository!){
        UpdateStandardPackageRepository(input: $input){
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

  createStandardPackageRepository (input: CreateStandardPackageRepository): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateStandardPackageRepository!){
        CreateStandardPackageRepository(input: $input){
          id
          name
          image
          default
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

  updateStandardPackageRepository (input: UpdateStandardPackageRepository): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateStandardPackageRepository!){
        UpdateStandardPackageRepository(input: $input){
          id
          name
          image
          default
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

  setDefaultStandardPackageRepository (id: string): Observable<any> {
    const input: SetDefaultStandardPackageRepository = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: SetDefaultStandardPackageRepository!){
        SetDefaultStandardPackageRepository(input: $input){
          id
          name
          image
          default
        }
      }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteStandardPackageRepository (id: string): Observable<any> {
    const input: DeleteStandardPackageRepository = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteStandardPackageRepository!){
        DeleteStandardPackageRepository(input: $input){
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
}
