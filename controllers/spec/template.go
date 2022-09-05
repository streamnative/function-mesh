// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package spec

const (
	javaLog4jXMLTemplate = `<Configuration>
    <name>pulsar-functions-kubernetes-instance</name>
    <monitorInterval>30</monitorInterval>
    <Properties>
        <Property>
            <name>pulsar.log.level</name>
            <value>{{ .Level }}</value>
        </Property>
        <Property>
            <name>bk.log.level</name>
            <value>{{ .Level }}</value>
        </Property>
    </Properties>
    <Appenders>
        <Console>
            <name>Console</name>
            <target>SYSTEM_OUT</target>
            <PatternLayout>
                <Pattern>%d{ISO8601_OFFSET_DATE_TIME_HHMM} [%t] %-5level %logger{36} - %msg%n</Pattern>
            </PatternLayout>
        </Console>
        {{- if .RollingEnabled }}
        <RollingRandomAccessFile>
            <name>RollingRandomAccessFile</name>
            <fileName>\${sys:pulsar.function.log.dir}/\${sys:pulsar.function.log.file}.log</fileName>
            <filePattern>\${sys:pulsar.function.log.dir}/\${sys:pulsar.function.log.file}.%d{yyyy-MM-dd-hh-mm}-%i.log.gz</filePattern>
            <PatternLayout>
                <Pattern>%d{yyyy-MMM-dd HH:mm:ss a} [%t] %-5level %logger{36} - %msg%n</Pattern>
            </PatternLayout>
            <Policies>
                {{ .Policy }}
            </Policies>
            <DefaultRolloverStrategy>
                <max>5</max>
            </DefaultRolloverStrategy>
        </RollingRandomAccessFile>
       {{- end }}
    </Appenders>
    <Loggers>
        <Logger>
            <name>org.apache.pulsar.functions.runtime.shaded.org.apache.bookkeeper</name>
            <level>\${sys:bk.log.level}</level>
            <additivity>false</additivity>
            <AppenderRef>
                <ref>Console</ref>
            </AppenderRef>
            <AppenderRef>
                <ref>RollingRandomAccessFile</ref>
            </AppenderRef>
        </Logger>
        <Root>
            <level>\${sys:pulsar.log.level}</level>
            <AppenderRef>
                <ref>Console</ref>
                <level>\${sys:pulsar.log.level}</level>
            </AppenderRef>
            <AppenderRef>
                <ref>RollingRandomAccessFile</ref>
            </AppenderRef>
        </Root>
    </Loggers>
</Configuration>`
	pythonLoggingINITemplate = `[loggers]
keys=root

[handlers]
keys={{ .Handlers }}

[formatters]
keys=formatter

[logger_root]
level={{ .Level }}
handlers={{ .Handlers }}

{{- if .RollingEnabled }}
{{ .Policy }}
{{- end }}

[handler_stream_handler]
class=StreamHandler
level={{ .Level }}
formatter=formatter
args=(sys.stdout,)

[formatter_formatter]
format=[%(asctime)s] [%(levelname)s] aa %(filename)s: %(message)s
datefmt=%Y-%m-%d %H:%M:%S %z`
)
