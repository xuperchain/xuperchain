bootNodeIP=`host boot_node |awk '{print $4}'`
./xchain-cli netURL gen -H boot_node:37101
bootNodeNetURL="  - `./xchain-cli netURL get -H boot_node:37101|sed s/127.0.0.1/$bootNodeIP/g`"
echo "bootNodenetURL:" $bootNodeNetURL
sed -i "s|#  - \"/ip4/<ip>/tcp/<port>/p2p/<node_hash>\" for p2pv2 or - \"<ip>:<port>\" for p2pv1|$bootNodeNetURL|g" conf/xchain.yaml 
sed -i 's/#bootNodes/bootNodes/g'  conf/xchain.yaml
cat conf/xchain.yaml
./xchain 
