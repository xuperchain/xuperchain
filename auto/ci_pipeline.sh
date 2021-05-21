#!/bin/bash

#-------------------------------
# 自动测试功能点：
#    1.搭建网络环境
#    2.转账
#    3.合约
#    4.权限
#------------------------------

# 脚本绝对路径
AbsPath=$(cd $(dirname "$BASH_SOURCE"); pwd)
# 根路径
WorkRoot=$AbsPath/..
# 测试网络路径
TestNet=$WorkRoot/testnet
# 工作路径：所有命令在node1路径下运行
WorkPath=$TestNet/node1

alias xchain-cli='$WorkPath/bin/xchain-cli'
alias info='echo INFO $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'
alias error='echo ERROR $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'

# testnet
function testnet() {
  # install
  source $WorkRoot/auto/deploy_testnet.sh || exit

  # start node
  (cd "$TestNet/node1" && sh control.sh start) &
  (cd "$TestNet/node2" && sh control.sh start) &
  (cd "$TestNet/node3" && sh control.sh start)
  wait

  cd "$WorkPath" || exit

  sleep 1s
  info "start testnet path=$(pwd)"
  xchain-cli status || exit
}

# account
function account() {
  ## 账户
  xchain-cli account newkeys --output data/alice || exit
  xchain-cli transfer --to "$(cat data/alice/address)" --amount 10000000 || exit
  balance=$(xchain-cli account balance --keys data/alice)
  info "account $(cat data/alice/address) balance $balance"

  ## 合约账户
  xchain-cli account new --account 1111111111111111 --fee 1000 || exit
  xchain-cli transfer --to XC1111111111111111@xuper --amount 100000001 || exit
  balance=$(xchain-cli account balance XC1111111111111111@xuper)
  info "account XC1111111111111111@xuper balance $balance"

  ## 合约账户：desc 文件
  xchain-cli account new --desc $WorkRoot/data/desc/NewAccount.json --fee 1000 || exit
  xchain-cli transfer --to XC2222222222222222@xuper --amount 100000002 || exit
  balance=$(xchain-cli account balance XC2222222222222222@xuper)
  info "account XC2222222222222222@xuper balance $balance"
}

# contract
function contract() {
  # 合约放在 $WorkPath/build 路径下
  # 目前合约还没有移植过来，使用xuperchain的编译产出
  cp -r $WorkRoot/build $WorkPath

  # native
  info "contract native"
  xchain-cli native deploy $WorkPath/build/counter --cname counter \
            --account XC1111111111111111@xuper \
            --runtime go -a '{"creator": "xuper"}' --fee 12975371 || exit
  info "contract native invoke"
  xchain-cli native invoke --method Increase -a '{"key":"test"}' counter --fee 100
  xchain-cli native query --method Get -a '{"key":"test"}' counter

  # wasm
  info "contract wasm"
  xchain-cli wasm deploy $WorkPath/build/counter.wasm --cname counter.wasm \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "xuper"}' --fee 155537 || exit
  info "contract wasm invoke"
  xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100
  xchain-cli wasm query --method get -a '{"key":"test"}' counter.wasm

  # 查询用户部署的合约
  info "contract XC1111111111111111@xuper"
  xchain-cli account contracts --account XC1111111111111111@xuper
  info "contract $(cat data/keys/address)"
  xchain-cli account contracts --address $(cat data/keys/address)
}

# 内置合约
function builtin() {
      # reserved_contracts
  info "contract reserved unified_check"
  xchain-cli wasm deploy $WorkPath/build/unified_check --cname unified_check \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"}' --fee 164735 || exit
  xchain-cli wasm invoke unified_check --method register_aks \
            -a '{"aks":"SmJG3rH2ZzYQ9ojxhbRCPwFiE9y6pD1Co,iYjtLcW6SVCiousAb5DFKWtWroahhEj4u"}' --fee 155 || exit

  # forbidden_contract
  info "contract forbidden"
  xchain-cli wasm deploy $WorkPath/build/forbidden --cname forbidden \
            --account XC1111111111111111@xuper \
            --runtime c -a '{"creator": "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"}' --fee 155679 || exit
  sleep 3s
}

function acl() {
  # acl addr
  mkdir -p data/acl

  # 设置合约账户acl
  info "acl account"
  echo "XC1111111111111111@xuper/$(cat $TestNet/node1/data/keys/address)" > data/acl/addrs
  xchain-cli acl query --account XC1111111111111111@xuper
  xchain-cli multisig gen --desc $WorkRoot/data/desc/SetAccountAcl.json --fee 100
  xchain-cli multisig sign --output sign.out
  xchain-cli multisig send sign.out sign.out --tx tx.out || exit
  sleep 2s
  xchain-cli acl query --account XC1111111111111111@xuper

  # 设置合约方法acl
  info "acl contract method"
  echo "XC1111111111111111@xuper/$(cat $TestNet/node2/data/keys/address)" >> data/acl/addrs
  xchain-cli multisig gen --desc $WorkRoot/data/desc/SetMethodAcl.json --fee 100
  xchain-cli multisig sign --keys $TestNet/node1/data/keys --output sign1.out
  xchain-cli multisig sign --keys $TestNet/node2/data/keys --output sign2.out
  xchain-cli multisig send sign1.out,sign2.out sign1.out,sign2.out --tx tx.out  || exit
  sleep 2s
  xchain-cli acl query --contract counter.wasm --method increase

  # 调用合约方法
  info "acl invoke contract method"
  xchain-cli transfer --to "$(cat $TestNet/node2/data/keys/address)" --amount 10000000 || exit
  xchain-cli transfer --to "$(cat $TestNet/node3/data/keys/address)" --amount 10000000 || exit
  xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100 --keys $TestNet/node1/data/keys || exit
  xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100 --keys $TestNet/node2/data/keys || exit
  # node3节点无权限，应该调用失败
  info "acl node3 invoke contract method: should 'ACL not enough'"
  xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100 --keys $TestNet/node3/data/keys && exit
}

function height() {
  height1=$(xchain-cli status -H:36301 | grep trunkHeight | awk '{print $2}')
  height2=$(xchain-cli status -H:36302 | grep trunkHeight | awk '{print $2}')
  height3=$(xchain-cli status -H:36303 | grep trunkHeight | awk '{print $2}')

  info "height1=$height1 height2=$height2 height3=$height3"
  diff=$((2*height1-height2-height3))
  if [ $diff -gt 3 ]; then
		error "height inconsistency: height1=$height1 height2=$height2 height3=$height3" && exit
	fi
}

function clean() {
  num=$(ps -ef | grep "xuperos/testnet" | grep -v grep | wc -l)
  if [ $num -gt 0 ]; then
    ps -ef | grep "xuperos/testnet" | grep -v grep | awk '{print $2}' | xargs kill -9
  fi

  [ -d "$TestNet" ] && rm -rf "$TestNet"
}

function main() {
  info "test start"
  clean

  info "test install"
  testnet

  info "test account"
  account

  info "test contract"
  contract
  builtin

  info "test acl"
  acl

  info "test height"
  height

  info "test done"
}

case X$1 in
    Xclean)
        clean
        ;;
    Xtestnet)
        testnet
        ;;
    Xaccount)
        account
        ;;
    Xcontract)
        contract
        builtin
        ;;
    Xacl)
        acl
        ;;
    Xheight)
        height
        ;;
    X*)
        main "$@"
esac