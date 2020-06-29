package com.baidu.xuper;

import java.util.HashMap;
/**
 * Cross Query Demo
 *
 */
public class CrossQueryDemo implements Contract
{
    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("initialize success".getBytes());
    }

    @ContractMethod
    public Response crossQuery(Context ctx) {

        String uri = "xuper://test.xuper?module=native&bcname=xuper&contract_name=counter&method_name=get";

        final HashMap<String, String> callArgs =
                new HashMap<String, String>() {
                    {
                        put("key", "test");
                    }
                };

        Response resp = ctx.crossQuery(uri, callArgs);

        return Response.ok(resp.body);
    }

    public static void main(String[] args) {
        Driver.serve(new CrossQueryDemo());
    }
}
