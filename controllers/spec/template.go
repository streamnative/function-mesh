package spec

const javaLog4jXMLTemplate = `<Configuration>
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
