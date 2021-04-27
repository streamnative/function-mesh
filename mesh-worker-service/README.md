## Mesh Worker Service

This is a proxy that is used to forward requests to the k8s.

### Run with pulsar broker

#### Prerequisite

* Pulsar 2.8.0 or later

#### Update configuration file

Add the following configuration to the `functions_worker.yml` configuration file:

```$xslt
functionsWorkerServiceNarPackage: /YOUR-NAR-PATH/mesh-worker-service-1.0-SNAPSHOT.nar
```
Replace the `YOUR-NAR-PATH` variable with your real path.


### Configuring the development environment


#### Start service

Importing this project into the idea development environment.

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
   --user-config "{"clusterName": "test-pulsar", "typeClassName": "java.lang.String"}"
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
   --user-config "{"clusterName": "test-pulsar", "typeClassName": "java.lang.String"}"
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

Note: add the field `preserveUnknownFields: false` to spec for avoid this [issue](https://github.com/kubernetes-client/java/issues/1254)

```shell script
CRD_FILE=compute.functionmesh.io_sources.yaml # Target CRD file

GEN_DIR=/tmp/functions-mesh/crd
mkdir -p $GEN_DIR
cp ../config/crd/bases/* $GEN_DIR
cd $GEN_DIR

LOCAL_MANIFEST_FILE=$GEN_DIR/$CRD_FILE

# yq site: https://mikefarah.gitbook.io/yq/
yq e ".spec.preserveUnknownFields = false" -i $CRD_FILE 

docker rm -f kind-control-plane
docker run \
  --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "$(pwd)":"$(pwd)" \
  -ti \
  --network host \
  docker.pkg.github.com/kubernetes-client/java/crd-model-gen:v1.0.3 \
  /generate.sh \
  -u $LOCAL_MANIFEST_FILE \
  -n io.functionmesh.compute \
  -p io.functionmesh.compute \
  -o "$(pwd)"

open $GEN_DIR
```

### A auto tool for generated crd

```shell script
./tool/generate-crd.sh
```

Then add license for crd model file
```shell script
mvn license:format
```
