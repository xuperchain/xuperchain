package com.baidu.xuper;

import com.google.protobuf.ProtocolStringList;
import java.math.BigInteger;

import com.baidu.xuper.contractpb.Contract.Transaction;
import com.baidu.xuper.contractpb.Contract.Block;

/**
 * Builtin Types
 *
 */
public class BuiltinTypes implements Contract
{
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
        if (tx == null) {
            return Response.error("no tx found");
        }

        printTx(tx);

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
        if (b == null) {
            return Response.error("no tx found");
        }

        printBlock(b);

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    // transfer native token from the contract to other acocunts
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

        ctx.transfer(to,amountStr);

        return Response.ok("ok".getBytes());
    }

    private void printTx(Transaction tx) {
        System.out.printf("txid: %s\n",tx.getTxid());
        System.out.printf("blockid: %s\n",tx.getBlockid());
        System.out.printf("desc: %s\n",tx.getDesc().toString());
        System.out.printf("initiator: %s\n",tx.getInitiator());

        ProtocolStringList authRequireList = tx.getAuthRequireList();
        for (String auth : authRequireList) {
            System.out.printf("auth require: %s\n", auth);
        }

        for (int i = 0; i < tx.getTxInputsCount(); i++){
            System.out.printf("[tx_input[%d]]: ref_txid: %s\n",i,tx.getTxInputs(i).getRefTxid());
            System.out.printf("[tx_input[%d]]: ref_offset: %d\n",i,tx.getTxInputs(i).getRefOffset());
            System.out.printf("[tx_input[%d]]: from_addr: %s\n",i,tx.getTxInputs(i).getFromAddr().toString());
            System.out.printf("[tx_input[%d]]: amount: %s\n",i,tx.getTxInputs(i).getAmount());
        }

        for (int i = 0; i < tx.getTxOutputsCount(); i++){
            System.out.printf("[tx_input[%d]]: amount: %s\n",i,tx.getTxOutputs(i).getAmount());
            System.out.printf("[tx_input[%d]]: to_addr: %s\n",i,tx.getTxOutputs(i).getToAddr().toString());
        }
    }

    private void printBlock(Block b) {
        System.out.printf("blockid: %s\n",b.getBlockid());
        System.out.printf("pre_hash: %s\n",b.getPreHash());
        System.out.printf("proposer: %s\n",b.getProposer());
        System.out.printf("sign: %s\n",b.getSign());
        System.out.printf("pubkey: %s\n",b.getPubkey());
        System.out.printf("height: %s\n",b.getHeight());

        for (int i = 0; i < b.getTxCount(); i++){
            System.out.printf("txid[%d]: %s\n",i,b.getTxids(i));
        }

        System.out.printf("tx_count: %s\n",b.getTxCount());
        System.out.printf("in_trunk: %s\n",b.getInTrunk());
        System.out.printf("next_hash: %s\n",b.getNextHash());
    }

    public static void main(String[] args) {
        Driver.serve(new BuiltinTypes());
    }
}
