package com.baidu.xuper;

import java.io.UnsupportedEncodingException;
import java.util.Map;
import java.util.HashMap;
import java.util.ArrayList;
import java.util.List;
import java.util.Collections;
import io.grpc.StatusRuntimeException;

import com.baidu.xuper.contractpb.Contract;
import com.baidu.xuper.contractpb.SyscallGrpc;
import com.google.protobuf.ByteString;


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

    @Override
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
    public ArrayList<String> authRequire() {
        List<ByteString> byteStringList =  this.callArgs.getAuthRequireList().asByteStringList();
        ArrayList<String> authRequires = new ArrayList();
        for(int i = 0;i < byteStringList.size();i++){
            authRequires.add(byteStringList.get(i).toString());
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
        return;
    }

//    @Override
    public Contract.IteratorResponse newIterator(ByteString start, ByteString limit) {
        Contract.IteratorRequest request = Contract.IteratorRequest.newBuilder().setHeader(this.header)
                .setStart(start).setLimit(limit).build();
        try{
            Contract.IteratorResponse resp = this.client.newIterator(request);
            return resp;
        } catch (StatusRuntimeException e){
            return null;
        }
    }

    @Override
    public Contract.Transaction queryTx(String txid) {
        Contract.QueryTxRequest request = Contract.QueryTxRequest.newBuilder().setHeader(this.header)
                .setTxid(txid).build();

        try{
            Contract.QueryTxResponse resp = this.client.queryTx(request);
            return resp.getTx();
        } catch (StatusRuntimeException e){
            return null;
        }
    }

    @Override
    public Contract.Block queryBlock(String blockid) {
        Contract.QueryBlockRequest request = Contract.QueryBlockRequest.newBuilder().setHeader(this.header)
                .setBlockid(blockid).build();

        try{
            Contract.QueryBlockResponse resp = this.client.queryBlock(request);
            return resp.getBlock();
        } catch (StatusRuntimeException e){
            return null;
        }
    }

    @Override
    public void transfer(String to, String amount) {
        Contract.TransferRequest request = Contract.TransferRequest.newBuilder().setHeader(this.header)
                .setTo(to).setAmount(amount).build();
        this.client.transfer(request);
        return;
    }

    @Override
    public Response call(String module, String contract, String method, HashMap<String,String> args) {
        Contract.ContractCallRequest.Builder requestBuild = Contract.ContractCallRequest.newBuilder().setHeader(this.header)
                .setModule(module).setContract(contract).setMethod(method);
        int i = 0;
        for (Map.Entry<String, String> entry : args.entrySet()){
            Contract.ArgPair.Builder argBuilder = requestBuild.addArgsBuilder();
            argBuilder.setKey(entry.getKey());
            ByteString value;
            try{
                value = ByteString.copyFrom(entry.getValue(),"UTF-8");
            } catch(UnsupportedEncodingException e){
                return null;
            }
            argBuilder.setValue(value);
            requestBuild.setArgs(i,argBuilder.build());
            i++;
        }
        Contract.ContractCallRequest request = requestBuild.build();
        try {
            Contract.ContractCallResponse contractCallResp = this.client.contractCall(request);
            Contract.Response contractResp = contractCallResp.getResponse();
            Response resp = new Response(contractResp.getStatus(),contractResp.getMessage(),contractResp.getBody()
                    .toByteArray());
            return resp;
        } catch(StatusRuntimeException e){
            return null;
        }
    }

    @Override
    public Response crossQuery(String uri, HashMap<String,String> args) {
        Contract.CrossContractQueryRequest.Builder requestBuild = Contract.CrossContractQueryRequest.newBuilder()
                .setHeader(this.header).setUri(uri);
        int i = 0;
        for (Map.Entry<String, String> entry : args.entrySet()){
            Contract.ArgPair.Builder argBuilder = requestBuild.addArgsBuilder();
            argBuilder.setKey(entry.getKey());
            ByteString value;
            try{
                value = ByteString.copyFrom(entry.getValue(),"UTF-8");
            } catch(UnsupportedEncodingException e){
                return null;
            }
            argBuilder.setValue(value);
            requestBuild.setArgs(i,argBuilder.build());
            i++;
        }
        Contract.CrossContractQueryRequest request = requestBuild.build();
        try {
            Contract.CrossContractQueryResponse crossContractResp = this.client.crossContractQuery(request);
            Contract.Response contractResp = crossContractResp.getResponse();
            Response resp = new Response(contractResp.getStatus(),contractResp.getMessage(),contractResp.getBody().toByteArray());
            return resp;
        } catch(StatusRuntimeException e){
            return null;
        }
    }

    @Override
    public void log(String msg) {
        Contract.PostLogRequest request = Contract.PostLogRequest.newBuilder().setHeader(this.header).setEntry(msg)
                .build();
        this.client.postLog(request);
        return;
    }
}