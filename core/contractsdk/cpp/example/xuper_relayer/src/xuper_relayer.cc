#include "xchain/xchain.h"
#include "relayer.pb.h"
#include <string>
#include <cstdint>
#include <cstdlib>
#include <memory>

// 常量字符串: bucket, 分隔符, 编码等
// 区块头Bucket
const char* blockHeaderBucket = "BLOCK_HEADER";
// 当前账本最新区块状态维护Bucket
const char* ledgerMetaBucket = "LEDGER_META";
// 用于编码
const char* hextable = "0123456789abcdef";
// MerkleTree验证时, 用户输入sibling的分隔符
const char delimiter = ',';

struct XuperRelayer : public xchain::Contract {}; 

// 将sibling拆分出来，客户端以','分割输入
void split(const std::string& rawProofPath, std::vector<std::string>& proof) {
    if (rawProofPath == "") {
        return;
    }
    int i = 0;
    int rawProofPathSize = rawProofPath.size();
    for (; i < rawProofPathSize; ++i) {
        if (rawProofPath[i] == delimiter) {
            continue;
        }   
        break;
    }   
    if (i >= rawProofPathSize) {
        return;
    }
    std::string delimStr = std::string(1, delimiter);
    std::string str = rawProofPath.substr(i) + delimStr;
    size_t pos = std::string::npos;
    while ((pos=str.find(delimStr)) != std::string::npos) {
        std::string temp = str.substr(0, pos);
        if (temp != "") {
            proof.push_back(temp);
        }   
        str = str.substr(pos+1,str.size());
    }   
    return;
}

// encodeHex, decodeFromHex, fromHexChar
// 编解码作用，可读字符串与不可读字符串之间的转换
bool fromHexChar(char src, char* dst) {
    if (src >= '0' && src <= '9') {
        *dst = src - '0';
    } else if (src >= '0' && src <= 'f') {
        *dst = src - 'a' + 10;
    } else if (src >= 'A' && src <= 'F') {
        *dst = src - 'A' + 10;
    } else {
        return false;
    }

    return true;
}

bool decodeFromHex(const std::string& src, std::string& dst) {
    int i = 0;
    int j = 1;
    int len = src.size();
    char tmp1;
    char tmp2;

    for (; j < len; j += 2) {
        bool succ = fromHexChar(src[j-1], &tmp1);
        if (!succ) {
            return false;
        }
        succ = fromHexChar(src[j], &tmp2);
        if (!succ) {
            return false;
        }
        dst[i] = ((uint8_t)tmp1) << 4 | ((uint8_t)tmp2);
        i++;
    }

    if (len%2 == 1) {
        bool succ = fromHexChar(src[j-1], &tmp2);
        if (!succ) {
            return false;
        }
    }

    return true;
}

bool encodeHex(const std::string& src, std::string& dst) {
    int len = src.size();
    if (len > 500) {
        return false;
    }
    int index[1000] = {0};
    int k = 0;
    for (int i = 0; i < len; i++) {
        index[k++] = (((uint8_t)(src[i]))>>4);
        index[k++] = (((uint8_t)(src[i]))&0x0f);
    }
    for (int i = 0; i < 2*len; i++) {
        dst[i] = hextable[index[i]];
    }

    return true;
}

// 3个确认块
bool within3Confirms(xchain::Context* ctx, const std::string& blockid, const std::string tipBlockid) {
    int i = 0;
    std::string currBlockid = tipBlockid;
    while (i < 3) {
        if (currBlockid == blockid) {
            return false;
        }
        std::string blockHeaderStr;
        const std::string currBlockHeaderKey = std::string(blockHeaderBucket) + currBlockid;
        std::unique_ptr<relayer::InternalBlock> blockHeader(new relayer::InternalBlock);
        if (!ctx->get_object(currBlockHeaderKey, &blockHeaderStr)) {
            return false;
        }
        blockHeader->ParseFromString(blockHeaderStr);
        const std::string preHashStr = blockHeader->pre_hash();
        std::string preHash = std::string(64, 'o');
        if (!encodeHex(preHashStr, preHash)) {
            return false;
        }
        currBlockid = preHash;
        i += 1;
    }

    return true;
}

