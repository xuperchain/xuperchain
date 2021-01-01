package com.baidu.xuper.example;

import java.math.BigDecimal;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

/**
 * Counter
 */
public class SourceTrace implements Contract {
    final String GOODS = "GOODS_";
    final String GOODSRECORD = "GOODSRECORD_";
    final String GOODSRECORDTOP = "GOODSRECORDTOP_";
    final String CREATE = "CREATE_";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        String admin = new String(ctx.args().get("admin".getBytes()));
        if (admin.length() == 0) {
            return Response.error("missing admin address");
        }
        ctx.putObject("admin".getBytes(), admin.getBytes());
        return Response.ok("ok".getBytes());
    }

    private boolean isAdmin(Context ctx, String caller) {
        String admin = new String(ctx.getObject("admin".getBytes()));
        return admin.equals(caller);
    }

    private boolean isAdmin(Context ctx) {
        String caller = ctx.caller();
        return isAdmin(ctx, caller);
    }

    @ContractMethod
    public Response createGoods(Context ctx) {
        if (!isAdmin(ctx)) {
            return Response.error("only the admin can create new goods");
        }
        String id = new String(ctx.args().get("id".getBytes()));
        String desc = new String(ctx.args().get("desc".getBytes()));

        if (id.length() == 0 || desc.length() == 0) {
            return Response.error("missing id or desc");
        }
        String goodsKey = GOODS + id;
        if (ctx.getObject(goodsKey.getBytes()).length > 0) {
            return Response.error("the id  already exist, please check again");
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
        if (!isAdmin(ctx)) {
            return Response.error("only the admin can update goods");
        }
        String id = new String(ctx.args().get("id".getBytes()));
        String reason = new String(ctx.args().get("reason".getBytes()));
        if (id.length() == 0 || reason.length() == 0) {
            return Response.error("missing argument id or argument reason");
        }
        String goodsRecordsTopKey = GOODSRECORDTOP + id;
        BigDecimal topRecord = new BigDecimal(new String(ctx.getObject(goodsRecordsTopKey.getBytes())));
        topRecord = topRecord.add(BigDecimal.ONE);
        String goodsRecordsKey = GOODSRECORD + id + "_" + topRecord.toString();
        ctx.putObject(goodsRecordsKey.getBytes(), reason.getBytes());
        ctx.putObject(goodsRecordsTopKey.getBytes(), topRecord.toString().getBytes());

        return Response.ok(topRecord.toString().getBytes());
    }

    @ContractMethod
    public Response queryRerords(Context ctx) {
        String id = new String(ctx.args().get("id"));
        if (id.length() == 0) {
            return Response.error("missing argument id");
        }
        String goodsKey = GOODS + id;
        if (ctx.getObject(goodsKey.getBytes()).length == 0) {
            return Response.error("good with id " + id + "not found");
        }
        String goodsRecordKey = GOODSRECORD + id + "_";
        String start = goodsRecordKey;
        String end = start + "~";
        StringBuffer buf = new StringBuffer();
        buf.append("\n");

        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                elem -> {
                    String key = elem.getKey().toString();
                    String[] goodsRecords = key.substring(GOODSRECORD.length()).split("_");
                    String goodsId = goodsRecords[0];
                    String updateRecord = goodsRecords[1];
                    String reason = new String(elem.getValue());
                    String record = "goodIds=" + goodsId + ",updateRecord=" + updateRecord + ",reason=" + reason + "\n";
                    buf.append(record.getBytes());
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new SourceTrace());
        System.out.println(BigDecimal.ZERO.toPlainString().getBytes() == "0".getBytes());
    }
}