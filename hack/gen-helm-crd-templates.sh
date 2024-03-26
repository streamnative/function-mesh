#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

set -x

INDENT="  " # space * 2
KUSTOMIZE="$1"
YQ="$2"

function crd::convert_metadata() {
    local source="$1"
    local webhook_enabled="$2"

    # metadata
    echo "metadata:"
    # metadata.annotations
    echo "annotations:" | sed  "s/^/$(crd::nindent 1)/"
    if "$webhook_enabled"; then
        echo "{{- if eq .Values.admissionWebhook.certificate.provider \"cert-manager\" }}" | sed  "s/^/$(crd::nindent 2)/"
        echo "{{- include \"function-mesh-operator.certManager.annotation\" . | nindent 4 -}}" | sed  "s/^/$(crd::nindent 3)/"
        echo "{{- end }}" | sed  "s/^/$(crd::nindent 2)/"
    fi
    # metadata.annotations - filter out ".cert-manager.io/inject-ca-from"
    "$YQ" '.metadata.annotations | with_entries(select(.key != "cert-manager.io/inject-ca-from"))' "$source" | grep -v "^{}" | sed  "s/^/$(crd::nindent 2)/"
    # metadata.name
    "$YQ" '.metadata | with_entries(select(.key == "name"))' "$source" | sed  "s/^/$(crd::nindent 1)/"
}

function crd::convert_spec() {
    local source="$1"
    local webhook_enabled="$2"

    # spec
    echo "spec:"
    if "$webhook_enabled"; then
        echo "conversion:" | sed "s/^/$(crd::nindent 1)/"
        # spec.conversion - filter out ".webhook"
        "$YQ" '.spec.conversion | with_entries(select(.key != "webhook"))' "$source" | grep -v "^{}" | sed "s/^/$(crd::nindent 2)/"
        # spec.conversion.webhook
        crd::convert_webhook "$source" | sed "s/^/$(crd::nindent 2)/"
    fi
    # spec - filter out ".conversion"
    "$YQ" '.spec | with_entries(select(.key != "conversion"))' "$source" | grep -v "^{}" | sed "s/^/$(crd::nindent 1)/"
}

function crd::convert_webhook() {
    local source="$1"

    # webhook
    echo "webhook:"
    echo "clientConfig:" | sed  "s/^/$(crd::nindent 1)/"
    # webhook.clientConfig.caBundle
    echo "{{- if eq .Values.admissionWebhook.certificate.provider \"custom\" }}" | sed  "s/^/$(crd::nindent 2)/"
    echo "{{- \$caSecret := (lookup \"v1\" \"Secret\" .Values.admissionWebhook.secretsWebhookNamespace (include \"function-mesh-operator.certificate.caSecret\" .)) -}}" | sed  "s/^/$(crd::nindent 3)/"
    echo "{{- if \$caSecret }}" | sed  "s/^/$(crd::nindent 3)/"
    echo "{{- \$caCert := (b64dec (get \$caSecret.data \"tls.crt\")) -}}" | sed  "s/^/$(crd::nindent 4)/"
    echo "{{ printf (include \"function-mesh-operator.caBundle\" .) (b64enc \$caCert) | nindent 8 }}" | sed  "s/^/$(crd::nindent 4)/"
    echo "{{- end }}" | sed  "s/^/$(crd::nindent 3)/"
    echo "{{- end }}" | sed  "s/^/$(crd::nindent 2)/"
    # webhook.clientConfig.service
    echo "service:" | sed  "s/^/$(crd::nindent 2)/"
    # webhook.clientConfig - filter out ".caBundle" and ".service"
    "$YQ" '.spec.conversion.webhook.clientConfig | with_entries(select(.key != "caBundle" and .key != "service"))' "$source" | grep -v "^{}" | sed "s/^/$(crd::nindent 3)/"
    echo "name: {{ include \"function-mesh-operator.webhook.service\" . }}" | sed  "s/^/$(crd::nindent 3)/"
    echo "namespace: {{ .Release.Namespace }}" | sed "s/^/$(crd::nindent 3)/"
    # webhook.clientConfig.service - filter out ".name" and ".namespace"
    "$YQ" '.spec.conversion.webhook.clientConfig.service | with_entries(select(.key != "name" and .key != "namespace" ))' "$source" | grep -v "^{}" | sed  "s/^/$(crd::nindent 3)/"
    # webhook - filter out ".clientConfig"
    "$YQ" '.spec.conversion.webhook | with_entries(select(.key != "clientConfig"))' "$source" | grep -v "^{}" | sed "s/^/$(crd::nindent 1)/"
}

