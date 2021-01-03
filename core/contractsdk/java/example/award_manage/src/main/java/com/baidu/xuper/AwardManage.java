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
        BigDecimal  totalSupply = new BigDecimal(new String(ctx.args().get("totalSupply")));
        if (totalSupply.compareTo(BigDecimal.ZERO) <=0) {
            return Response.error("totalSupply must be positive");
        }

        ctx.putObject((BALANCE+caller).getBytes(), totalSupply.toString().getBytes());
        ctx.putObject("totalSupply".getBytes(), totalSupply.toString().getBytes());
        ctx.putObject(MASTER.getBytes(), caller.getBytes());
        return Response.ok("ok".getBytes());
    }

    private boolean permCheck(Context ctx) {
        return true; // TODO
    }

    public Response addAward(Context ctx) {
        if (!permCheck(ctx)) {
            return Response.error("you do not have permission to do this operation");
        }

        BigDecimal amount = new BigDecimal(new String(ctx.args().get("amount".getBytes())));
        if (BigDecimal.ZERO.compareTo(amount) >= 0) {
            return Response.error("amount must be positive");
        }

        BigDecimal totalSupply = new BigDecimal(new String(ctx.getObject("totalSupply".getBytes())));

        if (new BigDecimal(0).compareTo(totalSupply) >= 0) {
            return Response.error("totalSupply must positive");
        }

        ctx.putObject("TotalSupply".getBytes(), totalSupply.add(amount).toString().getBytes());
        String key = BALANCE + ctx.caller();
        BigDecimal value = new BigDecimal(new String(ctx.getObject(key.getBytes()))).add(totalSupply);
        ctx.putObject(key.getBytes(), value.toString().getBytes());
        return Response.ok(value.toString().getBytes());
    }

    public Response totalSupply(Context ctx) {
        return Response.ok(ctx.getObject("totalSupply".getBytes()));
    }

    public Response balance(Context ctx) {
        String caller = new String(ctx.getObject("caller".getBytes()));
        return Response.ok(ctx.getObject((BALANCE + caller).getBytes()));
    }

    public Response allowance(Context ctx) {
        String from = new String(ctx.args().get("from".getBytes()));
        String to = new String(ctx.args().get("to".getBytes()));
        if (from.length() == 0 || to.length() == 0) {
            return Response.error("missing from or to");
        }
        return Response.ok(ctx.getObject((ALLOWANCE + from + "_" + to).getBytes()));

    }

    @ContractMethod
    public Response transfer(Context ctx) {
        String from = new String(ctx.args().get("from".getBytes()));
        String to = new String(ctx.args().get("to".getBytes()));

        BigDecimal token = new BigDecimal(new String(ctx.args().get("token")));
        if (from.length() == 0 || to.length() == 0 || token.compareTo(BigDecimal.ZERO) == 0) {
            return Response.error("missing from or to or token");
        }

        if (from.equals(to)) { // TODO
            return Response.error("you can not transfer to yourself");
        }

        if (token.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("token must be more than 0");
        }

        String from_key = BALANCE + from;

        BigDecimal from_balance = new BigDecimal(new String(ctx.getObject(from_key.getBytes())));

        if (from_balance.compareTo(token) < 0) {
            return Response.error("balance not enough");
        }
        String to_key = BALANCE + to;
        BigDecimal to_balance = new BigDecimal(new String(ctx.getObject(to_key.getBytes())));

        ctx.putObject(from_key.getBytes(), from_balance.subtract(token).toString().getBytes());
        ctx.putObject(to_key.getBytes(), to_balance.add(token).toString().getBytes());
        return Response.ok("transfer success".getBytes());
    }

    @ContractMethod
    public Response transferFrom(Context ctx) {

        String from = new String(ctx.args().get("from".getBytes()));
        String to = new String(ctx.args().get("to".getBytes()));
        BigDecimal token = new BigDecimal(new String(ctx.args().get("token")));
        String caller = new String(ctx.args().get("caller"));
        if (from.length() == 0 || to.length() == 0 || token.compareTo(BigDecimal.ZERO) == 0 || caller.length() == 0) {
            return Response.error("missing from or to or token");
        }
        if (from.equals(to)) {
            return Response.error("you can not transfer to yourself");
        }
        if (token.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("token must be more than 0");
        }
        String allowance_key = ALLOWANCE + from + "_" + caller;
        BigDecimal allowance_value = new BigDecimal(new String(ctx.getObject(allowance_key.getBytes())));
        if (allowance_value.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("you need to add allowance from_to");
        }

        if (allowance_value.compareTo(token) < 0) {
            return Response.error("The allowance of from_to not enough");
        }
        String from_key = BALANCE + from;
        BigDecimal from_balance = new BigDecimal(new String (ctx.getObject(from_key.getBytes())));


        if (from_balance.compareTo(token) < 0) {
            return Response.error("the balance of from not enough");
        }
        BigDecimal  to_balance = new BigDecimal(new String( ctx.getObject((BALANCE+ to).getBytes())));
        ctx.putObject(from_key.getBytes(), from_balance.subtract(token).toString().getBytes());
        ctx.putObject((BALANCE+to).getBytes(), to_balance.add(token).toString().getBytes());
        ctx.putObject(allowance_key.getBytes(), allowance_value.subtract(token).toString().getBytes());
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response approve(Context ctx) {
        String from =  new String(ctx.args().get("from".getBytes()));
        String  to = new String(ctx.args().get("to".getBytes()));
        BigDecimal  token = new BigDecimal(new String(ctx.args().get("token"))); // TODO
        if (from.length() == 0 || to.length() == 0 || token.equals(BigDecimal.ZERO)) {
            return Response.error("missing from or to or token");
        }
        if (from.equals(to)) {
            return Response.error("you can not transfer to yourself");
        }

        String  from_key = BALANCE+from;
        BigDecimal from_balance =new BigDecimal(new String( ctx.getObject(from_key.getBytes())));

        if (from_balance.compareTo(token)<0)  {
            return Response.error("balance not enough");
        }
        String  allowance_key = ALLOWANCE+ from+"_"+to;
        ctx.putObject(allowance_key.getBytes(), new BigDecimal(new String(ctx.getObject(allowance_key.getBytes()))).add(token).toString().getBytes());
        return Response.ok("approve succeed".getBytes());

    }

    public static void main(String[] args) {
        Driver.serve(new AwardManage());
    }
}