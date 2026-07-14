import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import { CreateRoute, DeleteRoute, UpdateRoute } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class RouteService {
  constructor (private apollo: Apollo) { }

  getRoutes (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Routes(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                protocol
                backends {
                  id
                  name
                  ratio
                  ip
                  port
                  instance {
                   id
                   cluster {name}
                   namespace
                  }
                }
                spec {
                  __typename
                  ... on AviRouteSpec {
                    networkProfile
                    lbAlgorithm
                    lbAlgorithmHash
                    lbAlgorithmConsistentHashHeader
                    manualAllocatedVsVIP
                    persistentVsVIP
                  }
                }
                component {id name}
                application {id name}
                loadBalancer {id name}
                status
                ip
                port
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

  getRoute (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($id: ID!) {
          Route(id: $id) {
            name
            backends {
              ratio
              instance {
                name
                cluster{name}
                namespace
                status {
                  state
                }
                release {name num}
              }
              name
            }
          }
        }`,
      variables: {
        id
      },
      errorPolicy: 'all'
    });
  }

  createRoute (input: CreateRoute): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateRoute!){
          CreateRoute(input: $input){
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

  updateRoute (input: UpdateRoute): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateRoute!){
          UpdateRoute(input: $input){
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

  deleteRoute (id: string): Observable<any> {
    const input: DeleteRoute = {
      id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteRoute!){
          DeleteRoute(input: $input){
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

  forceDeleteRoute (id: string): Observable<any> {
    const input: DeleteRoute = {
      id: id,
      force: true
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteRoute!){
          DeleteRoute(input: $input){
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
