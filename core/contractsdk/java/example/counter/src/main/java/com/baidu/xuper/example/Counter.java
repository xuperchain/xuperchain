package com.baidu.xuper.example;

import java.math.BigInteger;
import java.util.HashMap;
import java.util.Map;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

/**
 * Counter
 */
public class Counter implements Contract {

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response increase(Context ctx) {
        byte[] key = ctx.args().get("key");
        if (key == null) {
            return Response.error("missing key");
        }
        BigInteger counter;
        byte[] value = ctx.getObject(key);
        if (value != null) {
            counter = new BigInteger(value);
        } else {
            ctx.log("key " + new String(key) + " not found, initialize to zero");
            counter = BigInteger.valueOf(0);
        }
        ctx.log("get value " + counter.toString());
        counter = counter.add(BigInteger.valueOf(1));
        ctx.putObject(key, counter.toByteArray());

//        emit event with json
        Map<String, String> body = new HashMap<>();
        body.put("key", new String(key));
        body.put("value", counter.toString());
        ctx.emitJSONEvent("increase", body);

//        emit event directly, use 'toString().getBytes()' to make counter readable to human
//        ctx.emitEvent("increase", counter.toString().getBytes());

        return Response.ok(counter.toString().getBytes());
    }

    @ContractMethod
    public Response get(Context ctx) {
        byte[] key = ctx.args().get("key");
        if (key == null) {
            return Response.error("missing key");
        }
        BigInteger counter;
        byte[] value = ctx.getObject(key);
        if (value != null) {
            counter = new BigInteger(value);
        } else {
            return Response.error("key " + new String(key) + " not found)");
        }
        ctx.log("get value " + counter.toString());

        return Response.ok(counter.toString().getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new Counter());
    }
}