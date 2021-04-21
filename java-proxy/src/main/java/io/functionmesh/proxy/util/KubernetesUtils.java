/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package io.functionmesh.proxy.util;

import lombok.extern.slf4j.Slf4j;
import org.apache.commons.io.FileUtils;

import java.io.File;
import java.nio.charset.StandardCharsets;

@Slf4j
public class KubernetesUtils {

	private static final String KUBERNETES_NAMESPACE_PATH = "/var/run/secrets/kubernetes.io/serviceaccount/namespace";

	public static String getNamespace() {
		try {
			File file = new File(KUBERNETES_NAMESPACE_PATH);
			return FileUtils.readFileToString(file, StandardCharsets.UTF_8);
		} catch (java.io.IOException e) {
			log.error("Get namespace from kubernetes path {}, message: {}", KUBERNETES_NAMESPACE_PATH, e.getMessage());
		}
		// Use the default namespace
		return "default";
	}
}
