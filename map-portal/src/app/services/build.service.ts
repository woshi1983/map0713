import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Apollo, gql } from 'apollo-angular';
import {
  CreateBinaryImage,
  UpdateBinaryImage,
  DeleteBinaryImage,
  CreateBinaryBuild,
  CreateDockerfileImage,
  CreateDockerfileBuild,
  DeleteDockerfileImage,
  UpdateDockerfileImage,
  DeleteSourceImage,
  CreateSourceBuild,
  CreateSourceImage,
  UpdateSourceImage,
  DeleteBinaryBuild,
  DeleteDockerfileBuild,
  DeleteSourceBuild
} from '../classes/';

@Injectable({
  providedIn: 'root'
})
export class BuildService {
  constructor (private apollo: Apollo) { }

  getBinaryImages (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          BinaryImages(filters: $filters) {
            name
            annotations
            spec {baseImage distImage binaryDir url processUID}
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

  getBianryImageDetails (cluster: string, name: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          BinaryImage(cluster: "${cluster}", name: "${name}") {
              name
              annotations
              spec {
                baseImage
                distImage
                binaryDir
                url
                processUID
              }
              cluster {id name}
              createdAt
              lastModified
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getDockerfileImages (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          DockerfileImages(filters: $filters) {
            name
            annotations
            spec {source distImage branch url subDir fileName}
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

  getDockerfileImageDetails (cluster: string, name: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          DockerfileImage(cluster: "${cluster}", name: "${name}") {
              name
              annotations
              spec {
                source
                distImage
                branch
                subDir
                fileName
                url
              }
              cluster {id name}
              createdAt
              lastModified
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getSourceImages (filters: any): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          SourceImages(filters: $filters ) {
            name
            annotations
            spec {
              source {
                git {
                  revision
                  url
                }
                blob {
                  url
                }
                registry {
                  image
                }
                subPath
              }
              distImage
              env
            }
            status {
              ready
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

  getSourceImageDetails (cluster: string, name: string): Observable<any> {
    return this.apollo.query<any>({
      query: gql`
        query {
          SourceImage(cluster: "${cluster}", name: "${name}") {
              name
              annotations
              spec {
                source {
                  git {
                    revision
                    url
                  }
                  blob {
                    url
                  }
                  registry {
                    image
                  }
                  subPath
                }
                distImage
                env
              }
              status {
                ready
              }
              cluster {id name}
              createdAt
              lastModified
          }
        }`,
      errorPolicy: 'all'
    });
  }

  getBinaryBuildsByImage (image: string, cluster: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [cluster] }, { key: 'image', value: [image] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          BinaryBuilds(filters: $filters ) {
            name
            spec {url}
            status {
              num
              status
              image
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

  getDockerfileBuildsByImage (image: string, cluster: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [cluster] }, { key: 'image', value: [image] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          DockerfileBuilds(filters: $filters ) {
            name
            spec {url}
            status {
              num
              status
              image
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

  getSourceBuildsByImage (image: string, cluster: string): Observable<any> {
    const filters = [{ key: 'cluster', value: [cluster] }, { key: 'image', value: [image] }];
    return this.apollo.query<any>({
      query: gql`
        query($filters: [QueryFilter!]) {
          SourceBuilds(filters: $filters ) {
            name
            spec {
              source {
                git {
                  revision
                  url
                }
                blob {
                  url
                }
                registry {
                  image
                }
              }
            }
            status {
              num
              status
              image
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

  createDockerfileImage (input: CreateDockerfileImage): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateDockerfileImage!){
        CreateDockerfileImage(input: $input){
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

  createBinaryImage (input: CreateBinaryImage): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateBinaryImage!){
        CreateBinaryImage(input: $input){
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

  createSourceImage (sourceParam: CreateSourceImage): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateSourceImage!){
        CreateSourceImage(input: $input){
          name
          createdAt
          lastModified
        }
      }`,
      variables: {
        input: sourceParam
      },
      errorPolicy: 'all'
    });
  }

  createBinaryBuild (image: string, cluster: string): Observable<any> {
    const input: CreateBinaryBuild = {
      image: image,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateBinaryBuild!){
        CreateBinaryBuild(input: $input){
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

  createDockerfileBuild (image: string, cluster: string): Observable<any> {
    const input: CreateDockerfileBuild = {
      image: image,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateDockerfileBuild!){
        CreateDockerfileBuild(input: $input){
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

  createSourceBuild (image: string, cluster: string): Observable<any> {
    const input: CreateSourceBuild = {
      image: image,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: CreateSourceBuild!){
        CreateSourceBuild(input: $input){
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

  updateBinaryImage (input: UpdateBinaryImage): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateBinaryImage!){
        UpdateBinaryImage(input: $input){
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

  updateDockerfileImage (input: UpdateDockerfileImage): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateDockerfileImage!){
        UpdateDockerfileImage(input: $input){
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

  updateSourceImage (sourceParam: UpdateSourceImage): Observable<any> {
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: UpdateSourceImage!){
        UpdateSourceImage(input: $input){
          name
          createdAt
          lastModified
        }
      }`,
      variables: {
        input: sourceParam
      },
      errorPolicy: 'all'
    });
  }

  deleteBinaryImage (name: string, clusterId: string): Observable<any> {
    const input: DeleteBinaryImage = {
      name: name,
      cluster: clusterId
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteBinaryImage!){
        DeleteBinaryImage(input: $input){
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

  deleteDockerfileImage (name: string, clusterId: string): Observable<any> {
    const input: DeleteDockerfileImage = {
      name: name,
      cluster: clusterId
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteDockerfileImage!){
        DeleteDockerfileImage(input: $input){
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

  deleteSourceImage (name: string, clusterId: string): Observable<any> {
    const input: DeleteSourceImage = {
      name: name,
      cluster: clusterId
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteSourceImage!){
        DeleteSourceImage(input: $input){
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

  deleteBinaryBuild (name: string, cluster: string): Observable<any> {
    const input: DeleteBinaryBuild = {
      name: name,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteBinaryBuild!){
        DeleteBinaryBuild(input: $input){
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

  deleteDockerfileBuild (name: string, cluster: string): Observable<any> {
    const input: DeleteDockerfileBuild = {
      name: name,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteDockerfileBuild!){
        DeleteDockerfileBuild(input: $input){
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

  deleteSourceBuild (name: string, cluster: string): Observable<any> {
    const input: DeleteSourceBuild = {
      name: name,
      cluster: cluster
    };
    return this.apollo.mutate<any>({
      mutation: gql`mutation($input: DeleteSourceBuild!){
        DeleteSourceBuild(input: $input){
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
