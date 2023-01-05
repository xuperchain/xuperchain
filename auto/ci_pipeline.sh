#!/bin/bash

#-------------------------------
# 自动测试功能点：
#    1.搭建网络环境
#    2.转账
#    3.合约
#    4.权限
#------------------------------

# 脚本绝对路径
AbsPath=$(
    cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1
    pwd
)

# 根路径
WorkRoot=$AbsPath/..
# 测试网络路径
TestNet=$WorkRoot/testnet
# 工作路径：所有命令在node1路径下运行
WorkPath=${TestNet}/node1
WorkData=${WorkPath}/data
WorkKeys=${WorkData}/keys
# log
LogPath="${WorkPath}/logs"
ErrorLogFile="${LogPath}/xchain.log.wf"
# mine period (in second)
mine_period=3

# Make `alias` work in GitHub actions
shopt -s expand_aliases
alias xchain-cli='$WorkPath/bin/xchain-cli'
alias log_info='echo INFO $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'
alias log_error='echo ERROR $(date +"%Y-%m-%d %H:%M:%S") ${BASH_SOURCE##*/}:$LINENO'

## contract account
contract_account_number_utxo=3333333333333333
contract_account_utxo="XC${contract_account_number_utxo}@xuper"

# testnet
function testnet() {
    # install
    # shellcheck source=/dev/null
    source "$WorkRoot"/auto/deploy_testnet.sh || exit 2

    ## AK
    ak_node1=$(cat "${TestNet}"/node1/data/keys/address)
    ak_node2=$(cat "${TestNet}"/node2/data/keys/address)
    ak_node3=$(cat "${TestNet}"/node3/data/keys/address)

    # start nodes parallel
    (cd "$TestNet/node1" && bash control.sh start) &
    (cd "$TestNet/node2" && bash control.sh start) &
    (cd "$TestNet/node3" && bash control.sh start)
    wait

    cd "$WorkPath" || exit 2

    log_info "starting testnet path=$(pwd)"

    # check status
    for _ in $(seq 10); do
        sleep 3
        if xchain-cli status; then
            log_info "testnet started"
            return 0
        fi
    done

    log_error "start testnet failed"
    exit 2
}

# account
function account() {
    ## 账户
    xchain-cli account newkeys --output data/alice || exit 3
    xchain-cli transfer --to "$(cat data/alice/address)" --amount 10000000 || exit 3
    balance=$(xchain-cli account balance --keys data/alice)
    log_info "account $(cat data/alice/address) balance $balance"

    ## 合约账户
    xchain-cli account new --account 1111111111111111 --fee 1000 || exit 3
    xchain-cli transfer --to XC1111111111111111@xuper --amount 100000001 || exit 3
    balance=$(xchain-cli account balance XC1111111111111111@xuper)
    log_info "account XC1111111111111111@xuper balance $balance"

    ## 合约账户：desc 文件
    xchain-cli account new --desc "$WorkRoot"/data/desc/NewAccount.json --fee 1000 || exit 3
    xchain-cli transfer --to XC2222222222222222@xuper --amount 100000002 || exit 3
    balance=$(xchain-cli account balance XC2222222222222222@xuper)
    log_info "account XC2222222222222222@xuper balance $balance"
}

# contract
function contract() {
    cp "$WorkRoot"/auto/counter.wasm "$WorkPath"
    # wasm
    log_info "contract wasm"
    xchain-cli wasm deploy "$WorkPath"/counter.wasm --cname counter.wasm \
        --account XC1111111111111111@xuper \
        --runtime c -a '{"creator": "xuper"}' --fee 155537 || {
        tail "${ErrorLogFile}"
        exit 4
    }
    log_info "contract wasm invoke"
    xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100
    xchain-cli wasm query --method get -a '{"key":"test"}' counter.wasm

    # 查询用户部署的合约
    log_info "contract XC1111111111111111@xuper"
    xchain-cli account contracts --account XC1111111111111111@xuper
    log_info "contract $(cat data/keys/address)"
    xchain-cli account contracts --address "$(cat data/keys/address)"
}

