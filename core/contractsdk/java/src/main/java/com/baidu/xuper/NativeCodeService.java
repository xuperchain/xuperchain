package com.baidu.xuper;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.net.URI;

import com.baidu.xuper.contractpb.Contract;
import com.baidu.xuper.contractpb.NativeCodeGrpc;
import com.baidu.xuper.contractpb.SyscallGrpc;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;

/**
 * NativeCodeService
 */
class NativeCodeService extends NativeCodeGrpc.NativeCodeImplBase {
    private final ManagedChannel channel;
    private final SyscallGrpc.SyscallBlockingStub client;
    private final com.baidu.xuper.Contract contract;

    public static NativeCodeService newNativeCodeService(String chainAddr, com.baidu.xuper.Contract contract)
            throws Exception {
        URI uri = new URI(chainAddr);
        switch (uri.getScheme()) {
            case "tcp":
                break;
            default:
                throw new Exception("unsupported protocol " + uri.getScheme());
        }
        String target = uri.getHost() + ":" + String.valueOf(uri.getPort());
        ManagedChannel channel = ManagedChannelBuilder.forTarget(target).usePlaintext().build();
        return new NativeCodeService(channel, contract);
    }

    public NativeCodeService(ManagedChannel channel, com.baidu.xuper.Contract contract) {
        this.channel = channel;
        this.client = SyscallGrpc.newBlockingStub(this.channel);
        this.contract = contract;
    }

    @Override
    public void ping(Contract.PingRequest request, StreamObserver<Contract.PingResponse> responseObserver) {
        Contract.PingResponse resp = Contract.PingResponse.newBuilder().build();
        responseObserver.onNext(resp);
        responseObserver.onCompleted();
    }

    @Override
    public void call(Contract.NativeCallRequest request, StreamObserver<Contract.NativeCallResponse> responseObserver) {
        ContextImpl ctx = ContextImpl.newContext(this.client, request.getCtxid());
        String methodName = ctx.getMethodName();
        Response resp = callMethod(this.contract, methodName, ctx);
        ctx.setOutput(resp);
        Contract.NativeCallResponse callResp = Contract.NativeCallResponse.newBuilder().build();
        responseObserver.onNext(callResp);
        responseObserver.onCompleted();
    }

    private Response callMethod(com.baidu.xuper.Contract contract, String methodName, Context ctx) {
        try {
            if (methodName.equals("initialize")) {
                return contract.initialize(ctx);
            }
            Class cls = contract.getClass();
            Method method = cls.getMethod(methodName, Context.class);
            if (!method.isAnnotationPresent(ContractMethod.class)) {
                return new Response(400, "method not marked as contract method", null);
            }
            return (Response) method.invoke(this.contract, ctx);
        } catch (NoSuchMethodException e) {
            return new Response(400, "method not found " + methodName, null);
        } catch (InvocationTargetException e) {
            return new Response(500, e.getTargetException().toString(), null);
        } catch (Exception e) {
            return new Response(500, "call method exception " + e.toString(), null);
        }
    }

    public void pingXchain() {
        Contract.PingRequest request = Contract.PingRequest.newBuilder().build();
        this.client.ping(request);
    }
}