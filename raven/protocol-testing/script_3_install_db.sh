#!/bin/bash

set -e

if [[ $EUID -ne 0 ]]; then
    echo "requires root access to run ansible playbook"
    exit 1
fi

ANSIBLE_HOST_KEY_CHECKING=False ANSIBLE_ROLES_PATH=ansible/roles \
    ansible-playbook -i .rvn/ansible-hosts -i ansible/variables/hosts.ini \
    -e 'ansible_python_interpreter=/usr/bin/python3' \
    -e "@./ansible/variables/config.yml" --limit "!s0" \
    ./ansible/plays/install_etcd.yml
