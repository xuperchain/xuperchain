package com.baidu.xuper;

import com.baidu.xuper.contractpb.Contract;
import com.baidu.xuper.contractpb.SyscallGrpc;
import com.google.protobuf.ByteString;
import io.grpc.StatusRuntimeException;

import java.math.BigInteger;
import java.util.List;
import java.util.Map;
import java.util.Iterator;
import java.util.HashMap;
import java.util.Collections;
import java.util.ArrayList;


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
    public List<String> authRequire() {
        List<ByteString> byteStringList = this.callArgs.getAuthRequireList().asByteStringList();
        List<String> authRequires = new ArrayList<>();
        for (ByteString bytes : byteStringList) {
            authRequires.add(bytes.toStringUtf8());
        }

        return authRequires;
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
    }

    @Override
    public Iterator<ContractIteratorItem> newIterator(byte[] start, byte[] limit) {
        return ContractIterator.newIterator(this.client, this.header, start, limit);
    }

    @Override
    public Contract.Transaction queryTx(String txid) {
        Contract.QueryTxRequest request = Contract.QueryTxRequest.newBuilder().setHeader(this.header)
                .setTxid(txid).build();

        Contract.QueryTxResponse resp = this.client.queryTx(request);
        return resp.getTx();
    }

    @Override
    public Contract.Block queryBlock(String blockid) {
        Contract.QueryBlockRequest request = Contract.QueryBlockRequest.newBuilder().setHeader(this.header)
                .setBlockid(blockid).build();

        Contract.QueryBlockResponse resp = this.client.queryBlock(request);
        return resp.getBlock();
    }

    @Override
    public void transfer(String to, BigInteger amount) {
        if (amount.signum() == -1) {
            throw new RuntimeException("amount must not be negative");
        }

        Contract.TransferRequest request = Contract.TransferRequest.newBuilder().setHeader(this.header)
                .setTo(to).setAmount(amount.toString()).build();
        this.client.transfer(request);
    }

    @Override
    public BigInteger transferAmount() {
        BigInteger amount = new BigInteger(this.callArgs.getTransferAmount());
        if (amount.signum() == -1) {
            throw new RuntimeException("amount must not be negative");
        }

        return amount;
    }

    @Override
    public Response call(String module, String contract, String method, Map<String, byte[]> args) {
        Contract.ContractCallRequest.Builder requestBuild = Contract.ContractCallRequest.newBuilder().setHeader(this.header)
                .setModule(module).setContract(contract).setMethod(method);
        int i = 0;
        for (Map.Entry<String, byte[]> entry : args.entrySet()) {
            Contract.ArgPair.Builder argBuilder = requestBuild.addArgsBuilder();
            argBuilder.setKey(entry.getKey());
            argBuilder.setValue(ByteString.copyFrom(entry.getValue()));
            requestBuild.setArgs(i, argBuilder.build());
            i++;
        }
        Contract.ContractCallRequest request = requestBuild.build();
        Contract.ContractCallResponse contractCallResp = this.client.contractCall(request);
        Contract.Response contractResp = contractCallResp.getResponse();
        Response resp = new Response(contractResp.getStatus(), contractResp.getMessage(), contractResp.getBody()
                .toByteArray());
        return resp;
    }

    @Override
    public Response crossQuery(String uri, Map<String, byte[]> args) {
        Contract.CrossContractQueryRequest.Builder requestBuild = Contract.CrossContractQueryRequest.newBuilder()
                .setHeader(this.header).setUri(uri);
        int i = 0;
        for (Map.Entry<String, byte[]> entry : args.entrySet()) {
            Contract.ArgPair.Builder argBuilder = requestBuild.addArgsBuilder();
            argBuilder.setKey(entry.getKey());
            argBuilder.setValue(ByteString.copyFrom(entry.getValue()));
            requestBuild.setArgs(i, argBuilder.build());
            i++;
        }
        Contract.CrossContractQueryRequest request = requestBuild.build();
        Contract.CrossContractQueryResponse crossContractResp = this.client.crossContractQuery(request);
        Contract.Response contractResp = crossContractResp.getResponse();
        Response resp = new Response(contractResp.getStatus(), contractResp.getMessage(), contractResp.getBody().toByteArray());
        return resp;
    }

    @Override
    public void log(String msg) {
        Contract.PostLogRequest request = Contract.PostLogRequest.newBuilder().setHeader(this.header).setEntry(msg)
                .build();
        this.client.postLog(request);
    }
}