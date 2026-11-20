#!/usr/bin/env python3
"""Regression checks for the Actions trust boundaries, without cloud credentials."""

import copy
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest

import yaml

ROOT = Path(__file__).resolve().parents[2]
WORKFLOWS = ROOT / '.github/workflows'


class WorkflowLoader(yaml.SafeLoader):
    """Use YAML 1.2 booleans so GitHub's `on` key is not parsed as True."""


WorkflowLoader.yaml_implicit_resolvers = {
    key: [(tag, pattern) for tag, pattern in values
          if tag != 'tag:yaml.org,2002:bool']
    for key, values in yaml.SafeLoader.yaml_implicit_resolvers.items()
}
WorkflowLoader.add_implicit_resolver(
    'tag:yaml.org,2002:bool', re.compile(r'^(?:true|false)$', re.I), list('tTfF'))


def load(name):
    return yaml.load((WORKFLOWS / name).read_text(), Loader=WorkflowLoader)


def violations(workflow):
    errors = []
    events = workflow.get('on', {})
    if not isinstance(events, dict):
        events = dict.fromkeys(events if isinstance(events, list) else [events])
    permissions = workflow.get('permissions')
    if not isinstance(permissions, dict):
        errors.append('explicit permissions required')
    for name, job in workflow['jobs'].items():
        effective = job.get('permissions', permissions) or {}
        text = json.dumps(job)
        if effective.get('id-token') == 'write':
            errors.append('OIDC requires a separate security review')
        if 'pull_request' in events:
            if any(value == 'write' for value in effective.values()):
                errors.append('PR job must not have write permissions')
            if 'secrets.' in text:
                errors.append('PR job must not reference secrets')
        if 'pull_request_target' in events:
            for step in job.get('steps', []):
                if 'run' in step or step.get('uses', '').startswith(('actions/checkout@', './')):
                    errors.append('target event must be metadata only')
                if '${{' in step.get('with', {}).get('script', ''):
                    errors.append('never interpolate event data into JavaScript')
        for step in job.get('steps', []):
            action = step.get('uses', '')
            if action and not re.fullmatch(r'[^@]+@[0-9a-f]{40}', action):
                errors.append('external actions require a commit pin')
            if any(value in action.lower() for value in ('ssh-access', 'tmate', 'upterm')):
                errors.append('automatic remote access is forbidden')
            if action.startswith('actions/checkout@'):
                if step.get('with', {}).get('persist-credentials') is not False:
                    errors.append('checkout must not persist credentials')
            if action.startswith('actions/setup-go@') and step['with'].get('cache') is not False:
                errors.append('Go cache must be explicitly disabled')
    return errors


