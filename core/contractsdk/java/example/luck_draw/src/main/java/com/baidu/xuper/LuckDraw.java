package com.baidu.xuper;

import com.sun.tools.javac.util.ByteBuffer;

import java.math.BigDecimal;
import java.util.Arrays;
import java.util.Random;

/**
 * Counter
 */
public class LuckDraw implements Contract {
    final String admin_key = "admin";
    final String userID = "result";
    final String result = "tickets";
    final String ticketId = "userid";
    final String tickets = "tickets";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        String admin = new String(ctx.args().get(admin_key));
        if (admin.length() == 0) {
            return Response.error("missing admin address");
        }
        ctx.putObject(admin.getBytes(), admin_key.getBytes());

        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response getLuckId(Context ctx) {
        String caller = ctx.caller();
        if (caller.isEmpty()) {
            return Response.error("missing initiator");
        }
        if (ctx.getObject(result.getBytes()).length != 0) {
            return Response.error("this luck draw has finished");
        }
        byte[] userVal = ctx.getObject((userID + caller).getBytes());
        if (userVal.length > 0) {
            return Response.ok(null);
        }
        BigDecimal lastId = new BigDecimal(new String(ctx.getObject(tickets.getBytes())));
        lastId = lastId.add(BigDecimal.ONE);
        ctx.putObject((userID + caller).getBytes(), lastId.toString().getBytes());
        ctx.putObject(ticketId.getBytes(), lastId.toString().getBytes());
        ctx.putObject(tickets.getBytes(), lastId.toString().getBytes());
        return Response.ok(lastId.toString().getBytes());
    }

    public Response startLuckDraw(Context ctx) {

        String caller = ctx.caller();
        if (caller.isEmpty()) {
            return Response.error("missing initiator");
        }
        String admin = new String(ctx.getObject(admin_key.getBytes()));

        if (caller != admin) {
            return Response.error("only the admin can add new asset type");
        }

        BigDecimal seed = new BigDecimal(new String(ctx.args().get("seed")));
        if (seed.compareTo(BigDecimal.ZERO) == 0) {
            return Response.error("missing seed");
        }

        BigDecimal lastId = new BigDecimal(new String(ctx.getObject(tickets.getBytes())));
        if (lastId.compareTo(BigDecimal.ZERO) == 0) {
            return Response.error("no luck draw tickets");
        }
        Random rander = new Random(seed.longValue()); //TODO truncate

        long luckid = (rander.nextLong() % lastId.longValue()) + 1;
        byte[] luckUser = ctx.getObject((ticketId + String.format("%d", luckid)).getBytes()); // TODO
        ctx.putObject(result.getBytes(), luckUser);
        return Response.ok(luckUser);
    }


    public Response getResult(Context ctx) {
        String luckUser = new String(ctx.getObject(result.getBytes()));
        if (luckUser.isEmpty()) {
            return Response.error("get luck draw result failed");
        }
        return Response.ok(luckUser.getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new LuckDraw());
    }
}