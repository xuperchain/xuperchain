package com.baidu.xuper;

/**
 * Response
 */
public class Response {
    public int status;
    public String message;
    public byte[] body;

    public Response(int status, String msg, byte[] body) {
        this.status = status;
        this.message = msg;
        this.body = body;
    }

    public static Response ok(byte[] body) {
        return new Response(200, "", body);
    }

    public static Response error(String msg) {
        return new Response(500, msg, null);
    }
}