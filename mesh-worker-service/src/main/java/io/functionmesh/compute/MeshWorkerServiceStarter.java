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
package io.functionmesh.compute;

import lombok.extern.slf4j.Slf4j;
import org.apache.pulsar.functions.worker.WorkerConfig;

import java.io.File;

@Slf4j
public class MeshWorkerServiceStarter {

    public static void main(String[] args) throws Exception {

        // Test yaml file
        File file = new File("src/main/resources/functions_worker.yml");
        System.out.println(file.getAbsolutePath());
        WorkerConfig workerConfig = WorkerConfig.load(file.getAbsolutePath());
        final MeshWorker worker = new MeshWorker(workerConfig);
        try {
            worker.start();
            log.info("Start Function Mesh Proxy successfully");
        } catch (Throwable th) {
            log.error("Encountered error in function worker.", th);
            worker.stop();
            Runtime.getRuntime().halt(1);
        }
    }
}
