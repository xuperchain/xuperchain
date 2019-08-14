#!/bin/bash

#-------------------------------
#自动测试功能点：
#    1. 搭建网络环境
#    2. tdpos网络提名与投票
#    3. 合约的部署与调用
#------------------------------

basepath=$(cd `dirname $0`; pwd)
outputpath=$(cd `dirname $0`; cd ../; pwd)
echo "--------> basepath=$basepath"
echo "--------> outputpath=$outputpath"

function get_addrs()
{
	addr1=$(cat $basepath/node1/data/keys/address)
	addr2=$(cat $basepath/node2/data/keys/address)
	addr3=$(cat $basepath/node3/data/keys/address)
	addr=($addr1 $addr2 $addr3)
	echo "--------> addr=[$addr1 $addr2 $addr3]"
}
function deploy_env()
{
	mkdir node1 node2 node3
	#拷贝节点，更改node2,node3节点的相关配置
	for((i=3;i>=1;i--))
	{
		tcp_port="3710$i"
		metric_port="3720$i"
        p2p_port="4710$i"
		cp -r $outputpath/output/* $basepath/node$i
		if [ $i -ge 2 ];then
			rm -rf $basepath/node$i/data/keys && rm -rf $basepath/node$i/data/netkeys
			cd $basepath/node$i
			$basepath/node$i/xchain-cli account newkeys
			$basepath/node$i/xchain-cli netURL gen
			sed -i'' -e "18s/  port:.*/  port: :$tcp_port/" $basepath/node$i/conf/xchain.yaml
			sed -i'' -e "20s/  metricPort:.*/  metricPort: :$metric_port/" $basepath/node$i/conf/xchain.yaml
			sed -i'' -e "27s/  port:.*/  port: $p2p_port/" $basepath/node$i/conf/xchain.yaml
		fi
		#修改xuper.json文件
		if [ $i -eq 1 ];then
			addr1=`cat $basepath/node1/data/keys/address`
			addr2=`cat $basepath/node2/data/keys/address`
			neturl2=`cd $basepath/node2 && ./xchain-cli netURL preview --port 47102`
			timestamp=`date +%s`
			proposer_num=2
		    echo "timestamp=$timestamp proposer_num=$proposer_num addr2=$addr2 neturl2=$neturl2"
			cp $basepath/relate_file/xuper.json $basepath/node$i/data/config/xuper.json
			sed -i'' -e 's/\("timestamp": "\).*/\1'"$timestamp"'000000000"\,/' $basepath/node$i/data/config/xuper.json
			sed -i'' -e 's/\("proposer_num": "\).*/\1'"$proposer_num"'"\,/' $basepath/node$i/data/config/xuper.json
			sed -i'' -e "s/nodeaddress/$addr2/" $basepath/node$i/data/config/xuper.json
            sed -i'' -e "s@nodeneturl@$neturl2@" $basepath/node$i/data/config/xuper.json
            cp $basepath/node1/data/config/xuper.json $basepath/node2/data/config/xuper.json
            cp $basepath/node1/data/config/xuper.json $basepath/node3/data/config/xuper.json
			cd $basepath/node$i
			cd $basepath/node$i && $basepath/node$i/xchain-cli createChain
			nohup $basepath/node$i/xchain &
			sleep 5
			netUrl=$($basepath/node$i/xchain-cli netURL get)
			echo $netUrl > $basepath/node$i/neturl.txt
			hostname=`ifconfig -a | grep inet | grep -v 127.0.0.1 | grep -v inet6 | awk '{print $2}' | tr -d "addrs:"`
			sed -i'' -e 's/127.0.0.1/'"$hostname"'/' $basepath/node$i/neturl.txt
		fi
    }
	sed -i'' -e "s/#bootNodes/bootNodes/; s@#  - \"/ip4/<ip>.*@  - $(cat $basepath/node1/neturl.txt)@" $basepath/node2/conf/xchain.yaml
    sed -i'' -e "s/#bootNodes/bootNodes/; s@#  - \"/ip4/<ip>.*@  - $(cat $basepath/node1/neturl.txt)@" $basepath/node3/conf/xchain.yaml
    ##node2,3节点创建链并启动
	cd $basepath/node2 && $basepath/node2/xchain-cli createChain && nohup $basepath/node2/xchain &
	cd $basepath/node3 && $basepath/node3/xchain-cli createChain && nohup $basepath/node3/xchain &
	sleep 12
	get_height
	get_addrs
}

function get_height()
{
	for ((j=1;j<=3;j++))
	{
		node1_height=$($basepath/node1/xchain-cli status -H=localhost:37101 | grep trunkHeight | awk -F': ' '{print $NF-1}')
		node_height=$($basepath/node$j/xchain-cli status -H=localhost:3710$j | grep trunkHeight | awk -F': ' '{print $NF-1}')
		cond=$(($node_height-$node1_height))
		echo "node1_height=$node1_height node_height=$node_height cond=$cond"
		if [ $cond -lt 3 ]; then
			echo -e "\033[42;30m node$j: $node_height \033[0m \n"
		else
			echo -e "\033[43;35m the node is much behind the height!!!  \033[0m \n"
			exit 1
		fi
	}
}