function crd::generate_template() {
    local source="$1"
    local target="$2"
    local webhook_enabled="$3"

    if [ ! -f "$target" ];then
        tee "$target"
    fi

    > "$target"

    # head
    echo "{{- if .Values.admissionWebhook.enabled }}" >> "$target"
    # apiVersion
    "$YQ" '. | with_entries(select(.key == "apiVersion"))' "$source" | grep -v "^{}" >> "$target"
    # kind
    "$YQ" '. | with_entries(select(.key == "kind"))' "$source" | grep -v "^{}" >> "$target"
    # metadata
    crd::convert_metadata "$source" "$webhook_enabled" >> "$target"
    # spec
    crd::convert_spec "$source" "$webhook_enabled" >> "$target"
    # status
    "$YQ" '. | with_entries(select(.key == "status"))' "$source" | grep -v "^{}" >> "$target"
    # end
    echo "{{- end }}" >> "$target"
}

function crd::nindent() {
    local num="$1"
    printf "$INDENT%.0s" $(seq "$num")
}

function crd::main() {
    tmp=$(mktemp -d)

    # crd-functions
    file="crd-compute.functionmesh.io-functions.yaml"
    target_file="charts/function-mesh-operator/charts/admission-webhook/templates/$file"
    source_file="$tmp/$file"
    "$KUSTOMIZE" build config/crd | "$YQ" eval '. | select(.metadata.name == "functions.compute.functionmesh.io")' > "$source_file"
    crd::generate_template "$source_file" "$target_file" true

    # crd-sinks
    file="crd-compute.functionmesh.io-sinks.yaml"
    target_file="charts/function-mesh-operator/charts/admission-webhook/templates/$file"
    source_file="$tmp/$file"
    "$KUSTOMIZE" build config/crd | "$YQ" eval '. | select(.metadata.name == "sinks.compute.functionmesh.io")' > "$source_file"
    crd::generate_template "$source_file" "$target_file" true

    # crd-sources
    file="crd-compute.functionmesh.io-sources.yaml"
    target_file="charts/function-mesh-operator/charts/admission-webhook/templates/$file"
    source_file="$tmp/$file"
    "$KUSTOMIZE" build config/crd | "$YQ" eval '. | select(.metadata.name == "sources.compute.functionmesh.io")' > "$source_file"
    crd::generate_template "$source_file" "$target_file" true

    # crd-functionmeshes
    file="crd-compute.functionmesh.io-functionmeshes.yaml"
    target_file="charts/function-mesh-operator/charts/admission-webhook/templates/$file"
    source_file="$tmp/$file"
    "$KUSTOMIZE" build config/crd | "$YQ" eval '. | select(.metadata.name == "functionmeshes.compute.functionmesh.io")' > "$source_file"
    # set webhook_enabled to false since the functionmeshes don't have webhooks
    crd::generate_template "$source_file" "$target_file" false

    # crd-backendconfigs
    file="crd-compute.functionmesh.io-backendconfigs.yaml"
    target_file="charts/function-mesh-operator/charts/admission-webhook/templates/$file"
    source_file="$tmp/$file"
    "$KUSTOMIZE" build config/crd | "$YQ" eval '. | select(.metadata.name == "backendconfigs.compute.functionmesh.io")' > "$source_file"
    crd::generate_template "$source_file" "$target_file" true
}

crd::main