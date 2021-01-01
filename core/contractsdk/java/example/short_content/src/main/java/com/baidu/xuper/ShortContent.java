package com.baidu.xuper;

/**
 * ShortContent
 */
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
        String userId =new String( ctx.args().get("user_id"));
        String  title = new String(ctx.args().get("title"));
        String topic =  new String( ctx.args().get("topic"));
        String content = new String(ctx.args().get("content"));
        String userKey = userBucket+"/"+userId+"/"+topic+"/"+title;

        if( topic.length()  > contentLengthLimit||title.length() > titleLengthLimit || content.length() > contentLengthLimit){
            return Response.error("The length of topic or title or content is more than limitation");
        }

        ctx.putObject(userKey.getBytes(),content.getBytes());
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response queryByUser(Context ctx){
        String userId =new String( ctx.args().get("user_id"));
        StringBuffer buf = new StringBuffer();

        String start = userBucket + "/"  + userId + "/";
        String end = start + "~";

        ctx.newIterator(start.getBytes(),end.getBytes()).forEachRemaining(
                item ->{
                    buf.append(new String(item.getKey()))
                            .append("\n")
                            .append(new String(item.getValue()))
                            .append("\n");
                }
        );
        return Response.ok("ok".getBytes());
    }

    @ContractMethod
    public Response queryByTitle(Context ctx){
        String userId = new String(ctx.args().get("user_id"));
        String  title = new String(ctx.args().get("title"));
        String topic =  new String(ctx.args().get("topic"));
        String key = userBucket + "/" + userId + "/" + "/" + topic + "/" + title;
        return Response.ok(ctx.getObject(key.getBytes()));
    }

    @ContractMethod
    public Response queryByTopic(Context ctx){
        String userId =new String(ctx.args().get("user_id"));
        String  topic =  new String(ctx.args().get("topic"));

        String key = String.join("/",new String[]{userBucket,userId,topic});
        String start = key;
        String end = key+"~";

        StringBuffer buf = new StringBuffer();
        ctx.newIterator(start.getBytes(),end.getBytes()).forEachRemaining(
                item ->{
                    buf.append(new String(item.getKey()))
                            .append("\n")
                            .append(new String(item.getKey()))
                            .append("\n");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new ShortContent());
    }
}