function gen_file()
{
	mkdir $basepath/relate_file
	cd $basepath && python gen_file.py
}

function tdpos_nominate()
{
	for ((i=1;i<=3;i++))
	{
		nominate_addr=$(cat $basepath/node$i/data/keys/address)
		nominate_url=$(cd $basepath/node$i && ./xchain-cli netURL preview --port 4710$i)
		cd $basepath/node1
		sed -i'' -e 's@\("neturl": "\).*@\1'"$nominate_url"'"\,@' $basepath/relate_file/nominate.json
		sed -i'' -e 's/\("candidate": "\).*/\1'"$nominate_addr"'"/' $basepath/relate_file/nominate.json
		echo $addr1 > $basepath/node1/addrs
		echo ${addr[$i-1]} >> $basepath/node1/addrs
		./xchain-cli multisig gen --to=$(cat ./data/keys/address) --desc=$basepath/relate_file/nominate.json --multiAddrs=$basepath/node1/addrs --amount=1100000000027440 --frozen=-1
		./xchain-cli multisig sign --tx=./tx.out --keys=./data/keys --output=./key.sign
		./xchain-cli multisig sign --tx=./tx.out --keys=$basepath/node$i/data/keys --output=./key$i.sign
		./xchain-cli multisig send --tx=./tx.out ./key.sign ./key.sign,./key$i.sign
		sleep 3
	}
	sleep 6
	check_nominate
}

function check_nominate()
{
    cd $basepath/node1
    check_base=$(./xchain-cli tdpos query-candidates -H=127.0.0.1:37101)
    echo "nominate result of node1 is:"$check_base
    for ((i=1;i<=3;i++))
    {
        result=$(./xchain-cli tdpos query-candidates -H=127.0.0.1:3710$i)
        if [ "$check_base" = "$result" ];then
            echo -e "\033[42;30m node$i is the same as node1 \033[0m \n"
        else
            echo -e "\033[43;35m node$i is different from node1  \033[0m \n"
			exit 1
		fi
    }

}

#初始化：node1,node2出块；vote结束：node1.node3出块
function vote_nominate()
{
	#投票结果，使的后node2与node3节点的出块
	cd $basepath/node1
	nominate_list=("$addr2\",\"$addr3" "$addr1\",\"$addr3")
	echo "--------> ${nominate_list[0]} ${nominate_list[1]}"
	before_info=$(get_TermProposer)
	echo "--------> $before_info"
	for ((i=0;i<=1;i++))
	{
		sed -i'' -e "4s/\".*/\"${nominate_list[$i]}\"/" $basepath/relate_file/vote.json
		txid_out=$(./xchain-cli transfer --to=$(cat ./data/keys/address) --desc=$basepath/relate_file/vote.json --amount=$[$[i+1]*100] --frozen=-1)
		echo $txid_out > $basepath/relate_file/txid.txt
		sleep 1
	}
	before_term=$(echo $before_info | awk -F"=| " '{print $(NF-3)}')
	before_proposers=$(echo $before_info | awk -F"proposers=" '{print $NF}')
	if [ $before_term = 1 ];then
		before_term=$[$before_term+1]
	fi
	after_term2=$[$before_term+1]
	for ((i=1;i<=43;i++))
	{
	    for ch in - \\ \| /
	    {
	        printf "%ds waiting...%s\r" $[$i*4] $ch
	        sleep 1
	    }
	}
	result1_out=$(./xchain-cli tdpos query-checkResult -t=$before_term)
	result1=$(echo $result1_out | awk -F':' '{print $3}')
	result2_out=$(./xchain-cli tdpos query-checkResult -t=$after_term2)
	result2=$(echo $result2_out | awk -F':' '{print $3}')
	expected_results="[$addr1 $addr3]"
	expected_results2="[$addr3 $addr1]"
	echo "before vote:"$result1_out
	echo "after vote:"$result2_out
	if [ "$result2" = "$expected_results" ] || [ "$result2" = "$expected_results2" ];then
		echo -e "\033[42;30m  vote result is right~~~\033[0m \n"
	else
		echo -e "\033[43;35m  vote result is not right!!!\033[0m \n"
		exit 1
	fi
}

