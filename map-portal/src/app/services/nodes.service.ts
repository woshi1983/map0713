import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import { DrainNode, UpdateNode } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class NodesService {
  constructor (private apollo: Apollo) { }

  getAllNodes (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Nodes(filters: $filters) {
              name
              labels
              spec {
                unschedulable
                taints {
                  key
                  value
                  effect
                }
              }
              status {
                kernelVersion
                osImage
                internalIP
                externalIP
                ready
                capacity
              }
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

  getNodeResource (clusterId: string, nodeName: string): Observable<any> {
    const filters = [{ key: 'name', value: [nodeName] }, { key: 'cluster', value: [clusterId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Nodes(filters: $filters) {
            status {
              usage
              allocated {
                limits
                requests
              }
              allocatable
              capacity
            }
          }
      }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  updateNode (input: UpdateNode): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateNode!){
        UpdateNode(input: $input){
          name
        }
      }`,
      variables: {
        input
      },
      errorPolicy: 'all'
    });
  }

  listenDrainNode (input: DrainNode): Observable<any> {
    const sub = gql`subscription($input: DrainNode!){
      DrainNode(input: $input){
        evictedPods
        totalPods
        stdout
        stderr
        timestamp
      }
    }`
    return this.apollo.subscribe({
      query: sub,
      variables: { input },
      fetchPolicy: 'no-cache',
      errorPolicy: 'all'
    })
  }
}
