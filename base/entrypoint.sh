#!/bin/bash

user=playground
group=playground
program=$1
uid=$2

echo "127.0.0.1 $(hostname)" >> /etc/hosts

if [ -z $uid ]; then
	uid=10000
fi

groupadd $group && useradd -u "$uid" -g $group -d "/home/$user" -m $user
chgrp $group /code && chmod 0775 /code
cd /home/$user && sudo -u $user /bin/bash /run.sh $program
