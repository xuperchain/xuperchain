package com.baidu.xuper;

import com.baidu.xuper.contractpb.Contract.Block;
import com.baidu.xuper.contractpb.Contract.Transaction;
import com.google.protobuf.ProtocolStringList;

import java.io.UnsupportedEncodingException;
import java.math.BigInteger;
import java.util.Iterator;
import java.util.List;

/**
 * Builtin Types
 */
public class BuiltinTypes implements Contract {

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("initialize success".getBytes());
    }

    @ContractMethod
    public Response getTx(Context ctx) {
        byte[] txidByte = ctx.args().get("txid");
        if (txidByte == null) {
            return Response.error("missing txid");
        }
        String txid = new String(txidByte);

        Transaction tx = ctx.queryTx(txid);
        printTx(ctx, tx);

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response getBlock(Context ctx) {
        byte[] blockidByte = ctx.args().get("blockid");
        if (blockidByte == null) {
            return Response.error("missing blockid");
        }
        String blockid = new String(blockidByte);

        Block b = ctx.queryBlock(blockid);
        printBlock(ctx, b);

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response authRequire(Context ctx) {

        List<String> authRequireList = ctx.authRequire();
        printAuthRequire(ctx, authRequireList);

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    // transfer native token from the contract to other accounts
    public Response transfer(Context ctx) {
        byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        String to = new String(toByte);

        byte[] amountByte = ctx.args().get("amount");
        if (amountByte == null) {
            return Response.error("missing amount");
        }
        String amountStr = new String(amountByte);
        BigInteger amount = new BigInteger(amountStr);
        if (amount.signum() == -1) {
            return Response.error("amount must not be negative");
        }

        ctx.transfer(to, amount);

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response put(Context ctx) {
        byte[] key = ctx.args().get("key");
        if (key == null) {
            return Response.error("missing key");
        }

        byte[] value = ctx.args().get("value");
        if (value == null) {
            return Response.error("missing value");
        }

        ctx.putObject(key, value);
        return Response.ok(value);
    }

    @ContractMethod
    public Response get(Context ctx) {
        byte[] key = ctx.args().get("key");
        if (key == null) {
            return Response.error("missing key");
        }

        byte[] value = ctx.getObject(key);
        if (value == null) {
            return Response.error("key " + new String(key) + " not found");
        }

        return Response.ok(value);
    }

    @ContractMethod
    public Response getList(Context ctx) {
        byte[] start = ctx.args().get("start");
        if (start == null) {
            return Response.error("missing start");
        }

        byte[] limit = PrefixRange.generateLimit(start);
        Iterator<ContractIteratorItem> iter = ctx.newIterator(start, limit);
        int i = 0;
        while (iter.hasNext()) {
            ContractIteratorItem item = iter.next();
            String key = bytesToString(item.getKey());
            String value = bytesToString(item.getValue());
            ctx.log("item: " + i + ", key: " + key + ", value: " + value);
            i++;
        }

        return Response.ok("ok".getBytes());
    }

    private void printTx(Context ctx, Transaction tx) {
        ctx.log("txid: " + tx.getTxid());
        ctx.log("blockid: " + tx.getBlockid());
        ctx.log("blockid: " + tx.getDesc().toStringUtf8());
        ctx.log("initiator: " + tx.getInitiator());

        ProtocolStringList authRequireList = tx.getAuthRequireList();
        for (String auth : authRequireList) {
            ctx.log("auth require: " + auth);
        }

        for (int i = 0; i < tx.getTxInputsCount(); i++) {
            ctx.log("tx_input: " + i + ", ref_txid: " + tx.getTxInputs(i).getRefTxid());
            ctx.log("tx_input: " + i + ", ref_offset: " + tx.getTxInputs(i).getRefOffset());
            ctx.log("tx_input: " + i + ", from_addr: " + tx.getTxInputs(i).getFromAddr().toStringUtf8());
            ctx.log("tx_input: " + i + ", amount: " + tx.getTxInputs(i).getAmount());
        }

        for (int i = 0; i < tx.getTxOutputsCount(); i++) {
            ctx.log("tx_output: " + i + ", amount: " + tx.getTxOutputs(i).getAmount());
            ctx.log("tx_output: " + i + ", to_addr: " + tx.getTxOutputs(i).getToAddr().toStringUtf8());
        }
    }

    private void printBlock(Context ctx, Block b) {
        ctx.log("blockid: " + b.getBlockid());
        ctx.log("pre_hash: " + b.getPreHash());
        ctx.log("proposer: " + b.getProposer().toStringUtf8());
        ctx.log("sign: " + b.getSign());
        ctx.log("pubkey: " + b.getPubkey().toStringUtf8());
        ctx.log("height: " + b.getHeight());

        for (int i = 0; i < b.getTxCount(); i++) {
            ctx.log("txid: " + i + ", " + b.getTxids(i));
        }

        ctx.log("tx_count: " + b.getTxCount());
        ctx.log("in_trunk: " + b.getInTrunk());
        ctx.log("next_hash: " + b.getNextHash());
    }

    private void printAuthRequire(Context ctx, List<String> authRequireList) {
        for (int i = 0; i < authRequireList.size(); i++) {
            ctx.log("authRequire: " + i + ", " + authRequireList.get(i));
        }
    }

    private static String bytesToString(byte[] bytes) {
        if (null == bytes || bytes.length == 0) {
            return "";
        }
        String strContent = "";
        try {
            strContent = new String(bytes, "utf-8");
        } catch (UnsupportedEncodingException e) {
            e.printStackTrace();
        }
        return strContent;
    }

    public static void main(String[] args) {
        Driver.serve(new BuiltinTypes());
    }
}
