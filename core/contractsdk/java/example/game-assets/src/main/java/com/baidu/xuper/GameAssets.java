
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
            return Response.error("asset type " + typeId + " already exists");
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

        ctx.putObject(userAssetKey.getBytes(), (ASSETTYPE + typeId).getBytes());

        ctx.putObject(assetKey.getBytes(), userId.getBytes());
        return Response.ok(assetId.getBytes());
    }

    @ContractMethod
    public Response getAssetByUser(Context ctx) {
        if (ctx.args().get("user_id") == null) {
            return Response.error("missing user_id");
        }
        String userId = new String(ctx.args().get("user_id"));
        String userAsstKey = USERASSET + userId + "_";
        String start = userAsstKey;
        String end = start + "~";
        StringBuffer buf = new StringBuffer();
//        StringBuffer getAssetDescError = new StringBuffer();
        ctx.newIterator(start.getBytes(), end.getBytes()).forEachRemaining(
                elem -> {

                    String assetId = new String(elem.getKey()).substring((USERASSET + userId).length() + 1);
                    String typeId = new String(elem.getValue());
                    if (typeId.length() < ASSETTYPE.length()) {
                        return;
                    }
                    String assetTypeKey = typeId;
                    byte[] assetDescByte = ctx.getObject(assetTypeKey.getBytes());
                    String assetDesc = new String(assetDescByte);
                    buf.append("assetId=");
                    buf.append(assetId);
                    buf.append(",typeId=");
                    buf.append(typeId.substring((ASSETTYPE.length())));
                    buf.append(",assetDesc=");
                    buf.append(assetDesc);
                    buf.append("\n");
                }
        );
        return Response.ok(buf.toString().getBytes());
    }


    @ContractMethod
    public Response tradeAsset(Context ctx) {
        String from = ctx.caller();
        if (from.length() == 0) {
            return Response.error("missing initiator");
        }
        byte[] toByte = ctx.args().get("to");
        if (toByte == null || toByte.length == 0) {
            return Response.error("missing to");
        }
        byte[] assetIdByte = ctx.args().get("asset_id");
        if (assetIdByte == null || assetIdByte.length == 0) {
            return Response.error("missing asset_id");
        }
        String assetId = new String(assetIdByte);
        String to = new String(toByte);

        String userAssetKey = USERASSET + from + "_" + assetId;
        byte[] assetTypeByte = ctx.getObject(userAssetKey.getBytes());
        if (assetTypeByte == null || assetTypeByte.length == 0) {
            return Response.error("you do not have asset with id " + assetId);
        }

        ctx.deleteObject(userAssetKey.getBytes());

        String assetKey = ASSET2USER + assetId;
        String newUserAssetKey = USERASSET + to + "_" + assetId;

        ctx.putObject(newUserAssetKey.getBytes(), assetTypeByte);
        ctx.putObject(assetKey.getBytes(), to.getBytes());
        return Response.ok(assetId.getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new GameAssets());
    }
}