## Java Proxy

This is a proxy that is used to forward requests to the k8s.

### Configuring the development environment

#### Configuration kubernetes environment
 
```shell script
gcloud container clusters get-credentials cluster-1 --region us-central1-c --project sn-demo-279007
```

#### Start service

Importing a project into the idea development environment.

Configuration environment variable `KUBECONFIG` for idea development environment.

#### Test Interface

```shell script
./bin/pulsar-admin --admin-url http://localhost:6750 functions status --tenant public --namespace default --name functionmesh-sample-ex1
```

or 

```shell script
curl http://localhost:6750/admin/v3/functions/test/default/functionmesh-sample-ex1/status
```


## More tools

### Automatic generation java [crd model](https://github.com/kubernetes-client/java/blob/master/docs/generate-model-from-third-party-resources.md)

crd yaml [file](https://github.com/streamnative/function-mesh/tree/master/config/crd/bases)

Note: add the field `preserveUnknownFields: false` to spec for avoid this [issue]()https://github.com/kubernetes-client/java/issues/1254

```shell script
LOCAL_MANIFEST_FILE=/Users/tuteng/streamnative/temp-data/cloud.streamnative.io_functions.yaml
docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$(pwd)":"$(pwd)" \
  -ti \
  --network host \
  docker.pkg.github.com/kubernetes-client/java/crd-model-gen:v1.0.3 \
  /generate.sh \
  -u $LOCAL_MANIFEST_FILE \
  -n io.streamnative.cloud \
  -p io.streamnative.cloud \
  -o "$(pwd)"
```