# 内置合约
function builtin() {
    # reserved_contracts
    log_info "contract reserved unified_check"
    xchain-cli wasm deploy "$WorkPath"/build/unified_check --cname unified_check \
        --account XC1111111111111111@xuper \
        --runtime c -a '{"creator": "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"}' --fee 164735 || exit 4
    xchain-cli wasm invoke unified_check --method register_aks \
        -a '{"aks":"SmJG3rH2ZzYQ9ojxhbRCPwFiE9y6pD1Co,iYjtLcW6SVCiousAb5DFKWtWroahhEj4u"}' --fee 155 || exit 4

    # forbidden_contract
    log_info "contract forbidden"
    xchain-cli wasm deploy "$WorkPath"/build/forbidden --cname forbidden \
        --account XC1111111111111111@xuper \
        --runtime c -a '{"creator": "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"}' --fee 155679 || exit 4
    sleep ${mine_period}
}

function acl() {
    # acl addr
    mkdir -p data/acl

    # 设置合约账户acl
    log_info "acl account"
    echo "XC1111111111111111@xuper/${ak_node1}" > data/acl/addrs
    xchain-cli acl query --account XC1111111111111111@xuper
    xchain-cli multisig gen --desc "$WorkRoot"/data/desc/SetAccountACL.json --fee 100
    xchain-cli multisig sign --output sign.out
    xchain-cli multisig send sign.out sign.out --tx tx.out || {
        tail "${ErrorLogFile}"
        exit 5
    }
    sleep ${mine_period}
    xchain-cli acl query --account XC1111111111111111@xuper | grep "${ak_node1}" || exit 5

    # 设置合约方法acl
    log_info "acl contract method"
    echo "XC1111111111111111@xuper/${ak_node2}" >>data/acl/addrs
    xchain-cli multisig gen --desc "$WorkRoot"/data/desc/SetMethodACL.json --fee 100
    xchain-cli multisig sign --keys "$TestNet"/node1/data/keys --output sign1.out
    xchain-cli multisig sign --keys "$TestNet"/node2/data/keys --output sign2.out
    sleep ${mine_period}
    xchain-cli multisig send sign1.out,sign2.out sign1.out,sign2.out --tx tx.out || exit 5
    sleep ${mine_period}
    xchain-cli acl query --contract counter.wasm --method increase | grep "${ak_node2}" || exit 5

    # 调用合约方法
    log_info "acl invoke contract method"
    xchain-cli transfer --to "${ak_node2}" --amount 10000000 || exit 5
    xchain-cli transfer --to "${ak_node3}" --amount 10000000 || exit 5
    xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100 --keys "$TestNet"/node1/data/keys || exit 5
    xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100 --keys "$TestNet"/node2/data/keys || exit 5
    # node3节点无权限，应该调用失败
    log_info "acl node3 invoke contract method: should 'ACL not enough'"
    xchain-cli wasm invoke --method increase -a '{"key":"test"}' counter.wasm --fee 100 --keys "$TestNet"/node3/data/keys && exit 5
}

function height() {
    height1=$(xchain-cli status -H:37101 | grep trunkHeight | awk '{print $2}')
    height2=$(xchain-cli status -H:37102 | grep trunkHeight | awk '{print $2}')
    height3=$(xchain-cli status -H:37103 | grep trunkHeight | awk '{print $2}')

    log_info "height1=$height1 height2=$height2 height3=$height3"
    diff=$((2 * height1 - height2 - height3))
    if [ $diff -gt 3 ]; then
        log_error "height inconsistency: height1=$height1 height2=$height2 height3=$height3" && exit 6
    fi
}

