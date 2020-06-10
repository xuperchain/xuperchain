package com.baidu.xuper;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

import com.baidu.xuper.contractpb.Contract;
import com.baidu.xuper.contractpb.SyscallGrpc;
import com.google.protobuf.ByteString;

import io.grpc.StatusRuntimeException;

/**
 * ContextImpl implements the Context interface and provides an interface to
 * access the XuperBridge for a single contract call
 */
class ContextImpl implements Context {

    private final SyscallGrpc.SyscallBlockingStub client;
    private final Contract.SyscallHeader header;
    private Map<String, byte[]> args;
    private Contract.CallArgs callArgs;

    public static ContextImpl newContext(SyscallGrpc.SyscallBlockingStub client, long ctxid) {
        ContextImpl impl = new ContextImpl(client, ctxid);
        impl.init();
        return impl;
    }

    private ContextImpl(SyscallGrpc.SyscallBlockingStub client, long ctxid) {
        this.client = client;
        this.header = Contract.SyscallHeader.newBuilder().setCtxid(ctxid).build();
    }

    private void init() {
        Contract.GetCallArgsRequest request = Contract.GetCallArgsRequest.newBuilder().setHeader(this.header).build();
        this.callArgs = client.getCallArgs(request);
        this.args = new HashMap<String, byte[]>();
        for (int i = 0; i < this.callArgs.getArgsCount(); i++) {
            Contract.ArgPair pair = this.callArgs.getArgs(i);
            args.put(pair.getKey(), pair.getValue().toByteArray());
        }
    }

    public void setOutput(Response resp) {
        Contract.Response.Builder respBuilder = Contract.Response.newBuilder();
        respBuilder.setStatus(resp.status);
        if (resp.message != null) {
            respBuilder.setMessage(resp.message);
        }
        if (resp.body != null) {
            respBuilder.setBody(ByteString.copyFrom(resp.body));
        }
        Contract.Response outResp = respBuilder.build();
        Contract.SetOutputRequest request = Contract.SetOutputRequest.newBuilder().setHeader(this.header)
                .setResponse(outResp).build();
        this.client.setOutput(request);
    }

    public String getMethodName() {
        return this.callArgs.getMethod();
    }

    @Override
    public Map<String, byte[]> args() {
        return Collections.unmodifiableMap(this.args);
    }

    @Override
    public String caller() {
        return this.callArgs.getInitiator();
    }

    @Override
    public void putObject(byte[] key, byte[] value) {
        Contract.PutRequest request = Contract.PutRequest.newBuilder().setHeader(this.header)
                .setKey(ByteString.copyFrom(key)).setValue(ByteString.copyFrom(value)).build();
        this.client.putObject(request);
    }

    @Override
    public byte[] getObject(byte[] key) {
        Contract.GetRequest request = Contract.GetRequest.newBuilder().setHeader(this.header)
                .setKey(ByteString.copyFrom(key)).build();

        try {
            Contract.GetResponse resp = this.client.getObject(request);
            return resp.getValue().toByteArray();
        } catch (StatusRuntimeException e) {
            return null;
        }
    }

    @Override
    public void deleteObject(byte[] key) {
        Contract.DeleteRequest request = Contract.DeleteRequest.newBuilder().setHeader(this.header)
                .setKey(ByteString.copyFrom(key)).build();
        this.client.deleteObject(request);
        return;
    }

    @Override
    public void log(String msg) {
        Contract.PostLogRequest request = Contract.PostLogRequest.newBuilder().setHeader(this.header).setEntry(msg)
                .build();
        this.client.postLog(request);
        return;
    }
}