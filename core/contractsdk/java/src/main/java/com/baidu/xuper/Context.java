package com.baidu.xuper;

import java.util.Map;

public interface Context {
    public Map<String, byte[]> args();

    public String caller();

    public void putObject(byte[] key, byte[] value);

    public byte[] getObject(byte[] key);

    public void deleteObject(byte[] key);

    public void log(String msg);

}