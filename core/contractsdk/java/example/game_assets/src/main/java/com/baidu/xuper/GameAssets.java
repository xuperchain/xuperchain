
package com.baidu.xuper;

public class GameAssets implements Contract {
    final String ASSETTYPE = "AssetType_";
    final String USERASSET = "UserAsset";
    final String ASSET2USER = "Asset2User_";
    final String ADMIN = "admin";


    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        String admin = ctx.caller();
        if (admin == null || admin.isEmpty()) {
            return Response.error("missing admin address");
        }
        ctx.putObject(ADMIN.getBytes(), admin.getBytes());
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
    public Response addAssetType(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("you do not have permission to call this method");
        }
        if (ctx.args().get("type_id") == null || ctx.args().get("type_desc") == null) {
            return Response.error("missing type_id or type_desc");
        }

        String typeId = new String(ctx.args().get("type_id"));
        String typeDesc = new String(ctx.args().get("type_desc"));
        if (typeId.isEmpty() || typeDesc.isEmpty()) {
            return Response.error("missing type_id or type_desc");
        }

        String assetTypeKey = ASSETTYPE + typeId;
        if (ctx.getObject(assetTypeKey.getBytes()) != null) {
            return Response.error("asset type "+typeId+" already exists");
        }
        ctx.putObject(assetTypeKey.getBytes(), typeDesc.getBytes());
        return Response.ok(typeId.getBytes());
    }

    @ContractMethod
    public Response listAssetType(Context ctx) {
        StringBuffer buf = new StringBuffer();
        ctx.newIterator(ASSETTYPE.getBytes(), (ASSETTYPE + "~").getBytes())
                .forEachRemaining(
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
    public Response getAssetByUser(Context ctx) {
        if (ctx.args().get("user_id")==null){
            return Response.error("missing user_id");
        }
        String userId = new String(ctx.args().get("user_id"));
        String userAsstKey = USERASSET + userId + "_";
        String start = userAsstKey;
        String end = start + "~";
        StringBuffer buf = new StringBuffer();
        ctx.log("1");
        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                elem -> {
                    String assetId = new String(elem.getKey());
                    String typeId = new String(elem.getValue());
                    String assetTypeKey = ASSETTYPE + typeId;
                    String assetDesc = new String(ctx.getObject(assetTypeKey.getBytes()));
                    buf.append("assetId=");
                    buf.append(assetId.split("_")[1]);
                    buf.append(",type.id=");
                    buf.append(typeId);
                    buf.append(",assetDesc=");
                    buf.append(assetDesc);
                    buf.append("\n");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response newAssetToUser(Context ctx) {

        if (!this.isAdmin(ctx)) {
            return Response.error("you do not have permission to call this method");
        }
        if (ctx.args().get("user_id") == null ||
                ctx.args().get("type_id") == null ||
                ctx.args().get("asset_id") == null) {
            return Response.error("missing user_id or type_id or asset_id");
        }


        String userId = new String(ctx.args().get("user_id"));
        String typeId = new String(ctx.args().get("type_id"));
        String assetId = new String(ctx.args().get("asset_id"));
        if (userId.isEmpty() || typeId.isEmpty() || assetId.isEmpty()) {
            return Response.error("missing user_id or type_id or asst_id");
        }

        String assetKey = ASSET2USER + assetId;

        if (ctx.getObject(assetKey.getBytes()) != null) {
            return Response.error("asset " + assetId + " exists");
        }

        byte[] assetDesc = ctx.getObject((ASSETTYPE + typeId).getBytes());
        if (assetDesc == null || assetDesc.length == 0) {
            return Response.error("asset type " + typeId + " not found");
        }

        String userAssetKey = USERASSET + userId + "_" + assetId;

        ctx.putObject(userAssetKey.getBytes(), typeId.getBytes());

        ctx.putObject(assetKey.getBytes(), userId.getBytes());
        return Response.ok(assetId.getBytes());
    }

    public Response tradeAsset(Context ctx) {
        String from = ctx.caller();
        if (from.length() == 0) {
            return Response.error("missing initiator");
        }
        byte[] to = ctx.args().get("to");
        String assetId = new String(ctx.args().get("assetid"));
        if (to.length == 0 || assetId.length() == 0) {
            return Response.error("missing to or assetid");
        }
        String userAssetKey = USERASSET + from + "_" + assetId;
        String assetType = new String(ctx.getObject(userAssetKey.getBytes()));
        if (assetType.length() == 0) {
            return Response.error("you do not have asset with id " + assetId);
        }

        ctx.deleteObject(userAssetKey.getBytes());
        String assetKey = ASSET2USER + assetId;
        String newUserAssetKey = USERASSET + to + "_" + assetId;
        ctx.putObject(newUserAssetKey.getBytes(), assetId.getBytes());
        ctx.putObject(assetKey.getBytes(), to);
        return Response.ok(assetId.getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new GameAssets());
    }
}