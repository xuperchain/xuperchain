package com.baidu.xuper;

import java.math.BigInteger;
import java.util.Arrays;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

/**
 * Erc20
 *
 */
public class Erc20 implements Contract
{
    static final String BALANCEPRE = "balanceOf_";
    static final String ALLOWANCEPRE = "allowanceOf_";
    static final String MASTERPRE = "owner";
    static final String TOTALSUPPLY = "totalSupply";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        String caller = ctx.caller();
        if (caller.isEmpty()){
            return Response.error("missing caller");
        }

        byte[] totalSupplyByte = ctx.args().get("totalSupply");
        if (totalSupplyByte == null) {
            return Response.error("missing totalSupply");
        }
        String totalSupplyStr = new String(totalSupplyByte);
        BigInteger totalSupply = new BigInteger(totalSupplyStr);
        if (totalSupply.signum() == -1) {
            return Response.error("totalSupply must not be negative");
        }

        String balanceKey = BALANCEPRE + caller;

        ctx.putObject(TOTALSUPPLY.getBytes(), totalSupply.toByteArray());
        ctx.putObject(balanceKey.getBytes(), totalSupply.toByteArray());
        ctx.putObject(MASTERPRE.getBytes(), caller.getBytes());

        return Response.ok("initialize success".getBytes());
    }

    @ContractMethod
    public Response mint(Context ctx) {
        String caller = ctx.caller();
        if (caller.isEmpty()){
            return Response.error("missing caller");
        }

        byte[] ownerByte = ctx.getObject(MASTERPRE.getBytes());
        if (ownerByte == null) {
            return Response.error("no owner found");
        }

        if (!Arrays.equals(caller.getBytes(), ownerByte)){
            return Response.error("only the person who created the contract can mint");
        }

        byte[] increaseSupplyByte = ctx.args().get("amount");
        if (increaseSupplyByte == null) {
            return Response.error("missing increaseSupply");
        }
        String increaseSupplyStr = new String(increaseSupplyByte);
        BigInteger increaseSupply = new BigInteger(increaseSupplyStr);
        if (increaseSupply.signum() == -1) {
            return Response.error("amount must not be negative");
        }

        byte[] totalSupplyByte = ctx.getObject(TOTALSUPPLY.getBytes());
        if (totalSupplyByte == null) {
            return Response.error("no totalSupply found");
        }
        BigInteger totalSupply = new BigInteger(totalSupplyByte);
        BigInteger totalSupplyNow = totalSupply.add(increaseSupply);

        ctx.putObject(TOTALSUPPLY.getBytes(), totalSupplyNow.toByteArray());

        String balanceKey = BALANCEPRE + caller;
        byte[] callerBalanceByte = ctx.getObject(balanceKey.getBytes());
        if (callerBalanceByte == null) {
            return Response.error("no caller found");
        }
        BigInteger callerBalance = new BigInteger(callerBalanceByte);
        BigInteger callerBalanceNow = callerBalance.add(increaseSupply);

        ctx.putObject(balanceKey.getBytes(), callerBalanceNow.toByteArray());

        return Response.ok("mint success".getBytes());
    }

    @ContractMethod
    public Response totalSupply(Context ctx) {
        byte[] value = ctx.getObject(TOTALSUPPLY.getBytes());
        if (value == null) {
            return Response.error("key TOTALSUPPLY not found)");
        }
        BigInteger totalSupply = new BigInteger(value);

        return Response.ok(totalSupply.toString().getBytes());
    }

    @ContractMethod
    public Response balance(Context ctx) {
        byte[] accountByte = ctx.args().get("account");
        if (accountByte == null) {
            return Response.error("missing account");
        }
        String account = new String(accountByte);

        String balanceKey = BALANCEPRE + account;
        byte[] value = ctx.getObject(balanceKey.getBytes());
        if (value == null) {
            return Response.error("key " + account + " not found");
        }
        BigInteger balance = new BigInteger(value);

        return Response.ok(balance.toString().getBytes());
    }

    @ContractMethod
    public Response allowance(Context ctx) {
        byte[] fromByte = ctx.args().get("from");
        if (fromByte == null) {
            return Response.error("missing from");
        }
        String from = new String(fromByte);

        byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        String to = new String(toByte);

        String allowanceKey = ALLOWANCEPRE + from + "_" + to;
        byte[] value = ctx.getObject(allowanceKey.getBytes());
        if (value == null) {
            return Response.error("key " + allowanceKey + " not found");
        }
        BigInteger allowance = new BigInteger(value);

        return Response.ok(allowance.toString().getBytes());
    }

    @ContractMethod
    public Response owner(Context ctx) {
        byte[] ownerByte = ctx.getObject(MASTERPRE.getBytes());
        if (ownerByte == null) {
            return Response.error("no owner found");
        }

        return Response.ok(ownerByte);
    }

    @ContractMethod
    public Response transfer(Context ctx) {
        String from = ctx.caller();
        if (from.isEmpty()){
            return Response.error("missing from");
        }

        byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        String to = new String(toByte);

        byte[] tokenByte = ctx.args().get("token");
        if (tokenByte == null) {
            return Response.error("missing token");
        }
        String tokenAmountStr = new String(tokenByte);
        BigInteger tokenAmount = new BigInteger(tokenAmountStr);
        if (tokenAmount.signum() == -1) {
            return Response.error("token must not be negative");
        }

        String fromKey = BALANCEPRE + from;
        byte[] fromBalanceByte = ctx.getObject(fromKey.getBytes());
        if (fromBalanceByte == null) {
            return Response.error("no from found");
        }
        BigInteger fromBalance = new BigInteger(fromBalanceByte);
        if (fromBalance.compareTo(tokenAmount) == -1){
            return Response.error("no enough balance");
        }

        String toKey = BALANCEPRE + to;
        byte[] toBalanceByte = ctx.getObject(toKey.getBytes());
        BigInteger toBalance;
        if (toBalanceByte == null) {
            toBalance = BigInteger.valueOf(0);
        } else {
            toBalance = new BigInteger(toBalanceByte);
        }

        BigInteger fromBalanceNow = fromBalance.subtract(tokenAmount);
        BigInteger toBalanceNow = toBalance.add(tokenAmount);

        ctx.putObject(fromKey.getBytes(), fromBalanceNow.toByteArray());
        ctx.putObject(toKey.getBytes(), toBalanceNow.toByteArray());

        return Response.ok("transfer success".getBytes());
    }

    @ContractMethod
    public Response transferFrom(Context ctx) {
        String caller = ctx.caller();
        if (caller.isEmpty()){
            return Response.error("missing caller");
        }

        byte[] fromByte = ctx.args().get("from");
        if (fromByte == null) {
            return Response.error("missing from");
        }
        String from = new String(fromByte);

        byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        String to = new String(toByte);

        byte[] tokenByte = ctx.args().get("token");
        if (tokenByte == null) {
            return Response.error("missing token");
        }
        String tokenAmountStr = new String(tokenByte);
        BigInteger tokenAmount = new BigInteger(tokenAmountStr);
        if (tokenAmount.signum() == -1) {
            return Response.error("token must not be negative");
        }

        String allowanceKey = ALLOWANCEPRE + from + "_" + caller;
        byte[] value = ctx.getObject(allowanceKey.getBytes());
        if (value == null) {
            return Response.error("key " + allowanceKey + " not found");
        }
        BigInteger allowance = new BigInteger(value);
        if (allowance.compareTo(tokenAmount) == -1){
            return Response.error("The allowance is not enough");
        }

        String fromKey = BALANCEPRE + from;
        byte[] fromBalanceByte = ctx.getObject(fromKey.getBytes());
        if (fromBalanceByte == null) {
            return Response.error("no from found");
        }
        BigInteger fromBalance = new BigInteger(fromBalanceByte);
        if (fromBalance.compareTo(tokenAmount) == -1){
            return Response.error("The balance of from is not enough");
        }

        String toKey = BALANCEPRE + to;
        byte[] toBalanceByte = ctx.getObject(toKey.getBytes());
        BigInteger toBalance;
        if (toBalanceByte == null) {
            toBalance = BigInteger.valueOf(0);
        } else {
            toBalance = new BigInteger(toBalanceByte);
        }

        BigInteger fromBalanceNow = fromBalance.subtract(tokenAmount);
        BigInteger toBalanceNow = toBalance.add(tokenAmount);
        BigInteger allowanceBalanceNow = allowance.subtract(tokenAmount);

        ctx.putObject(fromKey.getBytes(), fromBalanceNow.toByteArray());
        ctx.putObject(toKey.getBytes(), toBalanceNow.toByteArray());
        ctx.putObject(allowanceKey.getBytes(), allowanceBalanceNow.toByteArray());

        return Response.ok("transferFrom success".getBytes());
    }

    @ContractMethod
    public Response approve(Context ctx) {
        String from = ctx.caller();
        if (from.isEmpty()){
            return Response.error("missing caller");
        }

        byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        String to = new String(toByte);

        byte[] tokenByte = ctx.args().get("token");
        if (tokenByte == null) {
            return Response.error("missing token");
        }
        String tokenAmountStr = new String(tokenByte);
        BigInteger tokenAmount = new BigInteger(tokenAmountStr);
        if (tokenAmount.signum() == -1) {
            return Response.error("token must not be negative");
        }

        String fromKey = BALANCEPRE + from;
        byte[] fromBalanceByte = ctx.getObject(fromKey.getBytes());
        if (fromBalanceByte == null) {
            return Response.error("no from found");
        }
        BigInteger fromBalance = new BigInteger(fromBalanceByte);
        if (fromBalance.compareTo(tokenAmount) == -1){
            return Response.error("The balance of from not enough");
        }

        String allowanceKey = ALLOWANCEPRE + from + "_" + to;
        byte[] allowanceBalanceByte = ctx.getObject(allowanceKey.getBytes());
        BigInteger allowanceBalance;
        if (allowanceBalanceByte == null) {
            allowanceBalance = BigInteger.valueOf(0);
        } else {
            allowanceBalance = new BigInteger(allowanceBalanceByte);
        }

        BigInteger allowanceBalanceNow = allowanceBalance.add(tokenAmount);

        ctx.putObject(allowanceKey.getBytes(), allowanceBalanceNow.toByteArray());

        return Response.ok("approve success".getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new Erc20());
    }
}
