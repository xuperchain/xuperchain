#!/bin/bash
set -e
# check input parameter
if [ $# -lt 2 ]; then
    echo "usage: sh -x contractCall.sh call.json fee"
	exit
fi
callJson=$1
fee=$2
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
./xchain-cli multisig gen --desc $callJson -H $ip_port --fee $fee --output rawTx.out
./xchain-cli multisig get --tx ./rawTx.out --host $ip_port --output complianceCheck.out
./xchain-cli multisig sign --tx ./rawTx.out --output my.sign
./xchain-cli multisig send my.sign complianceCheck.out --tx ./rawTx.out -H $ip_port
