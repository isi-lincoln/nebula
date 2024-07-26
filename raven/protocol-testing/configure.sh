#!/bin/bash

if [[ $EUID -ne 0 ]]; then
    echo "requires root access to run ansible playbook"
    exit 1
fi

#ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i .rvn/ansible-hosts -e 'ansible_python_interpreter=/usr/bin/python3'  ./ansible/plays/update_all_hosts.yml
#ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i .rvn/ansible-hosts -e 'ansible_python_interpreter=/usr/bin/python3'  ./ansible/plays/update_all_switches.yml
#ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook --limit "!s0" -i .rvn/ansible-hosts -e 'ansible_python_interpreter=/usr/bin/python3' -e "@./ansible/variables/config.yml" ./ansible/plays/configure_ip4_network.yml
#ANSIBLE_HOST_KEY_CHECKING=False ANSIBLE_ROLES_PATH=ansible/roles \
#    ansible-playbook --limit "!s0" -i .rvn/ansible-hosts \
#    -e 'ansible_python_interpreter=/usr/bin/python3' \
#    -e "@./ansible/variables/config.yml" \
#    ./ansible/plays/install_nebula.yml

#ANSIBLE_HOST_KEY_CHECKING=False ANSIBLE_ROLES_PATH=ansible/roles \
#    ansible-playbook --limit "!s0" -i .rvn/ansible-hosts \
#    -e 'ansible_python_interpreter=/usr/bin/python3' \
#    -e "@./ansible/variables/config.yml" \
#    ./ansible/plays/install_etcd.yml
 
ANSIBLE_HOST_KEY_CHECKING=False ANSIBLE_ROLES_PATH=ansible/roles \
    ansible-playbook --limit "!s0" -i .rvn/ansible-hosts \
    -e 'ansible_python_interpreter=/usr/bin/python3' \
    -e "@./ansible/variables/config.yml" \
    ./ansible/plays/install_avoid_go.yml
