import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateNamespace, UpdateNamespace, DeleteNamespace } from 'app/classes';

@Injectable({
  providedIn: 'root'
})
export class NamespaceService {
  constructor (private apollo: Apollo) { }

  getNamespaces (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Namespaces(filters: $filters) {
              name
              labels
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

  createNamespace (createNsParam: CreateNamespace): Observable<any> {
    const input: CreateNamespace = createNsParam;
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateNamespace!){
        CreateNamespace(input: $input){
          name
          labels
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

  updateNamespaceLabels (updateNsParam: UpdateNamespace): Observable<any> {
    const input: UpdateNamespace = updateNsParam;
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateNamespace!){
        UpdateNamespace(input: $input){
          name
          labels
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

  deleteNamespace (name: string, cluster: string): Observable<any> {
    const input: DeleteNamespace = {
      name: name,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteNamespace!){
        DeleteNamespace(input: $input){
          name
          labels
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
