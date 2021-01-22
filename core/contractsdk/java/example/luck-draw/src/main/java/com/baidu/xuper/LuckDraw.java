package com.baidu.xuper;

import java.math.BigInteger;
import java.util.Random;

/**
 * Counter
 */
public class LuckDraw implements Contract {
    final String ADMIN = "admin";
    final String RESULT = "result";
    final String TICKETS = "tickets";
    final String USER_ID = "userid";// id --> user
    final String USER = "user";// user--> id

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        byte[] value = ctx.args().get("admin");
        if (value == null || value.length == 0) {
            return Response.error("missing caller");
        }
        ctx.putObject(ADMIN.getBytes(), value);
        ctx.putObject(TICKETS.getBytes(), "0".getBytes());

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response getLuckId(Context ctx) {

        String caller = ctx.caller();
        if (caller == null || caller.isEmpty()) {
            return Response.error("missing caller");
        }
        if (ctx.getObject(RESULT.getBytes()) != null) {
            return Response.error("the luck draw has finished");
        }
        byte[] userVal = ctx.getObject((USER + caller).getBytes());
        if (userVal != null && userVal.length != 0) {
            return Response.ok(userVal);
        }
        BigInteger lastId = new BigInteger(new String(ctx.getObject(TICKETS.getBytes())));
        lastId = lastId.add(BigInteger.ONE);

        ctx.putObject((USER_ID + lastId).getBytes(), caller.getBytes());
        ctx.putObject(TICKETS.getBytes(), lastId.toString().getBytes());
        ctx.putObject((USER + caller).getBytes(), lastId.toString().getBytes());

        return Response.ok(lastId.toString().getBytes());
    }

    @ContractMethod
    public Response startLuckDraw(Context ctx) {

        String caller = ctx.caller();
        if (caller == null || caller.isEmpty()) {
            return Response.error("missing caller");
        }
        String admin = new String(ctx.getObject(ADMIN.getBytes()));

        if (!caller.equals(admin)) {
            return Response.error("you do not have permission to call this method");
        }
        byte[] seedByte = ctx.args().get("seed");
        if (seedByte == null || seedByte.length == 0) {
            return Response.error("missing seed");
        }

        BigInteger seed = new BigInteger(new String(seedByte));
        if (seed.compareTo(BigInteger.ZERO) == 0) {
            return Response.error("parse seed error");
        }

        BigInteger lastId = new BigInteger(new String(ctx.getObject(TICKETS.getBytes())));
        if (lastId.compareTo(BigInteger.ZERO) == 0) {
            return Response.error("no luck draw tickets");
        }

        Random rander = new Random(seed.longValue());
        long luckid = (rander.nextInt(lastId.intValue())) + 1;
        byte[] luckUser = ctx.getObject((USER_ID + String.format("%d", luckid)).getBytes());
        ctx.putObject(RESULT.getBytes(), luckUser);
        return Response.ok(luckUser);
    }

    @ContractMethod
    public Response getResult(Context ctx) {
        byte[] luckUser = ctx.getObject(RESULT.getBytes());
        if (luckUser == null || luckUser.length == 0) {
            return Response.error("get luck draw result failed");
        }
        return Response.ok(luckUser);
    }

    public static void main(String[] args) {
        Driver.serve(new LuckDraw());
    }
}