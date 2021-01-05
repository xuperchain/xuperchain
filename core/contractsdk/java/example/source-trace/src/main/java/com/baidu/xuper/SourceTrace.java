package com.baidu.xuper;

import java.math.BigDecimal;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

public class SourceTrace implements Contract {
    final String GOODS = "GOODS_";
    final String GOODSRECORD = "GOODSRECORD_";
    final String GOODSRECORDTOP = "GOODSRECORDTOP_";
    final String CREATE = "CREATE";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        byte[] adminByte = ctx.args().get("admin");
        if (adminByte == null || adminByte.length == 0) {
            return Response.error("missing admin");
        }

        ctx.putObject("admin".getBytes(), adminByte);
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response createGoods(Context ctx) {
        String caller = ctx.caller();
        if (caller == null || caller.length() == 0) {
            return Response.error("missing caller");
        }
        byte[] admin = ctx.getObject("admin".getBytes());
        if (!new String(admin).equals(caller)) {
            return Response.error("you do not have permission to call this method");
        }

        if (ctx.args().get("id") == null || ctx.args().get("id").length == 0 || ctx.args().get("desc") == null
                || ctx.args().get("desc").length == 0) {
            return Response.error("missing id or desc");
        }

        String id = new String(ctx.args().get("id"));
        String desc = new String(ctx.args().get("desc"));
        String goodsKey = GOODS + id;
        if (ctx.getObject(goodsKey.getBytes()) != null && ctx.getObject(goodsKey.getBytes()).length != 0) {
            return Response.error("goods type " + id + " already exists");
        }

        ctx.putObject(goodsKey.getBytes(), desc.getBytes());
        String goodsRecordsKey = GOODSRECORD + id + "_0";
        String goodsRecordsTopKey = GOODSRECORDTOP + id;
        ctx.putObject(goodsRecordsKey.getBytes(), CREATE.getBytes());
        ctx.putObject(goodsRecordsTopKey.getBytes(), "0".getBytes());
        return Response.ok(id.getBytes());
    }

    @ContractMethod
    public Response updateGoods(Context ctx) {
        String caller = ctx.caller();
        if (caller == null || caller.length() == 0) {
            return Response.error("missing caller");
        }
        byte[] admin = ctx.getObject("admin".getBytes());
        if (!new String(admin).equals(caller)) {
            return Response.error("you do not have permission to call this method");
        }

        if (ctx.args().get("id") == null || ctx.args().get("id").length == 0 || ctx.args().get("reason") == null
                || ctx.args().get("reason").length == 0) {
            return Response.error("missing argument id or argument reason");
        }

        String id = new String(ctx.args().get("id"));
        String reason = new String(ctx.args().get("reason"));

        String goodsRecordsTopKey = GOODSRECORDTOP + id;
        byte[] topRecordValue = ctx.getObject(goodsRecordsTopKey.getBytes());
        if (topRecordValue == null || topRecordValue.length == 0) {
            return Response.error("goods " + id + " not found");
        }
        BigDecimal topRecord = new BigDecimal(new String(topRecordValue));
        topRecord = topRecord.add(BigDecimal.ONE);
        String goodsRecordsKey = GOODSRECORD + id + "_" + topRecord.toString();
        ctx.putObject(goodsRecordsKey.getBytes(), reason.getBytes());
        ctx.putObject(goodsRecordsTopKey.getBytes(), topRecord.toString().getBytes());

        return Response.ok(topRecord.toString().getBytes());
    }

    @ContractMethod
    public Response queryRecords(Context ctx) {
        if (ctx.args().get("id") == null || ctx.args().get("id").length == 0) {
            return Response.error("missing id");

        }
        String id = new String(ctx.args().get("id"));

        String goodsKey = GOODS + id;
        if (ctx.getObject(goodsKey.getBytes()) == null || ctx.getObject(goodsKey.getBytes()).length == 0) {
            return Response.error("good with id " + id + "not found");
        }
        String goodsRecordKey = GOODSRECORD + id + "_";
        String start = goodsRecordKey;
        String end = start + "~";
        StringBuffer buf = new StringBuffer();

        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(elem -> {
            String key = new String(elem.getKey());
            String[] goodsRecords = key.split("_");
            String goodsId = goodsRecords[1];
            String updateRecord = goodsRecords[2];
            String reason = new String(elem.getValue());
            String record ="updateRecord=" + updateRecord + ",reason=" + reason+"\n";
            buf.append(record);
        });
        return Response.ok(buf.toString().getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new SourceTrace());
    }
}
