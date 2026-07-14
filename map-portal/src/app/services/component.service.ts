import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import { CreateComponent, DeleteComponent, UpdateComponent } from '../classes';

@Injectable({
  providedIn: 'root'
})
export class ComponentService {
  constructor (private apollo: Apollo) {}

  getAllComponents (num: number, limit: number, filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Components(page: {number: ${num}, limit: ${limit}}, filters: $filters) {
            totalCount
            edge {
              nodes {
                id
                name
                description
                type
                application {id name}
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

  getAllComponentsByApplication (appId: string): Observable<any> {
    const filters = [{ key: 'application', value: [appId] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          Components(page: {limit: -1}, filters: $filters) {
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

  getComponentDesc (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        Component(id: "${id}") {
            description
            type
        }
      }`,
      errorPolicy: 'all'
    });
  }

  getComponentById (id: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
      query {
        Component(id: "${id}") {
            name
            description
            type
            template {
              __typename
              ... on RAWComponentTemplate {
                raw
              }
              ... on HelmComponentTemplate {
                chart {
                  name
                  files
                }
                url
                values
                extraInfo
              }
              ... on BasicComponentTemplate {
                deployments {
                  labels
                  spec {
                    progressDeadlineSeconds
                    replicas
                    revisionHistoryLimit
                    template {
                      labels
                      spec {
                        initContainers {
                          name
                          image
                          command
                          args
                          workingDir
                          ports {name containerPort protocol}
                          env
                          resources {
                            limits {
                              cpu
                              memory
                            }
                            requests {
                              cpu
                              memory
                            }
                          }
                          volumeMounts {name mountPath subPath}
                          imagePullPolicy
                        }
                        containers {
                          name
                          image
                          command
                          args
                          workingDir
                          ports {name containerPort protocol}
                          env
                          resources {
                            limits {
                              cpu
                              memory
                            }
                            requests {
                              cpu
                              memory
                            }
                          }
                          volumeMounts {name mountPath subPath}
                          livenessProbe {
                            exec { command }
                            httpGet {path port host scheme httpHeaders {name value}}
                            tcpSocket {port host}
                            initialDelaySeconds
                            timeoutSeconds
                            periodSeconds
                            successThreshold
                            failureThreshold
                          }
                          readinessProbe {
                            exec { command }
                            httpGet {path port host scheme httpHeaders {name value}}
                            tcpSocket {port host}
                            initialDelaySeconds
                            timeoutSeconds
                            periodSeconds
                            successThreshold
                            failureThreshold
                          }
                          startupProbe {
                            exec { command }
                            httpGet {path port host scheme httpHeaders {name value}}
                            tcpSocket {port host}
                            initialDelaySeconds
                            timeoutSeconds
                            periodSeconds
                            successThreshold
                            failureThreshold
                          }
                          lifecycle {
                            postStart {
                              exec { command }
                              httpGet { path port host scheme httpHeaders { name value} }
                              tcpSocket { port host}
                            }
                            preStop {
                              exec { command }
                              httpGet { path port host scheme httpHeaders { name value} }
                              tcpSocket { port host}
                            }
                          }
                          imagePullPolicy
                        }
                        dnsPolicy
                        restartPolicy
                        schedulerName
                        serviceAccount
                        terminationGracePeriodSeconds
                        volumes {
                          name
                          volume {
                            __typename
                            ... on EmptyDir {
                              medium
                              size
                            }
                            ... on ConfigMapVolume {
                              name
                              items {key path mode}
                              defaultMode
                            }
                            ... on SecretVolume {
                              name
                              items {key path mode}
                              defaultMode
                            }
                            ... on PersistentVolumeClaimVolume {
                              name
                              readOnly
                            }
                          }
                        }
                        imagePullSecrets
                      }
                    }
                    strategy {
                      type
                      rollingUpdate {
                        maxUnavailable
                        maxSurge
                      }
                    }
                  }
                  hpa {
                    maxReplicas
                    minReplicas
                    targetCPUUtilizationPercentage
                    targetMemoryUtilizationPercentage
                  }
                }
                services {
                  name
                  labels
                  spec {
                    type
                    ports {port protocol targetPort}
                  }
                }
                secrets {
                  __typename
                  ... on BasicComponentTemplateGenericSecret {
                    name
                    data
                    type
                  }
                }
                configMaps {
                  name
                  data
                }
                persistentVolumeClaims {
                  name
                  spec {
                    accessModes
                    storageClass
                    volumeMode
                    resources {
                      requests {
                        storage
                      }
                      limits {
                        storage
                      }
                    }
                  }
                }
              }
            }
            application {id name}
            createdAt
            lastModified
        }
      }`,
      errorPolicy: 'all'
    });
  }

  createComponent (param: CreateComponent): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateComponent!){
          CreateComponent(input: $input){
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

  updateComponent (param: UpdateComponent): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateComponent!){
          UpdateComponent(input: $input){
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

  deleteComponent (id: string): Observable<any> {
    const input: DeleteComponent = {
      id: id
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteComponent!){
          DeleteComponent(input: $input){
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
