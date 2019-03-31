# Polaris Operator

A [Kubernetes operator](https://coreos.com/operators/) created via the [operator SDK](https://github.com/operator-framework/operator-sdk). It monitors polaris custom resources and takes various actions, for example creating a container registry on AWS on the creation of a polaris container registry custom resource in the cluster.

References:
- https://itnext.io/building-an-operator-for-kubernetes-with-operator-sdk-40a029ea056
- https://devops.college/developing-kubernetes-operator-is-now-easy-with-operator-framework-d3194a7428ff