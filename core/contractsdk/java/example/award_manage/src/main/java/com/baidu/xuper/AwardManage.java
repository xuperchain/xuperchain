package com.baidu.xuper;

import java.math.BigDecimal;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

/**
 * Counter
 */
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
        BigDecimal totalSupply = new BigDecimal(new String(ctx.args().get("totalSupply")));
        if (totalSupply.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("totalSupply must be positive");
        }

        ctx.putObject((BALANCE + caller).getBytes(), totalSupply.toString().getBytes());
        ctx.putObject("totalSupply".getBytes(), totalSupply.toString().getBytes());
        ctx.putObject(MASTER.getBytes(), caller.getBytes());

        return Response.ok("ok".getBytes());
    }

    private boolean permCheck(Context ctx) {
        ctx.log("111");
        return ctx.caller() != null &&
                !ctx.caller().isEmpty() &&
                ctx.caller().equals(new String(ctx.getObject(MASTER.getBytes())));
    }


    @ContractMethod
    public Response addAward(Context ctx) {
        if (!permCheck(ctx)) {
            return Response.error("you do not have permission to call this method");
        }
        ctx.log("1");
        BigDecimal amount = new BigDecimal(new String(ctx.args().get("amount")));
        if (BigDecimal.ZERO.compareTo(amount) >= 0) {
            return Response.error("amount must be greater than 0");
        }
        ctx.log("2");

        BigDecimal totalSupply = new BigDecimal(new String(ctx.getObject("totalSupply".getBytes())));

        if (new BigDecimal(0).compareTo(totalSupply) >= 0) {
            return Response.error("totalSupply must positive");
        }
        ctx.log("3");
        ctx.putObject("TotalSupply".getBytes(), totalSupply.add(amount).toString().getBytes());
        String key = BALANCE + ctx.caller();
        ctx.log("4");
        BigDecimal value = new BigDecimal(new String(ctx.getObject(key.getBytes()))).add(totalSupply);
        ctx.putObject(key.getBytes(), value.toString().getBytes());
        return Response.ok(value.toString().getBytes());
    }

    @ContractMethod
    public Response totalSupply(Context ctx) {
        return Response.ok(ctx.getObject("totalSupply".getBytes()));
    }

    @ContractMethod
    public Response balance(Context ctx) {
        String caller = ctx.caller();
        if (caller==null||caller.length()==0){
            return Response.error("missing caller");
        }
        byte[] value     = ctx.getObject((BALANCE+caller).getBytes());
        if (value==null||value.length==0){
            return Response.error("balance of "+caller+" not found");
        }
        return Response.ok(value);
    }

    @ContractMethod
    public Response allowance(Context ctx) {
        String from = new String(ctx.args().get("from"));
        String to = new String(ctx.args().get("to"));
        if (from.length() == 0 || to.length() == 0) {
            return Response.error("missing from or to");
        }
        return Response.ok(ctx.getObject((ALLOWANCE + from + "_" + to).getBytes()));

    }

    @ContractMethod
    public Response transfer(Context ctx) {
        String from = ctx.caller();
        if (from == null || from.isEmpty()) {
            return Response.error("missing caller");
        }
        if (ctx.args().get("to") == null || ctx.args().get("token") == null) {
            return Response.error("missing  to or token1");
        }


        String to = new String(ctx.args().get("to"));

        BigDecimal token = new BigDecimal(new String(ctx.args().get("token")));

        if (to.length() == 0 || token.compareTo(BigDecimal.ZERO) == 0) {
            return Response.error("missing to or token2");
        }

        if (from.equals(to)) {
            return Response.error("can not transfer to yourself");
        }

        if (token.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("token must be more than 0");
        }

        String from_key = BALANCE + from;

        byte[] balanceFromByte = ctx.getObject(from_key.getBytes());
        if (balanceFromByte == null || balanceFromByte.length == 0) {
            return Response.error("balance of " + from + " not found");
        }
        BigDecimal from_balance = new BigDecimal(new String(balanceFromByte));

        if (from_balance.compareTo(token) < 0) {
            return Response.error("balance not enough");
        }
        String to_key = BALANCE + to;
        byte[] balanceToByte = ctx.getObject(to_key.getBytes());
        BigDecimal to_balance = new BigDecimal(0);
        if (balanceToByte != null && balanceToByte.length != 0) {
            to_balance = new BigDecimal(new String(balanceToByte));
        }
        ctx.putObject(from_key.getBytes(), from_balance.subtract(token).toString().getBytes());
        ctx.putObject(to_key.getBytes(), to_balance.add(token).toString().getBytes());
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response transferFrom(Context ctx) {

        if (ctx.args().get("from") == null||ctx.args().get("token")==null) {
            return Response.error("missing from or token");
        }

        String from = new String(ctx.args().get("from"));
        BigDecimal token = new BigDecimal(new String(ctx.args().get("token")));

        String to = ctx.caller();
        if (to == null || to.isEmpty()) {
            return Response.error("missing caller");
        }
        if (from.length() == 0 || token.compareTo(BigDecimal.ZERO) == 0) {
            return Response.error("missing from or token");
        }
        if (from.equals(to)) {
            return Response.error("can not transfer to yourself");
        }
        if (token.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("token must be more than 0");
        }

        String allowance_key = ALLOWANCE + from + "_" + to;
        BigDecimal allowance = new BigDecimal(0);
        byte[] value = ctx.getObject(allowance_key.getBytes());
        if(value ==null){
            return Response.error("allowance from " + from + " to "+ to +" not found");
        }

        if (allowance.compareTo(token) < 0) {
            return Response.error("allowance not enough");
        }
        String from_key = BALANCE + from;
        BigDecimal from_balance = new BigDecimal(new String(ctx.getObject(from_key.getBytes())));


        if (from_balance.compareTo(token) < 0) {
            return Response.error("from balancfe not enough");
        }

        BigDecimal to_balance = new BigDecimal(0);
        value = ctx.getObject((BALANCE + to).getBytes());
        if (value!=null){
            to_balance = new BigDecimal(new String(value));
        }
        ctx.putObject(from_key.getBytes(), from_balance.subtract(token).toString().getBytes());
        ctx.putObject((BALANCE + to).getBytes(), to_balance.add(token).toString().getBytes());
        ctx.putObject(allowance_key.getBytes(), allowance.subtract(token).toString().getBytes());
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response approve(Context ctx) {
        String from = ctx.caller();
        if (from == null || from.isEmpty()) {
            return Response.error("missing caller");
        }

        if (ctx.caller() == null || ctx.args().get("token") == null || ctx.args().get("to") == null) {
            return Response.error("missing caller to or token");
        }

        String to = new String(ctx.args().get("to"));
        BigDecimal token = new BigDecimal(new String(ctx.args().get("token")));
        if (to.length() == 0 || token.equals(BigDecimal.ZERO)) {
            return Response.error("missing to or token4");
        }
        if (from.equals(to)) {
            return Response.error("you can not transfer to yourself");
        }

        String from_key = BALANCE + from;
        BigDecimal from_balance = new BigDecimal(new String(ctx.getObject(from_key.getBytes())));

        if (from_balance.compareTo(token) < 0) {
            return Response.error("balance not enough");
        }
        String allowance_key = ALLOWANCE + from + "_" + to;
        BigDecimal allowance = new BigDecimal(0);
        byte[] value = ctx.getObject(allowance_key.getBytes());
        if (value != null) {
            allowance = new BigDecimal(new String(value));
        }
        allowance = allowance.add(token);
        ctx.putObject(allowance_key.getBytes(), allowance.toString().getBytes());
        return Response.ok("ok~".getBytes());

    }

    public static void main(String[] args) {
        Driver.serve(new AwardManage());
    }
}