class WorkflowSecurity(unittest.TestCase):
    def test_all_workflows_obey_baseline(self):
        for path in sorted(WORKFLOWS.glob('*.yml')):
            with self.subTest(workflow=path.name):
                self.assertEqual([], violations(load(path.name)))

    def test_chart_testing_uses_supported_python(self):
        steps = load('test-helm-charts.yml')['jobs']['lint-test']['steps']
        python = next(step for step in steps
                      if step.get('uses', '').startswith('actions/setup-python@'))
        self.assertEqual('3.12', python['with']['python-version'])

    def test_shell_syntax_and_e2e_installers(self):
        for path in WORKFLOWS.glob('*.yml'):
            for job in load(path.name)['jobs'].values():
                for step in job.get('steps', []):
                    if 'run' in step:
                        subprocess.run(['bash', '-n'], input=step['run'], text=True, check=True)
        for path in (ROOT / '.ci/tests').glob('integration*/e2e*.yaml'):
            config = yaml.load(path.read_text(), Loader=WorkflowLoader)
            for step in config['setup']['steps']:
                command = step.get('command', '')
                subprocess.run(['bash', '-n'], input=command, text=True, check=True)
                self.assertNotIn('get-helm-3 | bash', command)
                if step['name'] == 'install helm':
                    self.assertIn('sha256sum --check --strict', command)
                    self.assertIn('helm-v3.22.1-linux-amd64.tar.gz', command)
                if step['name'] == 'install pulsar cluster':
                    self.assertIn('checkout --detach 0f64202d0bd529dea5ab27bbe27ef6be84b2bac9', command)

    def test_corrupt_helm_download_never_executes(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            binary = base / 'bin'
            binary.mkdir()
            curl = binary / 'curl'
            curl.write_text('#!/bin/bash\nwhile (($#)); do\n'
                            '  if [[ "$1" == -o ]]; then printf corrupt > "$2"; exit 0; fi\n'
                            '  shift\ndone\nexit 1\n')
            sudo = binary / 'sudo'
            sudo.write_text('#!/bin/bash\ntouch "$INSTALL_MARKER"\n')
            curl.chmod(0o755)
            sudo.chmod(0o755)
            for path in (ROOT / '.ci/tests').glob('integration*/e2e*.yaml'):
                config = yaml.load(path.read_text(), Loader=WorkflowLoader)
                command = next(s['command'] for s in config['setup']['steps']
                               if s['name'] == 'install helm')
                result = subprocess.run(['bash', '-ec', command], capture_output=True, text=True,
                    env={**os.environ, 'PATH': f"{binary}:{os.environ['PATH']}",
                         'TMPDIR': directory, 'INSTALL_MARKER': str(base / 'installed')})
                self.assertNotEqual(0, result.returncode)
                self.assertFalse((base / 'installed').exists())

    def test_preflight_authfile_is_not_logged(self):
        steps = load('bundle-release.yml')['jobs']['openshift']['steps']
        step = next(s for s in steps if s.get('name') == 'Trigger openshift preflight test')
        self.assertIn('umask 077', step['run'])
        self.assertIn("trap 'rm -f \"$AUTHFILE\"' EXIT", step['run'])
        self.assertNotIn('cat $AUTHFILE', step['run'])
        self.assertNotIn('PFLT_LOGLEVEL=trace', step['run'])
        self.assertIn('quay.io/opdev/preflight@sha256:', step['run'])

    def test_policy_rejects_unsafe_mutations(self):
        original = load('test-integration-skywalking-e2e.yml')
        cases = []
        w = copy.deepcopy(original)
        w['jobs']['integration']['env'] = {'TOKEN': '${{ secrets.CLOUD_TOKEN }}'}
        cases.append(w)
        w = copy.deepcopy(original)
        w['permissions']['contents'] = 'write'
        cases.append(w)
        w = copy.deepcopy(original)
        w['on'] = {'pull_request_target': None}
        cases.append(w)
        w = copy.deepcopy(original)
        w['jobs']['integration']['steps'][0]['uses'] = 'owner/action@main'
        cases.append(w)
        for workflow in cases:
            self.assertTrue(violations(workflow))

    def test_pr_integration_preserves_secretless_cases(self):
        workflow = load('test-integration-skywalking-e2e.yml')
        self.assertEqual({'pull_request'}, set(workflow['on']))
        cases = workflow['jobs']['integration']['strategy']['matrix']['case']
        self.assertEqual({'.ci/tests/integration/e2e.yaml',
                          '.ci/tests/integration/e2e_with_tls.yaml'},
                         {case['e2e'] for case in cases})
        self.assertIn('github.event.pull_request.number', workflow['concurrency']['group'])

    def test_cloud_integration_is_trusted_only(self):
        workflow = load('test-integration-cloud.yml')
        self.assertEqual({'push', 'workflow_dispatch'}, set(workflow['on']))
        self.assertEqual(['master'], workflow['on']['push']['branches'])
        self.assertIsNone(workflow['on']['workflow_dispatch'])
        job = workflow['jobs']['integration']
        for guard in ("github.repository == 'streamnative/function-mesh'",
                      "github.ref == 'refs/heads/master'",
                      "github.event_name == 'push'",
                      "github.event_name == 'workflow_dispatch'"):
            self.assertIn(guard, job['if'])
        self.assertNotIn('env', job)
        secret_steps = [s for s in job['steps'] if 'secrets.' in json.dumps(s)]
        self.assertEqual(1, len(secret_steps))
        self.assertTrue(secret_steps[0]['uses'].startswith('apache/skywalking-infra-e2e@'))
        self.assertIn("matrix.case.name == 'integration test with oauth2'",
                      secret_steps[0]['env']['GSA_CREDENTIAL'])
        for step in job['steps']:
            if step.get('uses', '').startswith('actions/checkout@'):
                self.assertNotIn('ref', step['with'])
                self.assertNotIn('repository', step['with'])

    def test_write_jobs_never_execute_builds_or_artifacts(self):
        for file in ('release.yml', 'trivy_scheduled_master.yml'):
            workflow = load(file)
            self.assertNotIn('pull_request', workflow['on'])
            self.assertEqual({'contents': 'read'}, workflow['permissions'])
            upload = workflow['jobs']['upload']
            self.assertEqual('build', upload['needs'])
            for step in upload['steps']:
                self.assertNotIn('run', step)
                self.assertFalse(step['uses'].startswith(('actions/checkout@', './')))
            download = upload['steps'][0]
            self.assertTrue(download['uses'].startswith('actions/download-artifact@'))
            self.assertEqual({'name'}, set(download['with']))

    def test_release_tag_checks_gate_credentials(self):
        for name, jobs in [('release.yml', ['build']),
                           ('bundle-release.yml', ['operatorhub', 'openshift'])]:
            workflow = load(name)
            validate = workflow['jobs']['validate-tag']
            self.assertNotIn('secrets.', json.dumps(validate))
            self.assertIn("github.repository == 'streamnative/function-mesh'", validate['if'])
            for job_name in jobs:
                job = workflow['jobs'][job_name]
                self.assertEqual('validate-tag', job['needs'])
                checkout = next(s for s in job['steps']
                                if s.get('uses', '').startswith('actions/checkout@'))
                self.assertEqual('${{ needs.validate-tag.outputs.sha }}', checkout['with']['ref'])
        self.assertIn("github.ref == 'refs/heads/master'",
                      load('bundle-release.yml')['jobs']['validate-tag']['if'])

    def test_tag_validation_with_local_git_history(self):
        scripts = [load(name)['jobs']['validate-tag']['steps'][1]['run']
                   for name in ('release.yml', 'bundle-release.yml')]
        self.assertEqual(scripts[0], scripts[1])
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            origin, checkout = base / 'origin', base / 'checkout'
            def git(*args, cwd=origin):
                return subprocess.check_output(['git', *args], cwd=cwd, stderr=subprocess.DEVNULL)
            origin.mkdir()
            git('init', '-b', 'master')
            git('config', 'user.email', 'ci@example.invalid')
            git('config', 'user.name', 'CI test')
            git('commit', '--allow-empty', '-m', 'trusted')
            git('tag', 'v1.2.3')
            git('tag', '-a', 'v1.2.3-rc.1', '-m', 'annotated tag')
            git('checkout', '-b', 'feature')
            git('commit', '--allow-empty', '-m', 'unmerged')
            git('tag', 'v9.9.9')
            git('checkout', 'master')
            git('clone', str(origin), str(checkout), cwd=base)
            tags = {'v1.2.3': True, 'v1.2.3-rc.1': True, 'v9.9.9': False,
                    'v0.0.0': False, 'master': False, 'refs/pull/831/head': False,
                    '--help': False, '$(touch injected)': False, 'v1.2.3\nevil': False}
            for tag, expected in tags.items():
                with self.subTest(tag=tag):
                    output = base / 'output'
                    output.write_text('')
                    result = subprocess.run(['bash', '-e', '-c', scripts[0]], cwd=checkout,
                                            env={**os.environ, 'TAG': tag, 'GITHUB_OUTPUT': str(output)},
                                            capture_output=True, text=True)
                    self.assertEqual(expected, result.returncode == 0, result.stderr)
                    self.assertEqual(expected, output.read_text().startswith('sha='))
            self.assertFalse((checkout / 'injected').exists())

    def test_label_script_only_changes_documentation_labels(self):
        script = load('documentbot.yml')['jobs']['labeling']['steps'][0]['with']['script']
        harness = '''
        const calls = [];
        const context = {repo: {owner: 'test', repo: 'test'}, payload: {pull_request:
          {number: 1, body: process.env.BODY, labels: [{name:'no-need-doc'}, {name:'approved'}]}}};
        const github = {rest: {issues: {
          addLabels: async args => calls.push({add: args.labels}),
          removeLabel: async args => calls.push({remove: args.name})}}};
        (async () => {SCRIPT})().then(() => console.log(JSON.stringify(calls)));
        '''.replace('SCRIPT', script)
        result = subprocess.check_output(['node', '-e', harness], text=True, env={**os.environ,
            'BODY': '- [x] `approved`\n- [x] `doc-required`\n- [ ] `no-need-doc`\n'
                    '- [x] `doc`\n${{ secrets.TOKEN }}\n$(touch malicious)'} )
        self.assertEqual([{'add': ['doc-required', 'doc']}, {'remove': 'no-need-doc'}], json.loads(result))


if __name__ == '__main__':
    unittest.main(verbosity=2)
