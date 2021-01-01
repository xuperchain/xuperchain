package com.baidu.xuper.example;

import java.math.BigDecimal;
import java.util.concurrent.atomic.AtomicInteger;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;
import com.baidu.xuper.utils.ByteUtils;
import com.sun.tools.javac.util.ByteBuffer;

/**
 * Counter
 */
public class CharityDemo implements Contract {
    final String USERDONATE = "UserDonate_";
    final String ALLDONATE = "AllDonate_";
    final String ALLCOST = "AllCost_";
    final String TOTALRECEIVED = "TotalDonates";
    final String TOTALCOSTS = "TotalCosts";
    final String BALANCE = "Balance_";
    final String DONATECOUNT = "DonateCount_";
    final String COSTCOUNT = "CostCount_";
    final String ADMIN = "admin";
    final int MAX_LIMIT = 100; // uint64???


    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        byte[] admin = ctx.args().get("admin".getBytes());
        ctx.putObject(ADMIN.getBytes(), admin);
        ctx.putObject(TOTALRECEIVED.getBytes(), "0".getBytes());
        ctx.putObject(TOTALCOSTS.getBytes(), "0".getBytes());
        ctx.putObject(BALANCE.getBytes(), "0".getBytes());
        ctx.putObject(DONATECOUNT.getBytes(), "0".getBytes());
        ctx.putObject(COSTCOUNT.getBytes(), "0".getBytes());
        return Response.ok("ok".getBytes());
    }

    private boolean isAdmin(Context ctx) {
        String caller = ctx.caller();
        return new String(ctx.getObject(ADMIN.getBytes())).equals(caller);
    }

    private String getIdFromNum(BigDecimal num) {

        String numStr = num.toString();
        if (numStr.length() >= 20) {
            return numStr;
        }
        int count = 20 - numStr.length(); // magic number is from cpp examples
        return new String(new char[count]).replace('\0', '0') + numStr;
    }

    @ContractMethod
    public Response donate(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("only admin can add donate record");
        }
        String donor = new String(ctx.args().get("donor"));
        BigDecimal amount = new BigDecimal(new String(ctx.args().get("amount")));
        String timestamp = new String(ctx.args().get("timestamp"));
        String comments = new String(ctx.args().get("comments"));
        if (donor.length() == 0 || timestamp.length() == 0 || comments.length() == 0) {
            return Response.error("missing donor or amount or timestamp or comments");
        }
        if (amount.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("amount must be positive");
        }
        BigDecimal donateCount = new BigDecimal(new String(ctx.getObject(DONATECOUNT.getBytes())));
        BigDecimal totalRecv = new BigDecimal(new String(ctx.getObject(TOTALRECEIVED.getBytes())));
        BigDecimal balance = new BigDecimal(new String(ctx.getObject(BALANCE.getBytes())));
        totalRecv = totalRecv.add(amount);
        balance = balance.add(amount);
        donateCount = donateCount.add(BigDecimal.ONE);

        String donateId = getIdFromNum(donateCount);

        String userDonateKey = USERDONATE + donor + "%" + donateId;
        String allDonateKey = (ALLDONATE + donateId);
        String donateDetail = "donor=" + donor +
                ",amount=" + amount +
                "timestamp=" + timestamp +
                "comments = " + comments;
        ctx.putObject(userDonateKey.getBytes(), donateDetail.getBytes());
        ctx.putObject(allDonateKey.getBytes(), donateDetail.getBytes());
        ctx.putObject(DONATECOUNT.getBytes(), donateCount.toString().getBytes());
        ctx.putObject(TOTALRECEIVED.getBytes(), totalRecv.toString().getBytes());
        ctx.putObject(BALANCE.getBytes(), balance.toString().getBytes());
        return Response.ok(donateId.getBytes());
    }

    @ContractMethod
    public Response cost(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("only admin can add cost record");
        }
        String to = new String(ctx.args().get("to"));
        BigDecimal amount = new BigDecimal(new String(ctx.args().get("amount")));
        String timestamp = new String(ctx.args().get("timestamp"));
        String comments = new String(ctx.args().get("comments"));

        if (to.length() == 0 || timestamp.length() == 0 || comments.length() == 0) {
            return Response.error("missing to or amount or timestamp or comments");
        }
        if (amount.compareTo(BigDecimal.ZERO) <= 0) {
            return Response.error("amount must be positive");
        }

        String totalCost = new String(ctx.getObject(TOTALCOSTS.getBytes()));
        BigDecimal balance = new BigDecimal(new String(ctx.getObject(BALANCE.getBytes())));
        BigDecimal costCount = new BigDecimal(new String(ctx.getObject(COSTCOUNT.getBytes())));
        if (balance.compareTo(amount) < 0) {
            return Response.error("fund balance not enough");
        }


        String costId = getIdFromNum(costCount);
        String allCostKey = ALLCOST + costId;
        String costDetail = "to=" + to +
                ",amount=" + amount +
                ",timestamp=" + timestamp +
                ",comments=" + comments;


        ctx.putObject(allCostKey.getBytes(), costDetail.getBytes());
        ctx.putObject(COSTCOUNT.getBytes(), costCount.toString().getBytes()); // TODO 好像差一个
        ctx.putObject(TOTALCOSTS.getBytes(), totalCost.getBytes());
        ctx.putObject(BALANCE.getBytes(), balance.toString().getBytes());
        return Response.ok(costId.getBytes());
    }

    @ContractMethod
    public Response statistics(Context ctx) {
        String totalCost = new String(ctx.getObject(TOTALCOSTS.getBytes()));
        String balance = new String(ctx.getObject(BALANCE.getBytes()));
        String totalRec = new String(ctx.getObject(TOTALRECEIVED.getBytes()));
        String result =
                "totalDonates=" + totalRec + "," +
                        "totalCosts=" + totalCost + "," +
                        "fundBalance=" + balance;
        return Response.ok(result.getBytes());
    }

    @ContractMethod
    public Response queryDonor(Context ctx) {
        String donor = new String(ctx.args().get("donor".getBytes()));
        if (donor.length() == 0) {
            return Response.error("missing donor argument");
        }
        String userDonateKey = USERDONATE + donor + "%";
        ByteBuffer buf = new ByteBuffer();
        AtomicInteger donateCount = new AtomicInteger(); // TODO @fengjin
        ctx.newIterator(userDonateKey.getBytes(), (userDonateKey + "~").getBytes()).forEachRemaining(
                elem -> {
//                    TODO 验证
                    donateCount.getAndIncrement();
                    String donateId = new String(elem.getKey()).substring(userDonateKey.length());
                    String content = ByteUtils.toString(elem.getValue());
                    buf.appendBytes((
                            "id=" + donateId + "," +
                                    "content=" + content + "\n"
                    ).getBytes());
                }
        );
        return Response.ok(buf.elems); // TODO @fengjin

    }

    @ContractMethod
    public Response queryDonates(Context ctx) {
        String  startId = new String(ctx.args().get("startid"));
        BigDecimal  limit = new BigDecimal( new String(ctx.args().get("limit")));
        if (startId.length() == 0) {
            return Response.error("missing startid or limit");
        }
        if (limit.compareTo(BigDecimal.ZERO) <=0){
            return Response.error("limit must be positive");
        }

        if (limit.compareTo( new BigDecimal(MAX_LIMIT))>=0) {
            return Response.error("limit must be less than" + MAX_LIMIT);
        }

        String  donateKey = ALLDONATE+ startId;
        int limit1 = 0;
        AtomicInteger selected = new AtomicInteger();
        ByteBuffer buf = new ByteBuffer();
        ctx.newIterator(donateKey.getBytes(), (donateKey+ "~").getBytes()).forEachRemaining(
                elem -> {
                    if (selected.get() >= limit1) {// 注意边界条件
                        return;
                    }
//                    TODO 检查
                    selected.getAndIncrement();
                    String donateId = ByteUtils.toString(elem.getKey()).substring(ALLDONATE.length());
                    String content = ByteUtils.toString(elem.getValue());
                    buf.appendBytes((
                            "id=" + donateId + "," +
                                    content + "\n" // TODO
                    ).getBytes());
                }
        );
        return Response.ok(buf.elems);
    }

    @ContractMethod
    public Response queryCosts(Context ctx) {
        String  startId = new String(ctx.args().get("startid"));
        BigDecimal  limit = new BigDecimal( new String(ctx.args().get("limit")));
        if (startId.length() == 0) {
            return Response.error("missing startid or limit");
        }
        if (limit.compareTo(BigDecimal.ZERO) <=0){
            return Response.error("limit must be positive");
        }

        if (limit.compareTo( new BigDecimal(MAX_LIMIT))>=0) {
            return Response.error("limit must be less than" + MAX_LIMIT);
        }

        String donateKey = ALLCOST+startId;
        int limit1 = 0; // TODO
        AtomicInteger selected = new AtomicInteger();
        ByteBuffer buf = new ByteBuffer();
        ctx.newIterator(donateKey.getBytes(), (donateKey+ "~").getBytes()).forEachRemaining(
                elem -> {
                    if (selected.get() >= limit1) {// 注意边界条件
                        return;
                    }
//                    TODO 检查
                    selected.getAndIncrement();
                    String donateId = ByteUtils.toString(elem.getKey()).substring(ALLCOST.length());
                    String content = ByteUtils.toString(elem.getValue());
                    buf.appendBytes((
                            "id=" + donateId + "," +
                                    content + "\n" // TODO
                    ).getBytes());
                }
        );
        return Response.ok(buf.elems);
    }


    public static void main(String[] args) {
        Driver.serve(new CharityDemo());
    }
}