# run.tanzu

This package is duplicated from [guest-cluster-controller](https://gitlab.eng.vmware.com/core-build/guest-cluster-controller/-/tree/master/apis/run.tanzu).

## Why it exist?

The `guest-cluster-controller` uses `kubebuilder` for code generation, but `kubebuilder` does not support client generator. Besides, `guest-cluster-controller` has the following dependency
resolve issues:

- k8s-related library's version does not match, as mentioned in its `go.mod`
- `k8s.io/client-go`'s version is always incorrectly resolved to `v1.5.2` while using `sigs.k8s.io/cluster-api` as a dependency of `guest-cluster-controller`

Thus, we copy the most critical part of source code to here, and generate client with []