# Map Console
<img alt="map" src=./src/assets/images/logo.png">    

## Introdution   
The application modernization platform is an application containerization platform based on Kubernetes. It provides a rich and out-of-the-box experience for cluster create/operation teams, application development/testing teams, and application release/operation teams. As a modern application platform, it provides better cluster management, application construction and release, multi-tenancy and other functions, improving the efficiency of users to build and run modern applications using Kubernetes.

This project is the frontend of mapd, For installation, please refer to []. 

## Features
#### Cluster management
* Fully automated delivery of Kubernetes clusters using ClusterAPI
* Integrated with vSphere, full Tanzu Kubernetes (TKGs and TKGm) cluster lifecycle management
* Support for attaching any existing Kubernetes cluster
* Kubernetes cluster backup and recovery using Velero
#### Multi-tenancy support
* Project management, Project isolation, and Project resource quota settings
* Permission management under project
#### Operation services
* supports one-click deployment and configuration of operation services, and currently supports Octant, monitoring, and logging.
#### App build
* Binary builds: Build the image from the binary, using kaniko that does not require root privileges
* Dockerfile builds: Build the image from the dockerfile, using kaniko that does not require root privileges
* source builds: Implement source code to image using TBS 
#### Application management*
\**** This feature is currently in the technical preview, and related features may change in future versions and are not enabled by default. For trial, please refer to [Technology Preview](./Technology-Preview.md) to enabled. 
* Support application lifecycle management
* Support deploy application to multiple clusters one time
* Expose application to external access based on L4/L7
* Multi-cluster traffic based on traffic percentage

## Documentation
Full documentation is available on the [Docs]().

## Contributing
Check out [CONTRIBUTING](./CONTRIBUTING.md) to see how to develop with map-console.
