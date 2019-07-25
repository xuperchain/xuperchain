#!/bin/bash
set -e
# check input parameter
if [ $# -lt 4 ]; then
	echo "usage: sh -x deployContract.sh accountName contractName contractNamePath args fee"
	exit
fi
accountName=$1
contractName=$2
contractNamePath=$3
args=$4
fee=$5
address=`cat ./data/keys/address`
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
echo "$accountName/$address" >> data/acl/addrs
./xchain-cli wasm deploy --account $accountName --cname $contractName -H $ip_port -m $contractNamePath --arg $args --output contractRawTx.out --fee $fee
./xchain-cli multisig get --tx ./contractRawTx.out --host $ip_port --output complianceCheck.out
./xchain-cli multisig sign --tx ./contractRawTx.out --output my.sign
./xchain-cli multisig send my.sign complianceCheck.out,my.sign --tx ./contractRawTx.out -H $ip_port
