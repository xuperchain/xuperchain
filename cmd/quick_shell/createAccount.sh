#!/bin/bash
set -e
# check input parameter
if [ $# -lt 1 ]; then
	echo "usage: sh -x createAccount.sh newAccount.json"
	exit
fi
newAccount=$1
# get testnet conf including ip:port and complianceCheck public key
source "./testnet.conf"
acl="./data/acl"
if [ ! -d "$acl" ]; then
    mkdir $acl
fi
addrs="./data/acl/addrs"
if [ ! -f "$addrs" ]; then
    touch $addrs
fi
echo $public_key > data/acl/addrs
./xchain-cli multisig gen --desc $newAccount -H $ip_port --fee 1000 --output rawTx.out
./xchain-cli multisig get --tx ./rawTx.out --host $ip_port --output complianceCheck.out
./xchain-cli multisig sign --tx ./rawTx.out --output my.sign
./xchain-cli multisig send my.sign complianceCheck.out --tx ./rawTx.out -H $ip_port
