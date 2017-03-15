import argparse
import os
import subprocess
import sys

if not os.path.exists('tmp/src/github.com/taowen'):
    os.makedirs('tmp/src/github.com/taowen')

os.environ['GOBIN'] = os.path.abspath('output')
os.environ['GOPATH'] = os.path.abspath('tmp')
WORK_DIR = os.path.abspath('tmp/src/github.com/taowen/function-tracer')

if not os.path.exists(WORK_DIR):
    os.chdir(os.path.dirname(WORK_DIR))
    try:
        os.remove('motrix')
    except:
        pass
    try:
        os.symlink('../../../../', 'function-tracer')
    except:
        pass


def main():
    if len(sys.argv) == 1:
        handle_build()
        return
    argument_parser = argparse.ArgumentParser()
    sub_parsers = argument_parser.add_subparsers()
    sub_command = sub_parsers.add_parser('dep')
    sub_command.set_defaults(handler=handle_dep)
    sub_command = sub_parsers.add_parser('build')
    sub_command.set_defaults(handler=handle_build)
    sub_command = sub_parsers.add_parser('setup-transparent-proxy')
    args, _ = argument_parser.parse_known_args()
    handler_args = dict(args.__dict__)
    handler_args.pop('handler')
    args.handler(**handler_args)


def handle_build():
    pass


def handle_dep():
    GOVENDOR = '%s/bin/govendor' % os.getenv('GOPATH')
    if not os.path.exists(GOVENDOR):
        subprocess.check_call('GOBIN=%s/bin go get github.com/kardianos/govendor' % os.getenv('GOPATH'), shell=True)
    if not os.path.exists('%s/vendor' % WORK_DIR):
        subprocess.check_call('cd %s && %s init' % (WORK_DIR, GOVENDOR), shell=True)
    try:
        subprocess.check_call('cd %s && %s %s' % (WORK_DIR, GOVENDOR, ' '.join(sys.argv[2:])), shell=True)
    except:
        pass


main()
