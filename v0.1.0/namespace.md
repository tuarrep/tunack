# Running in a custom namespace

In tunack v0.1.0 you can only change where tunack is deloyed.
Tunack is staticaly built to search for Nginx config map in `ingress-nginx` namespace.

## Deploying to a custom namespace

To deploy tunack to a custom namespace you need to patch the [sample deployment file][1].
Change the namespace by yours at lines 5, 36, 44, 63 and 71.
You also can safely replace all `ingress-nginx` occurences by your custom namespace name in this file.

## Run with Nginx Ingres controller in another namespace
In v0.1.0 you need to edit source code.
You need to patch [sync.go:19][2] and replace `ingresss-nginx` by your custom namespace name.
Example:
```go
configMapClient := client.CoreV1().ConfigMaps("my-custom-namespace")
```

[1]: https://github.com/dahus/tunack/blob/7c1a57fe152c10aed2fc5d03e0df4d14437081ce/deploy/with-rbac.yaml
[2]: https://github.com/dahus/tunack/blob/7c1a57fe152c10aed2fc5d03e0df4d14437081ce/sync.go#L19