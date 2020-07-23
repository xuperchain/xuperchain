package com.baidu.xuper;

import java.math.BigInteger;
import java.util.HashMap;

/**
 * c1
 */
public class c1 implements Contract {
    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("initialize success".getBytes());
    }

    @ContractMethod
    public Response invoke(Context ctx) {
        String cntKey = "cnt";
        BigInteger counter;
        byte[] value = ctx.getObject(cntKey.getBytes());
        if (value == null) {
            counter = BigInteger.valueOf(0);
        } else {
            counter = new BigInteger(new String(value));
        }

        // 发起转账
        final byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        final String to = new String(toByte);

        ctx.transfer(to, BigInteger.valueOf(1));

        final HashMap<String, byte[]> callArgs =
                new HashMap<String, byte[]>() {
                    {
                        put("to", toByte);
                    }
                };

        // 发起跨合约调用
        Response resp = ctx.call("native", "callc2", "invoke", callArgs);
        // 根据合约调用结果记录到call变量里面并持久化
        ctx.putObject("call".getBytes(), resp.body);
        // 对cnt变量加1并持久化
        counter = counter.add(BigInteger.valueOf(1));
        ctx.putObject(cntKey.getBytes(), counter.toString().getBytes());

        return Response.ok("ok".getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new c1());
    }
}
