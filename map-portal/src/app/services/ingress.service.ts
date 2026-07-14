import { Injectable } from '@angular/core';
import { Apollo, gql } from 'apollo-angular';
import { Observable } from 'rxjs';
import { CreateIngress, DeleteIngress, UpdateIngress } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class IngressService {
  constructor (private apollo: Apollo) { }

  getIngresses (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Ingresses(filters: $filters) {
            name
            namespace
            annotations
            labels
            spec {
              ingressClass
              tls {
                hosts
                secret
              }
              rules {
                host
                http {
                  paths {
                    path
                    pathType
                    backend {service {name port{ number}}}
                  }
                }
              }
            }
            component {name}
            application {name}
            status {
              ingress {
                ip
              }
            }
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

  getAllIngressClass (clusterId: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [clusterId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          IngressClasses(filters: $filters) {
            name
          }
        }`,
      variables: {
        filters: filters
      },
      errorPolicy: 'all'
    });
  }

  createIngress (input: CreateIngress): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateIngress!){
          CreateIngress(input: $input){
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

  updateIngress (input: UpdateIngress): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateIngress!){
          UpdateIngress(input: $input){
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

  deleteIngress (input: DeleteIngress): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteIngress!){
          DeleteIngress(input: $input){
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
