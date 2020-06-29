package com.baidu.xuper;

import java.math.BigInteger;

/**
 * c2
 *
 */
public class c2 implements Contract
{
    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("initialize success".getBytes());
    }

    @ContractMethod
    public Response invoke(final Context ctx) {
        String cntKey = "cnt";
        BigInteger counter;
        byte[] value = ctx.getObject(cntKey.getBytes());
        if (value == null) {
            counter = BigInteger.valueOf(0);
        } else {
            counter = new BigInteger(value);
        }

        counter = counter.add(BigInteger.valueOf(1000));

        byte[] toByte = ctx.args().get("to");
        if (toByte == null) {
            return Response.error("missing to");
        }
        final String to = new String(toByte);
        ctx.transfer(to,"1000");

        ctx.putObject(cntKey.getBytes(), counter.toByteArray());

        return Response.ok(counter.toByteArray());
    }


    public static void main(String[] args) {
        Driver.serve(new c2());
    }
}
