# Migration Tools

This tool is mainly used to migrate functions from function worker of pulsar cluster to `function-mesh`.

## How to use

Add configuration file `pulsarctl.properties` for this tool

```
webServiceUrl=http://localhost:8080
tlsAllowInsecureConnection=false
tlsTrustCertsFilePath=
brokerServiceUrl=
authParams=
authPlugin=
tlsEnableHostnameVerification=false
```

Declare environment variables `PULSAR_CLIENT_CONF`

```
export PULSAR_CLIENT_CONF=/PATH/pulsarctl.properties
```

Replace the `PATH` variable with the absolute path to the configuration file.

### build from the source code

```
git clone https://github.com/streamnative/function-mesh
cd function-mesh/tools
go build
```

### Generate function mesh configuration file

```
./tools
```

```
functions
└── public
    └── default
        └── test-func.yaml
```

This is the structure of the final generated configuration file, which can be created directly using the `kubectl` command function.

* public => tenant
* default => namespace
* test-func => function name