# Tunack
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fdahus%2Ftunack.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fdahus%2Ftunack?ref=badge_shield)


:warning: Tunack is under heavy development, __NOT SUITABLE FOR PRODUCTION__ (or at your own risks). PR are warmly welcomed

Tcp and Udp Nginx Auto Config in Kubernetes

Auto configuration service for TCP and UDP services for Kubernetes Nginx ingress manager

## Getting Started

Tunack is made to be deployed in your cluster. YAML files are provided in [deploy folder](./deploy).

### Prerequisites

 You need to have Nginx ingress manager fully functionnal ([doc](https://github.com/kubernetes/ingress-nginx/blob/master/deploy/README.md))

### Installing


```bash
kubectl apply -f https://raw.githubusercontent.com/mafzst/tunack/v0.1.0/deploy/with-rbac.yaml
```

It will create a deployement in `ingress-nginx` namespace

## Usage

Tunack uses annotations in services to detect ports to expose.

### Synopsis

`tunack.dahus.io/[protocol]-service-[proxyPort]: [servicePort]`

- __protocol__: `tcp|udp`: Specify service type
- __proxyPort__: Port to export to outside (ie. public port)
- __servicePort__: Target service port (ie. port on which your service is listening to)

__ONLY TCP IS IMPLEMENTED YET__

### Example

``` yaml
apiVersion: v1
kind: Service
metadata:
  name: my-tcp-service
  annotations:
    - tunack.dahus.io/tcp-service-3000: 80
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: my-tcp-service

```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/mafzst/tunack/tags).

## Authors

* **Nicolas Perraut** - *Initial work* - [Dahus](https://dahus.net)

See also the list of [contributors](https://github.com/mafzst/tunack/contributors) who participated in this project.

## License

This project is licensed under the GNU GPLv3 License - see the [LICENSE](LICENSE) file for details


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fdahus%2Ftunack.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fdahus%2Ftunack?ref=badge_large)

## Acknowledgments

* Kubernetes client-go [contributors](https://github.com/kubernetes/client-go/graphs/contributors)