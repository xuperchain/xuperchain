package com.baidu.xuper;


/**
 * Counter
 */
public class ScoreRecord implements Contract {
    private final String OWNER_KEY = "owner";
    private final String RECORD_KEY = "R_";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        byte[] owner = ctx.args().get("owner");
        if (owner == null || owner.length == 0) {
            return Response.error("missing caller");
        }
        ctx.putObject(OWNER_KEY.getBytes(), owner);
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response addScore(Context ctx) {
        String caller = ctx.caller();
        if (caller == null || caller.isEmpty()) {
            return Response.error("missing caller");
        }
        byte[] userIdByte = ctx.args().get("user_id");
        byte[] dataByte = ctx.args().get("data");
        if (userIdByte == null || userIdByte.length == 0) {
            return Response.error("missing user_id");
        }
        if (dataByte == null || dataByte.length == 0) {
            return Response.error("missing data");
        }
        if (!new String(ctx.getObject(OWNER_KEY.getBytes())).equals(caller)) {
            return Response.error("you do not have permission to call this method");
        }
        ctx.putObject((RECORD_KEY + new String(userIdByte)).getBytes(), dataByte);
        return Response.ok(userIdByte);
    }

    @ContractMethod
    public Response queryScore(Context ctx) {
        byte[] userIdByte = ctx.args().get("user_id");
        if (userIdByte == null || userIdByte.length == 0) {
            return Response.error("missing user_id");
        }
        byte[] data = ctx.getObject((RECORD_KEY + new String(userIdByte)).getBytes());
        if (data == null) {
            return Response.error("record of " + new String(userIdByte) + " not found'");
        }
        return Response.ok(data);
    }

    @ContractMethod
    public Response queryOwner(Context ctx) {
        return Response.ok(ctx.getObject(OWNER_KEY.getBytes()));
    }

    public static void main(String[] args) {
        Driver.serve(new ScoreRecord());
    }
}