// 分叉管理
bool handleFork(xchain::Context* ctx, const std::string& oldTip, const std::string& newTipPre, std::string nextHash) {
    // nextHash是不可见的
    // oldTip是可见的
    // newTipPre是可见的
    std::string p = oldTip;
    std::string q = newTipPre;
    while (p != q) {
        std::string pBlockStr;
        const std::string pKey = std::string(blockHeaderBucket) + p;
        if (!ctx->get_object(pKey, &pBlockStr)) {
            return false;
        }
        std::unique_ptr<relayer::InternalBlock> pBlock(new relayer::InternalBlock);
        pBlock->ParseFromString(pBlockStr);
        pBlock->set_in_trunk(false);
        pBlock->set_next_hash("");

        std::string qBlockStr;
        const std::string qKey = std::string(blockHeaderBucket) + q;
        if (!ctx->get_object(qKey, &qBlockStr)) {
            return false;
        }
        std::unique_ptr<relayer::InternalBlock> qBlock(new relayer::InternalBlock);
        qBlock->ParseFromString(qBlockStr);
        qBlock->set_in_trunk(true);
        qBlock->set_next_hash(nextHash);

        nextHash = qBlock->blockid();
        // 编码成可视化的blockid
        if (!encodeHex(pBlock->pre_hash(), p)) {
            return false;
        }
        if (!encodeHex(qBlock->pre_hash(), q)) {
            return false;
        }

        pBlock->SerializeToString(&pBlockStr);
        qBlock->SerializeToString(&qBlockStr);
        if (!ctx->put_object(pKey, pBlockStr) ||
            !ctx->put_object(qKey, qBlockStr)) {
            return false;
        }
    }

    std::string splitBlockStr;
    const std::string splitKey = std::string(blockHeaderBucket) + q;
    if (!ctx->get_object(splitKey, &splitBlockStr)) {
        return false;
    }
    std::unique_ptr<relayer::InternalBlock> splitBlock(new relayer::InternalBlock);
    splitBlock->ParseFromString(splitBlockStr);
    splitBlock->set_in_trunk(true);
    splitBlock->set_next_hash(nextHash);
    splitBlock->SerializeToString(&splitBlockStr);
    if (!ctx->put_object(splitKey, splitBlockStr)) {
        return false;
    }

    return true;
}

