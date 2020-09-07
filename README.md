# kuwait - Wait with Style

`kuwait` is short for `kubewait`, which is short for `kubernetes wait helper program utility`,
and is meant to be a small, handy command-line application to make waiting for certain
conditions in Kubernetes clusters easier and more efficient.

```
./kuwait deployment/cert-manager/cert-manager
INFO[Tue, 08 Sep 2020 00:42:09 CEST] Waiting 10m0s for the following conditions to be met:
INFO[Tue, 08 Sep 2020 00:42:09 CEST] The Deployment "cert-manager/cert-manager" must be ready.
INFO[Tue, 08 Sep 2020 00:42:10 CEST] All conditions are met.
```

`kuwait` has the following design goals:

* make it easy to use
* allow to wait for arbitrary resources
* do busy waiting (i.e. polling) *without* backoff by default, because backoffs are too
  often used "just because that's what you do" and waste tons of times in controlled
  environments like CI cluster

## Operation

`kuwait` parses each argument as a "wait condition". Each condition is a string consisting
of multiple parts, separated by slashes (e.g. `kuwait foo/bar/bla another/condition and/another/condition`).

A condition always begins with the Kind of the resource to wait for. This can be `pod`,
`deployment`, `clusterissuer` (cert-manager) or any other Kind.

Depending on the Kind, the condition is then followed by namespace and name or just a name
of a resource to wait for. Namespaced resources must always specify a concrete namespace,
but for the name `*` is allowed as a wildcard.

After the identifier, the condition is given. The default condition is "ready" and does what
you would expect: Check if the found resource has the "Ready" condition.

A few examples:

* `pod/cert-manager/cert-manager-378436-2746/ready`
* `clusterrole/foobar/exist` (here the condition is "ready", because ClusterRoles are non-namespaced)
* `pvc/minio/minio-data/gone`

After parsing all conditions, one goroutine is started per condition, which actively polls
the cluster until the condition is met or the global timeout interrupts it.

## Installation

You need Go 1.14 installed on your machine.

```
go get go.xrstf.de/kuwait
```

A Docker image is available as [`xrstf/kuwait`](https://hub.docker.com/r/xrstf/kuwait).

## Usage

```
Usage of kuwait:
  -debug
        enable more verbose logging
  -interval duration
        time inbetween status checks (default 1s)
  -kubeconfig string
        kubeconfig file to use
  -timeout duration
        maximum time to wait for all conditions to be met (default 10m0s)
```

The kubeconfig can also be given using the `KUBECONFIG` environment variable.

## License

MIT
