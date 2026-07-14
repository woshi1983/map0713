import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateLoadBalancer, DeleteLoadBalancer, LoadBalancerType, UpdateLoadBalancer } from '../classes';

@Injectable({
  providedIn: 'root'
})

export class LoadbalancerService {
  constructor (private apollo: Apollo) { }

  getAllLB (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          LoadBalancers(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                type
                endpoint
                caCert
                insecureSkipTLSVerify
                tlsCertChain {
                  notAfter
                  isCA
                }
                authUser
                spec {
                  __typename
                  ... on AviLoadBalancerSpec {
                      licenseTier
                      preferredNetwork
                      seGroup
                  }
                }
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

  getLbNames (orgId: string): Observable<any> {
    const filters = [{ key: 'organization', value: [orgId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          LoadBalancers(page: {limit: -1}, filters: $filters) {
            edge {
              nodes {
                id
                name
                type
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

  getLbNetworkprofile (lbId: string, protocol: string): Observable<any> {
    const filters = [{ key: 'protocol', value: [protocol] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          LoadBalancer(id: "${lbId}") {
            status {
              __typename
              ... on AviLoadBalancerStatus {
                networkProfiles(filters: $filters) {
                  name
                }
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

  validateCreateLB (input: CreateLoadBalancer): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateLoadBalancer!){
          CreateLoadBalancer(input: $input){
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

  validateUpdateLB (input: UpdateLoadBalancer): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateLoadBalancer!){
          UpdateLoadBalancer(input: $input){
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

  createLB (input: CreateLoadBalancer): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateLoadBalancer!){
          CreateLoadBalancer(input: $input){
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

  updateLB (input: UpdateLoadBalancer): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateLoadBalancer!){
          UpdateLoadBalancer(input: $input){
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

  deleteLB (id: string): Observable<any> {
    const input: DeleteLoadBalancer = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteLoadBalancer!){
          DeleteLoadBalancer(input: $input){
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
