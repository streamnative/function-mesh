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

#### getFunctionStatus
```shell script
./bin/pulsar-admin --admin-url http://localhost:6750 functions status --tenant public --namespace default --name functionmesh-sample-ex1
```

or 

```shell script
curl http://localhost:6750/admin/v3/functions/test/default/functionmesh-sample-ex1/status
```

#### registerFunction
 ```shell script
./bin/pulsar-admin --admin-url http://localhost:6750 functions create \
   --jar target/my-jar-with-dependencies.jar \
   --classname org.example.functions.WordCountFunction \
   --tenant public \
   --namespace default \
   --name word-count \
   --inputs persistent://public/default/sentences \
   --output persistent://public/default/count \
   --input-specs "{"source": {"serdeClassName": "java.lang.String"}}" \
   --output-serde-classname java.lang.String \
   --cpu 0.1 \
   --ram 1 \
   --user-config "{"clusterName": "test-pulsar"}"
```

#### updateFunction
 ```shell script
./bin/pulsar-admin --admin-url http://localhost:6750 functions update \
   --jar target/my-jar-with-dependencies.jar \
   --classname org.example.functions.WordCountFunction \
   --tenant public \
   --namespace default \
   --name word-count \
   --inputs persistent://public/default/sentences \
   --output persistent://public/default/count \
   --input-specs "{"source": {"serdeClassName": "java.lang.String"}}" \
   --output-serde-classname java.lang.String \
   --cpu 0.2 \
   --ram 1 \
   --user-config "{"clusterName": "test-pulsar"}"
```

#### getFunctionInfo
 ```shell script
./bin/pulsar-admin --admin-url http://localhost:6750 functions get \
   --tenant public \
   --namespace default \
   --name word-count
```

#### deregisterFunction
 ```shell script
./bin/pulsar-admin --admin-url http://localhost:6750 functions delete \
   --tenant public \
   --namespace default \
   --name word-count
```

## More tools

### Automatic generation java [crd model](https://github.com/kubernetes-client/java/blob/master/docs/generate-model-from-third-party-resources.md)

crd yaml [file](https://github.com/streamnative/function-mesh/tree/master/config/crd/bases)

Note: add the field `preserveUnknownFields: false` to spec for avoid this [issue]()https://github.com/kubernetes-client/java/issues/1254

```shell script
tmpDir=/tmp/functions-mesh/crd
mkdir -p $tmpDir
cp ../config/crd/bases $tmpDir

LOCAL_MANIFEST_FILE=${tmpDir}/cloud.streamnative.io_sources.yaml

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
