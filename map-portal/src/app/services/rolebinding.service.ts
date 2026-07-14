import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import { CreateRoleBinding, UpdateRoleBinding } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class RolebindingService {
  constructor (
    private apollo: Apollo
  ) { }

  getAllRolebindings (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          RoleBindings(page: {number: ${num}, limit: ${limit}}, filters: $filters){
            totalCount
            edge {
              nodes {
                id
                user { id name }
                role
              }
            }
          }
        }`,
      errorPolicy: 'all',
      variables: {
        filters: filters
      }
    });
  }

  createRolebinding (input: CreateRoleBinding): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: CreateRoleBinding!){
          CreateRoleBinding(input: $input){
            id
            user {name}
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    })
  }

  updateRolebinding (input: UpdateRoleBinding): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: UpdateRoleBinding!){
          UpdateRoleBinding(input: $input){
            id
            user {name}
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    })
  }

  deleteRolebinding (id: string): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`
        mutation($input: DeleteRoleBinding!) {
          DeleteRoleBinding(input: $input) {
            id
            user {name}
          }
        }`,
      variables: {
        input: {
          id: id
        }
      },
      errorPolicy: 'all'
    });
  }
}
