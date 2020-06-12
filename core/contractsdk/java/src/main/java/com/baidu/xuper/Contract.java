package com.baidu.xuper;

/**
 * Contract stipulates the method that contract needs to implement. All
 * contracts require an initialize method, and other methods need to be
 * annotated through ContractMethod
 */
public interface Contract {
    public Response initialize(Context ctx);
}