package com.baidu.xuper.example;

import com.baidu.xuper.*;

public class ShortContent implements Contract {
    final private String userBucket = "USER";
    final private int titleLengthLimit = 100;
    final private int contentLengthLimit = 3000;

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response storeShortContent(Context ctx) {
        if (ctx.args().get("user_id") == null ||ctx.args().get("user_id").length==0||
                ctx.args().get("title") == null ||ctx.args().get("title").length==0||
                ctx.args().get("topic") == null || ctx.args().get("topic").length==0||
                ctx.args().get("content") == null||ctx.args().get("content").length==0) {
            return Response.error("missing user_id or title of topic or content");
        }
        String userId = new String(ctx.args().get("user_id"));
        String title = new String(ctx.args().get("title"));
        String topic = new String(ctx.args().get("topic"));
        String content = new String(ctx.args().get("content"));
        String userKey = userBucket + "/" + userId + "/" + topic + "/" + title;

        if (topic.length() > contentLengthLimit || title.length() > titleLengthLimit || content.length() > contentLengthLimit) {
            return Response.error("The length of topic or title or content is more than limitation");
        }

        ctx.putObject(userKey.getBytes(), content.getBytes());
        return Response.ok("ok~".getBytes());
    }

    @ContractMethod
    public Response queryByUser(Context ctx) {
        if (ctx.args().get("user_id") ==null||ctx.args().get("user_id").length==0){
            return  Response.error("missing user_id");
        }
        String userId = new String(ctx.args().get("user_id"));
        StringBuffer buf = new StringBuffer();

        String start = userBucket + "/" + userId + "/";
        String end = start + "~";

        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                item -> {
                    String []fields = new String(item.getKey()).split("/");
                    buf.append(fields[1])
                            .append("\t")
                            .append(fields[2])
                            .append("\t")
                            .append(fields[3])
                            .append("\t")
                            .append(new String(item.getValue()))
                            .append("\n");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response queryByTitle(Context ctx) {
        if (ctx.args().get("user_id") ==null||
        ctx.args().get("title")==null||
        ctx.args().get("topic")==null){
            return Response.error("missing user_id of title or topic");
        }
        String userId = new String(ctx.args().get("user_id"));
        String title = new String(ctx.args().get("title"));
        String topic = new String(ctx.args().get("topic"));
        String key = userBucket + "/" + userId + "/"  + topic + "/" + title;
        byte[] value = ctx.getObject(key.getBytes());
        if (value==null){
            return Response.error("content not found");
        }
        String[] fields = new String(value).split("/");
        return Response.ok(value);
    }

    @ContractMethod
    public Response queryByTopic(Context ctx) {
        String userId = new String(ctx.args().get("user_id"));
        String topic = new String(ctx.args().get("topic"));

        String key = String.join("/", new String[]{userBucket, userId, topic});
        String start = key;
        String end = key + "~";

        StringBuffer buf = new StringBuffer();
        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                item -> {
                    String []fields = new String(item.getKey()).split("/");
                    buf.append(fields[1])
                            .append("\t")
                            .append(fields[2])
                            .append("\t")
                            .append(fields[3])
                            .append("\t")
                            .append(new String(item.getValue()))
                            .append("\n");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new ShortContent());
    }
}