package com.baidu.xuper.example;


import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

/**
 * Counter
 */
public class GameAssets implements Contract {
    final String ASSETTYPE = "AssetType_";
    final String USERASSET = "UserAsset";
    final String ASSET2USER = "Asset2User_";


    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        byte[] admin = ctx.args().get("admin");
        if (admin.length == 0) {
            return Response.error("missing admin address");
        }
        ctx.putObject("admin".getBytes(), admin);
        return Response.ok("ok".getBytes());
    }

    private boolean isAdmin(Context ctx) {
        return this.isAdmin(ctx, ctx.caller());
    }

    private boolean isAdmin(Context ctx, String caller) {
        String admin = new String(ctx.getObject("admin".getBytes()));
        return admin.equals(caller);
    }

    @ContractMethod
    public Response AddAssetType(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("only admin can add new asset type");
        }
        String typeId = new String(ctx.args().get("typeid"));
        String typeDesc = new String(ctx.args().get("typedesc"));
        if (typeId.length() == 0 || typeDesc.length() == 0) {
            return Response.error("missing typeid or typedesc");
        }
        String assetTypeKey = ASSETTYPE + typeId;
        if (ctx.getObject(assetTypeKey.getBytes()).length != 0) {
            return Response.error("type id does already exist");
        }
        ctx.putObject(assetTypeKey.getBytes(), typeDesc.getBytes());
        return Response.ok(typeId.getBytes());
    }

    @ContractMethod
    public Response listAssetType(Context ctx) {
        StringBuffer buf = new StringBuffer();
        ctx.newIterator(ASSETTYPE.getBytes(), (ASSETTYPE + "~").getBytes()).forEachRemaining(
                elem -> {
                    String assetType = new String((elem.getKey()));
                    buf.append((assetType.substring(ASSETTYPE.length()) + ":" +
                            new String(elem.getValue()) +
                            "\n"));
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response getAssetsByUser(Context ctx) {
        //        admin cat get asset of any user
        String userId = ctx.caller();
        if (isAdmin(ctx, userId)) {
            if (ctx.args().get("userid").length != 0) { // TODO 参数格式
                userId = new String(ctx.args().get("user_id"));
            }
        }
        String userAsstKey = USERASSET + userId + "_";
        String start = userAsstKey;
        String end = start + "~";
        StringBuffer buf = new StringBuffer();
        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                elem -> {
                        String assetId = new String(elem.getKey()).substring(userAsstKey.length());
                        String typeId = new String (elem.getValue());
                        String assetTypeKey = ASSETTYPE + typeId;
                        String assetDesc = new String (ctx.getObject(assetTypeKey.getBytes()));
                        buf.append(("assetid=" + assetId + ",typeid=" + typeId + ",assetDesc=" + assetDesc + "\n"));
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response newAssetToUser(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("only admin can add new asset type");
        }
        String  userId = new String(ctx.args().get("user_id"));
        String  typeId = new String(ctx.args().get("type_id"));
        String  assetId = new String(ctx.args().get("asset_id"));
        if (userId.length() == 0 || typeId.length() == 0 || assetId.length() == 0) {
            return Response.error("missing user_id or type_id or asst_id");
        }
        String  assetKey = ASSET2USER+ assetId;

        if (ctx.getObject(assetKey.getBytes()).length > 0) {
            return Response.error("asset id already exist");
        }

        String  userAssetKey = (USERASSET + userId + "_" + assetId);
        ctx.putObject(userAssetKey.getBytes(), typeId.getBytes());
        ctx.putObject(assetKey.getBytes(), userId.getBytes());
        return Response.ok(assetId.getBytes());
    }

    public Response tradeAsset(Context ctx) {
        String  from = ctx.caller();
        if (from.length() == 0) {
            return Response.error("missing initiator");
        }
        byte[] to = ctx.args().get("to");
        String assetId = new String(ctx.args().get("assetid"));
        if (to.length == 0 || assetId.length() == 0) {
            return Response.error("missing to or assetid");
        }
        String  userAssetKey = USERASSET + from + "_" + assetId;
        String assetType = new String(ctx.getObject(userAssetKey.getBytes()));
        if (assetType.length() == 0) {
            return Response.error("you do not have asset with id " + assetId);
        }

        ctx.deleteObject(userAssetKey.getBytes());
        String assetKey = ASSET2USER+ assetId;
        String  newUserAssetKey = USERASSET + to + "_" + assetId;
        ctx.putObject(newUserAssetKey.getBytes(), assetId.getBytes());
        ctx.putObject(assetKey.getBytes(), to);
        return Response.ok(assetId.getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new GameAssets());
    }
}