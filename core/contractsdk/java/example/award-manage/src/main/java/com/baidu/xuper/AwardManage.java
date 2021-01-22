package com.baidu.xuper;

import java.math.BigInteger;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

public class AwardManage implements Contract {
    final String BALANCE = "balanceOf_";
    final String ALLOWANCE = "allowanceOf_";
    final String MASTER = "owner";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        String caller = ctx.caller();
        if (caller.length() == 0) {
            return Response.error("missing caller");
        }
        byte[] totalSupplyByte = ctx.args().get("totalSupply");
        if (totalSupplyByte == null || totalSupplyByte.length == 0) {
            return Response.error("missing totalSupply");
        }
        BigInteger totalSupply = new BigInteger(new String(totalSupplyByte));
        if (totalSupply.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("totalSupply must be positive");
        }

        ctx.putObject((BALANCE + caller).getBytes(), totalSupply.toString().getBytes());
        ctx.putObject("totalSupply".getBytes(), totalSupply.toString().getBytes());
        ctx.putObject(MASTER.getBytes(), caller.getBytes());

        return Response.ok("ok".getBytes());
    }

    private boolean permCheck(Context ctx) {
        return ctx.caller() != null && !ctx.caller().isEmpty()
                && ctx.caller().equals(new String(ctx.getObject(MASTER.getBytes())));
    }

    @ContractMethod
    public Response addAward(Context ctx) {
        if (!permCheck(ctx)) {
            return Response.error("you do not have permission to call this method");
        }
        byte[] amountByte = ctx.args().get("amount");
        if (amountByte == null || amountByte.length == 0) {
            return Response.error("missing amount");
        }

        BigInteger amount = new BigInteger(new String(ctx.args().get("amount")));
        if (amount.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("amount must be greater than 0");
        }

        BigInteger totalSupply = new BigInteger(new String(ctx.getObject("totalSupply".getBytes())));

        ctx.putObject("totalSupply".getBytes(), totalSupply.add(amount).toString().getBytes());

        String key = BALANCE + ctx.caller();
        BigInteger value = new BigInteger(new String(ctx.getObject(key.getBytes())));

        ctx.putObject(key.getBytes(), value.add(amount).toString().getBytes());
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response totalSupply(Context ctx) {
        return Response.ok(ctx.getObject("totalSupply".getBytes()));
    }

    @ContractMethod
    public Response balance(Context ctx) {
        String caller = ctx.caller();
        if (caller == null || caller.length() == 0) {
            return Response.error("missing caller");
        }
        byte[] value = ctx.getObject((BALANCE + caller).getBytes());
        return Response.ok(value);
    }

    @ContractMethod
    public Response allowance(Context ctx) {
        byte[] fromByte = ctx.args().get("from");
        byte[] toByte = ctx.args().get("to");
        if (fromByte == null || fromByte.length == 0 || toByte == null || toByte.length == 0) {
            return Response.error("missing from or to");
        }
        String from = new String(fromByte);
        String to = new String(toByte);
        byte[] allowanceByte = ctx.getObject((ALLOWANCE + from + "_" + to).getBytes());
        if (allowanceByte == null || allowanceByte.length == 0) {
            return Response.error("allowance from " + from + " to " + to + " not found");
        }
        return Response.ok(allowanceByte);

    }

    @ContractMethod
    public Response transfer(Context ctx) {
        String from = ctx.caller();
        if (from == null || from.isEmpty()) {
            return Response.error("missing caller");
        }
        byte[] toByte = ctx.args().get("to");
        byte[] tokenByte = ctx.args().get("token");
        if (toByte == null || toByte.length == 0 || tokenByte == null || toByte.length == 0) {
            return Response.error("missing  to or token");
        }

        String to = new String((toByte));
        BigInteger token = new BigInteger(new String(tokenByte));

        if (from.equals(to)) {
            return Response.error("can not transfer to yourself");
        }

        if (token.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("token must be more than 0");
        }

        String fromKey = BALANCE + from;
        byte[] balanceFromByte = ctx.getObject(fromKey.getBytes());
        if (balanceFromByte == null || balanceFromByte.length == 0) {
            return Response.error("balance of " + from + " not found");
        }
        BigInteger fromBalance = new BigInteger(new String(balanceFromByte));
        if (fromBalance.compareTo(token) < 0) {
            return Response.error("balance not enough");
        }

        String toKey = BALANCE + to;
        byte[] balanceToByte = ctx.getObject(toKey.getBytes());
        BigInteger toBalance = new BigInteger("0");
        if (balanceToByte != null && balanceToByte.length != 0) {
            toBalance = new BigInteger(new String(balanceToByte));
        }

        ctx.putObject(fromKey.getBytes(), fromBalance.subtract(token).toString().getBytes());
        ctx.putObject(toKey.getBytes(), toBalance.add(token).toString().getBytes());
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response transferFrom(Context ctx) {
        String to = ctx.caller();
        if (to == null || to.isEmpty()) {
            return Response.error("missing caller");
        }

        if (ctx.args().get("from") == null || ctx.args().get("from").length == 0 || ctx.args().get("token") == null
                || ctx.args().get("token").length == 0) {
            return Response.error("missing from or token");
        }

        String from = new String(ctx.args().get("from"));
        BigInteger token = new BigInteger(new String(ctx.args().get("token")));

        if (token.compareTo(BigInteger.ZERO) == 0) {
            return Response.error("missing token");
        }
        if (from.equals(to)) {
            return Response.error("can not transfer from yourself");
        }
        if (token.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("token must be more than 0");
        }

        String allowanceKey = ALLOWANCE + from + "_" + to;
        byte[] allowanceByte = ctx.getObject(allowanceKey.getBytes());
        if (allowanceByte == null || allowanceByte.length == 0) {
            return Response.error("allowance from " + from + " to " + to + " not found");
        }

        BigInteger allowance = new BigInteger(new String(allowanceByte));

        if (allowance.compareTo(token) < 0) {
            return Response.error("allowance not enough");
        }
        String fromKey = BALANCE + from;
        String toKey = BALANCE + to;
        byte[] fromBalanceByte = ctx.getObject(fromKey.getBytes());
        if (fromBalanceByte == null || fromBalanceByte.length == 0) {
            return Response.error("balance of" + from + " not found");
        }

        BigInteger fromBalance = new BigInteger(new String(fromBalanceByte));
        if (fromBalance.compareTo(token) < 0) {
            return Response.error("from balancfe not enough");
        }

        BigInteger toBalance = new BigInteger("0");
        byte[] value = ctx.getObject((BALANCE + to).getBytes());
        if (value != null && value.length > 0) {
            toBalance = new BigInteger(new String(value));
        }
        ctx.putObject(fromKey.getBytes(), fromBalance.subtract(token).toString().getBytes());
        ctx.putObject((toKey).getBytes(), toBalance.add(token).toString().getBytes());
        ctx.putObject(allowanceKey.getBytes(), allowance.subtract(token).toString().getBytes());
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response approve(Context ctx) {
        String from = ctx.caller();
        if (from == null || from.isEmpty()) {
            return Response.error("missing caller");
        }

        if (ctx.caller() == null || ctx.caller().isEmpty() || ctx.args().get("token") == null
                || ctx.args().get("token").length == 0 || ctx.args().get("to") == null
                || ctx.args().get("to").length == 0) {
            return Response.error("missing caller to or token");
        }

        String to = new String(ctx.args().get("to"));
        BigInteger token = new BigInteger(new String(ctx.args().get("token")));

        if (token.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("token must be greater than 0");
        }
        if (from.equals(to)) {
            return Response.error("can not transfer to yourself");
        }

        String allowanceKey = ALLOWANCE + from + "_" + to;
        BigInteger allowance = new BigInteger("0");
        byte[] value = ctx.getObject(allowanceKey.getBytes());
        if (value != null) {
            allowance = new BigInteger(new String(value));
        }

        allowance = allowance.add(token);
        ctx.putObject(allowanceKey.getBytes(), allowance.toString().getBytes());
        return Response.ok("ok~".getBytes());

    }

    public static void main(String[] args) {
        Driver.serve(new AwardManage());
    }
}