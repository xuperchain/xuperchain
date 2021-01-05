package com.baidu.xuper;

import java.util.Iterator;

/**
 * Counter
 */
public class HashDeposit implements Contract {
    final String USER_BUCKET = "USER";
    final String HASH_BUCKET = "HASH";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response storeFileInfo(Context ctx) {
        if (ctx.args().get("user_id") == null || ctx.args().get("user_id").length == 0 ||
                ctx.args().get("hash_id") == null || ctx.args().get("hash_id").length == 0 ||
                ctx.args().get("file_name") == null || ctx.args().get("file_name").length == 0
        ) {
            return Response.error("missing user_id or hash_id or filename");
        }
        String user_id = new String(ctx.args().get("user_id"));
        String hash_id = new String(ctx.args().get("hash_id"));
        String file_name = new String(ctx.args().get("file_name"));

        String userKey = USER_BUCKET + "/" + user_id + "/" + hash_id;
        String hashKey = HASH_BUCKET + "/" + hash_id;
        String value = user_id + "\t" + hash_id + "\t" + file_name;
        if (ctx.getObject(hashKey.getBytes()) != null) {
            return Response.error("hashid " + hash_id + " already exists");
        }
        ctx.putObject(userKey.getBytes(), value.getBytes());
        ctx.putObject(hashKey.getBytes(), value.getBytes());
        return Response.ok("".getBytes());
    }

    @ContractMethod
    public Response queryUserList(Context ctx) {
        String key = USER_BUCKET + "/";
        String start = key;
        String end = key + "~";
        Iterator<ContractIteratorItem> iter = ctx.newIterator(start.getBytes(), end.getBytes());
        StringBuffer buf = new StringBuffer();
        iter.forEachRemaining(
                item -> {
                    buf.append(new String(item.getKey()).split("/")[1]);
                    buf.append("\t");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response queryFileInfoByUser(Context ctx) {
        String key = USER_BUCKET + "/" + new String(ctx.args().get("user_id"));
        String start = key;
        String end = start + "~";
        StringBuffer buf = new StringBuffer();
        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                item -> {
                    buf.append(new String(item.getValue()));
                    buf.append("\n");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response queryFileInfoByHash(Context ctx) {
        byte[] hashId = ctx.args().get("hash_id");
        if (hashId == null || hashId.length == 0) {
            return Response.error("missing hash_id");
        }
        String key = HASH_BUCKET + "/" + new String(hashId);
        byte[] info = ctx.getObject(key.getBytes());
        if (info == null) {
            return Response.error("file info not exists");
        }
        return Response.ok(info);
    }

    public static void main(String[] args) {
        Driver.serve(new HashDeposit());
    }
}