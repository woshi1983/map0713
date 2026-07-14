import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import { CreateUser, UpdateUser } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class UsersService {
  constructor (
    private apollo: Apollo
  ) { }

  getAllUsers (num: number, limit: number, filters: any[]): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Users(page: {number: ${num}, limit: ${limit}}, filters: $filters){
            totalCount
            edge {
              nodes {
                id
                name
                email
                displayName
                admin
                managed
                roleBindings(page: {limit: -1}) {
                  edge {
                    nodes {
                      role
                      organization {id name}
                    }
                  }
                }
                createdAt
                lastModified
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

  getAllUserNames (): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          Users(page: {limit: -1}) {
            edge {
              nodes {
                id
                name
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getUserInfoById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          User(id: "${id}"){
            id
            name
            displayName
            admin
            managed
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getUserInfoAndRoleById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          User(id: "${id}"){
            id
            name
            displayName
            admin
            managed
            roleBindings(page: {limit: -1}) {
              edge {
                nodes {
                  role
                  organization {id}
                }
              }
            }
          }
        }`,
      errorPolicy: 'all'
    });
  }

  createUser (input: CreateUser): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: CreateUser!){
          CreateUser(input: $input){
            id
            name
            password
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    })
  }

  resetUserPasswd (id: string, newpasswd: string): Observable<any> {
    const input: UpdateUser = {
      id: id,
      password: newpasswd
    };
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: UpdateUser!){
          UpdateUser(input: $input){
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

  updateUserPasswd (id: string, oldPasswd: string, newpasswd: string): Observable<any> {
    const input: UpdateUser = {
      id: id,
      oldPassword: oldPasswd,
      password: newpasswd
    };
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: UpdateUser!){
          UpdateUser(input: $input){
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

  updateUserAdmin (id: string, admin: boolean): Observable<any> {
    const input: UpdateUser = {
      id: id,
      admin: admin
    };
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: UpdateUser!){
          UpdateUser(input: $input){
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

  updateUserInfo (input: UpdateUser): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`
      mutation($input: UpdateUser!){
          UpdateUser(input: $input){
            id
            name
            displayName
          }
        }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  deleteUser (id: string): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`
        mutation($input: DeleteUser!) {
          DeleteUser(input: $input) {
            id
            name
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
