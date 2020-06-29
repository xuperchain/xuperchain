package com.baidu.xuper;

import java.util.HashMap;
import java.util.ArrayList;
import java.util.Map;

import com.baidu.xuper.contractpb.Contract;

public interface Context {
    public Map<String, byte[]> args();

    public String caller();

    public ArrayList<String> authRequire();

    public void putObject(byte[] key, byte[] value);

    public byte[] getObject(byte[] key);

    public void deleteObject(byte[] key);

//    public Contract.IteratorResponse newIterator(ByteString start, ByteString limit);

    public Contract.Transaction queryTx(String txid);

    public Contract.Block queryBlock(String blockid);

    public void transfer(String to, String amount);

    public Response call(String module, String contract, String method, HashMap<String,String> args);

    public Response crossQuery(String uri, HashMap<String,String> args);

    public void setOutput(Response resp);

    public void log(String msg);

}