package com.baidu.xuper;

import java.net.InetSocketAddress;

import io.grpc.Server;
import io.grpc.netty.shaded.io.grpc.netty.NettyServerBuilder;

/**
 * Driver
 */
public class Driver {
    final private static String XCHAIN_CHAIN_ADDR = "XCHAIN_CHAIN_ADDR";
    final private static String XCHAIN_CODE_PORT = "XCHAIN_CODE_PORT";

    public static void serve(Contract contract) {
        try {
            String chainAddr = System.getenv(XCHAIN_CHAIN_ADDR);
            NativeCodeService codeService = NativeCodeService.newNativeCodeService(chainAddr, contract);
            int codePort = Integer.parseInt(System.getenv(XCHAIN_CODE_PORT));
            Server server = NettyServerBuilder.forAddress(new InetSocketAddress("127.0.0.1", codePort))
                    .addService(codeService).build();
            server.start();
            waitAndKeepAlive(codeService);
            server.shutdown();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static void waitAndKeepAlive(NativeCodeService codeService) {
        for (;;) {
            try {
                codeService.pingXchain();
                Thread.sleep(1000);
            } catch (Exception e) {
                System.out.println("ping xchain node error " + e.toString());
                return;
            }
        }
    }
}