package com.baidu.xuper;

import com.baidu.xuper.contractpb.Contract;

import java.math.BigInteger;
import java.util.Iterator;
import java.util.List;
import java.util.Map;

public interface Context {
    public Map<String, byte[]> args();

    public String caller();

    public List<String> authRequire();

    public void putObject(byte[] key, byte[] value);

    public byte[] getObject(byte[] key);

    public void deleteObject(byte[] key);

    public Iterator<ContractIteratorItem> newIterator(byte[] start, byte[] limit);

    public Contract.Transaction queryTx(String txid);

    public Contract.Block queryBlock(String blockid);

    public void transfer(String to, BigInteger amount);

    public BigInteger transferAmount();

    public Response call(String module, String contract, String method, Map<String, byte[]> args);

    public Response crossQuery(String uri, Map<String, byte[]> args);

    public void emitEvent(String name, byte[] body);

    public void emitJSONEvent(String name, Map<String, byte[]> body);

    public void log(String msg);
}