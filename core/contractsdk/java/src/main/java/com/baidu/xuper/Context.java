package com.baidu.xuper;

import java.math.BigInteger;
import java.util.List;
import java.util.Map;

import com.baidu.xuper.contractpb.Contract;

public interface Context {
    public Map<String, byte[]> args();

    public String caller();

    public List<String> authRequire();

    public void putObject(byte[] key, byte[] value);

    public byte[] getObject(byte[] key);

    public void deleteObject(byte[] key);

//    public Contract.IteratorResponse newIterator(byte[] start, byte[] limit) throws Exception;

    public Contract.Transaction queryTx(String txid) throws Exception;

    public Contract.Block queryBlock(String blockid) throws Exception;

    public void transfer(String to, BigInteger amount) throws Exception;

    public Response call(String module, String contract, String method, Map<String,byte[]> args) throws Exception;

    public Response crossQuery(String uri, Map<String,byte[]> args) throws Exception;

    public void setOutput(Response resp);

    public void log(String msg);

}