// 初始化工作，将锚点区块写入，初始化LedgerMeta
DEFINE_METHOD(XuperRelayer, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(XuperRelayer, initAnchorBlockHeader) {
    xchain::Context* ctx = self.context();
    std::unique_ptr<relayer::LedgerMeta> meta(new relayer::LedgerMeta);
    std::unique_ptr<relayer::InternalBlock> anchorBlockHeader(new relayer::InternalBlock);
    std::string anchorBlockHeaderStr = ctx->arg("blockHeader");
    if (anchorBlockHeaderStr.size()==0) {
        ctx->error("missing blockHeader");
        return;
    }   
    bool succ = anchorBlockHeader->ParseFromString(anchorBlockHeaderStr);
    if (!succ) {
        ctx->error("parse from string error");
        return;
    }   

    std::string blockidBuf = anchorBlockHeader->blockid();
    std::string visualBlockid = std::string(64, '0');
    if (!encodeHex(blockidBuf, visualBlockid)) {
        ctx->error("encodeHex blockid failed");
        return;
    }

    std::string metaStr;
    const std::string metaKey = std::string(ledgerMetaBucket);
    if (ctx->get_object(metaKey, &metaStr)) {
        ctx->error("initAnchorBlockHeader should be called only once");
        return;
    }
    const std::string anchorBlockHeaderKey = std::string(blockHeaderBucket) + visualBlockid;
    meta->set_root_blockid(visualBlockid);
    meta->set_tip_blockid(visualBlockid);
    meta->set_trunk_height(anchorBlockHeader->height());
    meta->SerializeToString(&metaStr);
    if (!ctx->put_object(anchorBlockHeaderKey, anchorBlockHeaderStr) ||
        !ctx->put_object(metaKey, metaStr)) {
        ctx->error("put anchorBlockHeader or ledgerMeta failed");
        return;
    }

    ctx->ok("initAnchorBlockHeader succeed");
}

DEFINE_METHOD(XuperRelayer, putBlockHeader) {
    xchain::Context* ctx = self.context();
    std::unique_ptr<relayer::LedgerMeta> meta(new relayer::LedgerMeta);
    std::unique_ptr<relayer::InternalBlock> blockHeader(new relayer::InternalBlock);
    // 提取ledgerMeta
    std::string metaStr;
    const std::string metaKey = std::string(ledgerMetaBucket);
    if (!ctx->get_object(metaKey, &metaStr)) {
        ctx->error("missing ledger meta info");
        return;
    }
    meta->ParseFromString(metaStr);
    // 提取blockHeader
    std::string blockHeaderStr = ctx->arg("blockHeader");
    bool succ = blockHeader->ParseFromString(blockHeaderStr);
    if (!succ) {
        ctx->error("parse from string error");
        return;
    }
    std::string blockidBuf = blockHeader->blockid();
    std::string visualBlockid = std::string(64, 'o');
    if (!encodeHex(blockidBuf, visualBlockid)) {
        ctx->error("encodeHex blockid failed");
        return;
    }
    const std::string blockHeaderKey = std::string(blockHeaderBucket) + visualBlockid;
    std::string tmp;
    if (ctx->get_object(blockHeaderKey, &tmp)) {
        ctx->error(visualBlockid + " has existed already");
        return;
    }
    // 判断区块类型
    std::string preHashBuf = blockHeader->pre_hash();
    // root链
    if (preHashBuf.size() == 0) {
        ctx->error("preHash shouldn't be empty");
        return;
    }
    std::string visualPreHash = std::string(64, 'o');
    if (!encodeHex(preHashBuf, visualPreHash)) {
        ctx->error("encodeHex preHash failed");
        return;
    }
    // 判断preBlockcHeader是否存在
    const std::string preBlockHeaderKey = std::string(blockHeaderBucket) + visualPreHash;
    std::string preBlockHeaderStr;
    std::unique_ptr<relayer::InternalBlock> preBlockHeader(new relayer::InternalBlock);
    if (!ctx->get_object(preBlockHeaderKey, &preBlockHeaderStr)) {
        ctx->error("missing preHash:" + visualPreHash);
        return;
    }
    preBlockHeader->ParseFromString(preBlockHeaderStr);
    // 在主干上添加
    if (meta->tip_blockid() == visualPreHash) {
        blockHeader->set_in_trunk(true);
        preBlockHeader->set_next_hash(blockHeader->blockid());
        meta->set_tip_blockid(visualBlockid);
        meta->set_trunk_height(meta->trunk_height()+1);
        // 更新preBlockHeader
        preBlockHeaderStr = "";
        preBlockHeader->SerializeToString(&preBlockHeaderStr);
        if (!ctx->put_object(preBlockHeaderKey, preBlockHeaderStr)) {
            ctx->error("put " + visualPreHash + " failed");
            return;
        }
    } else {
        // 在分支上添加
        if (preBlockHeader->height()+1 > meta->trunk_height()) {
            // 分支变主干
            meta->set_trunk_height(preBlockHeader->height()+1);
            meta->set_tip_blockid(visualBlockid);
            blockHeader->set_in_trunk(true);
            // 处理分叉
            bool succ = handleFork(ctx, meta->tip_blockid(), visualPreHash, blockHeader->blockid());
            if (!succ) {
                ctx->error("handle fork failed");
                return;
            }
        }
    }
    // 判断blockid是否正确
    // 判断矿工签名是否正确
    // 判断2/3签名是否正确
    
    // 更新区块头信息
    blockHeader->SerializeToString(&blockHeaderStr);
    if (!ctx->put_object(blockHeaderKey, blockHeaderStr)) {
        ctx->error("put " + visualBlockid + " failed");
        return;
    }
    // 更新ledger meta信息
    meta->SerializeToString(&metaStr);
    if (!ctx->put_object(metaKey, metaStr)) {
        ctx->error("put ledger meta failed");
        return;
    }
}

DEFINE_METHOD(XuperRelayer, verifyTx) {
    xchain::Context* ctx = self.context();
    const std::string blockid = ctx->arg("blockid");
    const std::string txid = ctx->arg("txid");
    const std::string proofPathStr = ctx->arg("proofPath");
    const int txIndex = atoi(ctx->arg("txIndex").c_str());
    // 输入参数检查
    if (blockid.size() != 64) {
        ctx->error("blockid's size, expect 64, but got " + std::to_string(blockid.size()));
        return;
    }
    if (txid.size() != 64) {
        ctx->error("txid's, expect 64, but got " + std::to_string(txid.size()));
        return;
    }
    if (txIndex < 0) {
        ctx->error("txIndex expect >= 0, but got " + std::to_string(txIndex));
        return;
    }
    std::vector<std::string> proofPath;
    std::vector<std::string> proofPathEncode;
    // 交易存在确认
    split(proofPathStr, proofPath);
    std::string merkleRoot;
    // 将可读字符串编码为不可读字符串
    std::string txidEncode = std::string(32, 'o');
    if (!decodeFromHex(txid, txidEncode)) {
        ctx->error("encodeHex " + txid + " failed");
        return;
    }
    for (int i=0; i < proofPath.size(); i++) {
        std::string tmp = std::string(32, 'o');
        if (!decodeFromHex(proofPath[i], tmp)) {
            ctx->error("encodeHex proof path failed");
            return;
        }
        proofPathEncode.push_back(tmp);
    }
    // ctx->calc_merkle_root(txidEncode, txIndex, proofPathEncode, merkleRoot);
    // 终局状态确认
    const std::string metaKey = std::string(ledgerMetaBucket);
    std::string metaStr;
    if (!ctx->get_object(metaKey, &metaStr)) {
        ctx->error("get ledger meta failed");
        return;
    }
    // 是否在主干上
    std::string blockHeaderKey = blockHeaderBucket + blockid;
    std::string blockHeaderStr;
    if (!ctx->get_object(blockHeaderKey, &blockHeaderStr)) {
        ctx->error("get blockid failed");
        return;
    }
    std::unique_ptr<relayer::InternalBlock> blockHeader(new relayer::InternalBlock);
    blockHeader->ParseFromString(blockHeaderStr);
    if (blockHeader->in_trunk() == false) {
        ctx->error("blockid is not in trunk");
        return;
    }
    // merkle compare
    /*
    if (merkleRoot != blockHeader->merkle_root()) {
    }*/
    std::unique_ptr<relayer::LedgerMeta> meta(new relayer::LedgerMeta);
    meta->ParseFromString(metaStr);
    bool confirmed = within3Confirms(ctx, blockid, meta->tip_blockid());
    if (!confirmed) {
        ctx->error("block is not within 3 blocks");
        return;
    }

    ctx->ok("tx has been on chain and has confirmed for at least 3 blocks");
}

// 打印区块头
DEFINE_METHOD(XuperRelayer, printBlockHeader) {
    xchain::Context* ctx = self.context();
    const std::string key = std::string(blockHeaderBucket) + ctx->arg("blockid");
    std::string blockHeaderStr;
    if (!ctx->get_object(key, &blockHeaderStr)) {
        ctx->error("get block header faile");
        return;
    }

    std::string preHash = std::string(64, 'o');
    std::string merkleRoot = std::string(64, 'o');
    std::string nextHash = std::string(64, 'o');
    std::string sign = std::string(144, 'o');

    std::unique_ptr<relayer::InternalBlock> blockHeader(new relayer::InternalBlock);
    bool succ = blockHeader->ParseFromString(blockHeaderStr);
    if (!succ) {
        ctx->error("parse block header error");
        return;
    }

    int32_t version = blockHeader->version();
    int32_t nonce = blockHeader->nonce();
    const std::string preHashBuf = blockHeader->pre_hash();
    const std::string proposerBuf = blockHeader->proposer();
    const std::string signBuf = blockHeader->sign();
    const std::string pubkeyBuf = blockHeader->pubkey();
    const std::string merkleRootBuf = blockHeader->merkle_root();
    if (!encodeHex(preHashBuf, preHash)) {
        ctx->error("encodeHex pre hash error");
        return;
    }
    if (!encodeHex(merkleRootBuf, merkleRoot)) {
        ctx->error("encodeHex merkle root error");
        return;
    }
    if (!encodeHex(signBuf, sign)) {
        ctx->error("encodeHex sign error");
        return;
    }
    int64_t height = blockHeader->height();
    int64_t timestamp = blockHeader->timestamp();
    // transactions TODO
    int32_t txCount = blockHeader->tx_count();
    // merkle tree TODO
    int64_t curTerm = blockHeader->curterm();
    bool inTrunk = blockHeader->in_trunk();
    const std::string nextHashBuf = blockHeader->next_hash();
    if (!encodeHex(nextHashBuf, nextHash)) {
        ctx->error("encodeHex next hash error");
        return;
    }

    std::string val;
    val += "\nversion:" + std::to_string(version) + \
           "\nnonce:" + std::to_string(nonce) + \
           "\npre_hash:" + preHash + \
           "\nproposer:" + proposerBuf + \
           "\nsign:" + sign + \
           "\npubkey:" + pubkeyBuf + \
           "\nmerkle_root:" + merkleRoot + \
           "\nheight:" + std::to_string(height) + \
           "\ntimestamp:" + std::to_string(timestamp) + \
           "\ntx_count:" + std::to_string(txCount) + \
           "\ncurTerm:" + std::to_string(curTerm) + \
           "\nnext_hash:" + nextHash + "\n";
    if (inTrunk) {
        val += std::string("in_trunk: true");
    } else {
        val += std::string("in_trunk: false");
    }

    ctx->ok(val);
}