function utxo_ak() {
    local utxo_count

    log_info "test utxo for ak"
    cd "$TestNet"/node3 || exit 7
    xchain-cli transfer --to "${ak_node3}" --amount 1000 --keys "${WorkKeys}"
    xchain-cli account balance "${ak_node3}"

    # test list
    log_info "utxo list"
    utxo_count=$(xchain-cli utxo list -N 2 -A "${ak_node3}" |
    grep utxoCount | awk -F '"' '{count += $4};END {print count}')
    if [ "${utxo_count}" -lt 1 ]; then
        log_error "before split, ${ak_node3} has ${utxo_count} UTXO, expect >= 1"
        exit 7
    fi

    # test split
    log_info "utxo split"
    xchain-cli utxo split -N 2 -A "${ak_node3}"|| {
        log_error "utxo split failed"
        exit 7
    }

    utxo_count=$(xchain-cli utxo list -N 2 -A "${ak_node3}" |
    grep utxoCount | awk -F '"' '{count += $4};END {print count}')
    if [ "${utxo_count}" -ne 2 ]; then
        log_error "after split, ${ak_node3} has ${utxo_count} UTXO, expect 2"
        exit 7
    fi

    ## test merge
    log_info "utxo merge"
    xchain-cli utxo merge -A "${ak_node3}" || {
        log_error "utxo merge failed"
        exit 7
    }

    utxo_count=$(xchain-cli utxo list -N 2 -A "${ak_node3}" |
    grep utxoCount | awk -F '"' '{count += $4};END {print count}')
    if [ "${utxo_count}" -ne 1 ]; then
        log_error "after merge, ${ak_node3} has ${utxo_count} UTXO, expect 1"
        exit 7
    fi
}

function utxo_account() {
    local utxo_count

    log_info "test utxo for account"
    cd "$TestNet"/node1 || exit 7
    xchain-cli transfer --to "${ak_node3}" --amount 1000
    xchain-cli account new --account ${contract_account_number_utxo} --fee 1000 \
        --keys "${TestNet}/node3/data/keys"
    sleep ${mine_period} # make sure account exists
    xchain-cli account balance ${contract_account_utxo}
    xchain-cli transfer --to ${contract_account_utxo} --amount 10000
    xchain-cli account balance ${contract_account_utxo}


    cd "$TestNet"/node3 || exit 7
    # test list
    log_info "utxo list"
    utxo_count=$(xchain-cli utxo list -N 2 -A ${contract_account_utxo} \
        | grep utxoCount | awk -F '"' '{count += $4};END {print count}')
    if [ "${utxo_count}" -lt 1 ]; then
        log_error "before split, ${contract_account_utxo} has ${utxo_count} UTXO, expect >= 1"
        exit 7
    fi

    # test split
    log_info "utxo split"
    xchain-cli utxo split -N 2 -A "${contract_account_utxo}" -P "${TestNet}/node3/data" \
        --keys "${TestNet}/node3/data/keys" || {
        log_error "utxo split failed"
        exit 7
    }

    utxo_count=$(xchain-cli utxo list -N 2 -A ${contract_account_utxo} \
        | grep utxoCount | awk -F '"' '{count += $4};END {print count}')
    if [ "${utxo_count}" -ne 2 ]; then
        log_error "after split, ${contract_account_utxo} has ${utxo_count} UTXO, expect 2"
        exit 7
    fi

    ## test merge
    log_info "utxo merge"
    xchain-cli utxo merge -A "${contract_account_utxo}" -P "${TestNet}/node3/data" \
        --keys "${TestNet}/node3/data/keys" || {
        log_error "utxo merge failed"
        exit 7
    }

    utxo_count=$(xchain-cli utxo list -N 2 -A ${contract_account_utxo} |
    grep utxoCount | awk -F '"' '{count += $4};END {print count}')
    if [ "${utxo_count}" -ne 1 ]; then
        log_error "after merge, ${contract_account_utxo} has ${utxo_count} UTXO, expect 1"
        exit 7
    fi
}

function utxo() {
    log_info "test utxo"
    utxo_ak
    utxo_account
}

function clean() {
    running_process_cnt=$(pgrep "xchain" | wc -l)
    if [ "${running_process_cnt}" -gt 0 ]; then
        pgrep "xchain" | xargs kill -9
    fi

    [ -d "$TestNet" ] && rm -rf "$TestNet"
}

function main() {
    log_info "test start"
    clean

    log_info "test install"
    testnet

    log_info "test account"
    account

    log_info "test contract"
    contract
    #builtin

    log_info "test acl"
    acl

    log_info "test height"
    height

    utxo

    log_info "test done"
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
        #builtin
        ;;
    Xacl)
        acl
        ;;
    Xheight)
        height
        ;;
    Xutxo)
        utxo
        ;;
    X*)
        main "$@"
        ;;
esac