function get_TermProposer()
{
	cd $basepath/node1
	while((1))
	do
		log_out=$(tail -30 ./logs/xchain.log | grep "getTermProposer" 2>&1)
		if [ "$log_out" = "" ];then
			continue
		else
			echo "$log_out"
			break
		fi
	done
}
#revoke结束：node2，node3出块
function tdpos_revoke()
{
	cd $basepath/node1
	while read -r line
	do
		sed -i'' -e 's/\("txid": "\).*/\1'"$line"'"/' $basepath/relate_file/revoke.json
		./xchain-cli transfer --to=$(cat ./data/keys/address) --desc=$basepath/relate_file/revoke.json --amount=1
	done < $basepath/relate_file/txid.txt
    before_revoke=$(get_TermProposer)
	before_term=$(echo $before_revoke | awk -F"=| " '{print $(NF-3)}')
	after_term=$[$before_term+3]
	for ((i=1;i<=80;i++))
	{
	    for ch in - \\ \| /
	    {
	        printf "%ds waiting...%s\r" $[$i*4] $ch
	        sleep 1
	    }
	}
	term1_out=$(./xchain-cli tdpos query-checkResult -t=$before_term)
	term2_out=$(./xchain-cli tdpos query-checkResult -t=$after_term)
	result1=$(echo $term1_out | awk -F':' '{print $3}')
	result2=$(echo $term2_out | awk -F':' '{print $3}')
	expected_results="[$addr2 $addr3]"
	expected_results2="[$addr3 $addr2]"
	echo "before revoke:"$term1_out
	echo "after revoke:"$term2_out
	if [ "$result2" = "$expected_results" ] || [ "$result2" = "$expected_results2" ];then
		echo "result2=$result2 expected_results=$expected_results"
		echo -e "\033[42;30m  revoke result is right~~~\033[0m \n"
	else
		echo "result2=$result2 expected_results=$expected_results"
		echo -e "\033[43;35m  revoke result is not right!!!\033[0m \n"
		exit 1
	fi
}

#合约部分
function deploy_invoke_contract()
{
	cp $basepath/counter.wasm  $basepath/relate_file/
	cd $basepath/node1 
	rand_name=`date +%y%s%m%d`
	sed -i'' -e "s/\(\"account_name\": \"\).*/\1$rand_name\"\,/; s/TH/$addr2/" $basepath/relate_file/account.json
	account_out=$(./xchain-cli account new --desc=$basepath/relate_file/account.json --fee=1000)
	account_name=$(echo $account_out | awk -F'account name: ' '{print $2}')
	echo $account_name > $basepath/relate_file/account_name.txt

	mkdir $basepath/node1/data/acl && touch addrs
	echo "$account_name/$addr1" > $basepath/node1/data/acl/addrs
	echo "$account_name/$addr2" >> $basepath/node1/data/acl/addrs
	./xchain-cli transfer --to=$account_name --amount=5000000
	sleep 2
	#发起部署
	will_takeout=$(./xchain-cli wasm deploy --account=$account_name --cname counter --arg '{"creator":"counterwasm"}' -m --multiAddrs=$basepath/node1/data/acl/addrs --output=./tx.out $basepath/relate_file/counter.wasm)
	will_fee=$(echo $will_takeout | awk -F' ' '{print $8}')
	./xchain-cli wasm deploy --account=$account_name --cname counter --arg '{"creator":"counterwasm"}' -m --multiAddrs=$basepath/node1/data/acl/addrs --output=./tx.out $basepath/relate_file/counter.wasm --fee=$will_fee
	./xchain-cli multisig sign --tx=./tx.out --keys=$basepath/node1/data/keys --output=./key1.sign
	./xchain-cli multisig sign --tx=./tx.out --keys=$basepath/node2/data/keys --output=./key2.sign
	sleep 2
	./xchain-cli multisig send --tx ./tx.out ./key1.sign,./key2.sign ./key1.sign,./key2.sign
	sleep 3
	will_out=$(./xchain-cli wasm invoke counter -a '{"key":"dudu"}' --method increase)
	will_fee=$(echo $will_out | awk -F' ' '{print $9}')
	output=$(./xchain-cli wasm invoke counter -a '{"key":"dudu"}' --method increase --fee=$will_fee 2>&1)
	echo "output :$output"
	will_result=$(echo $output | awk -F' ' '{print $3}')
	query_out=$(./xchain-cli wasm query  -a '{"key":"dudu"}' counter --method get 2>&1)
	echo "will_result :$will_result"
	echo "query_out :$query_out"
	query_result=$(echo $query_out | awk -F': ' '{print $2}')
	echo "query_result :$query_result"
	if [ $will_result = $query_result ];then
		echo -e "\033[42;30m will_result=$will_result query_result=$query_result \033[0m \n"
	else
		echo -e "\033[43;35m will_result=$will_result query_result=$query_result \033[0m \n"
		exit 1
	fi
}
echo "--------> start to gen relate file"
gen_file
sleep 2
echo "--------> start to deploy env"
deploy_env
sleep 3
echo "--------> start to tdpos nominate"
get_addrs
tdpos_nominate
sleep 2
echo "--------> start to tdpos vote nominate"
vote_nominate
sleep 2
echo "--------> start to tdpos revoke"
tdpos_revoke
sleep 2
echo "--------> start to deploy_invoke_contract"
deploy_invoke_contract
echo "test is end~~"
