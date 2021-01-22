package com.baidu.xuper.example;

import java.math.BigInteger;
import java.util.concurrent.atomic.AtomicInteger;

import com.baidu.xuper.Context;
import com.baidu.xuper.Contract;
import com.baidu.xuper.ContractMethod;
import com.baidu.xuper.Driver;
import com.baidu.xuper.Response;

/**
 * Counter
 */
public class CharityDemo implements Contract {
    final String USERDONATE = "UserDonate_";
    final String ALLDONATE = "AllDonate_";
    final String ALLCOST = "AllCost_";
    final String TOTALDONATES = "TotalDonates";
    final String TOTALCOSTS = "TotalCosts";
    final String BALANCE = "Balance_";
    final String DONATECOUNT = "DonateCount_";
    final String COSTCOUNT = "CostCount_";
    final String ADMIN = "admin";
    final String MAX_LIMIT = "100";

    @Override
    @ContractMethod
    public Response initialize(Context ctx) {
        byte[] admin = ctx.args().get("admin");
        ctx.putObject(ADMIN.getBytes(), admin);
        ctx.putObject(TOTALDONATES.getBytes(), "0".getBytes());
        ctx.putObject(TOTALCOSTS.getBytes(), "0".getBytes());
        ctx.putObject(BALANCE.getBytes(), "0".getBytes());
        ctx.putObject(DONATECOUNT.getBytes(), "0".getBytes());
        ctx.putObject(COSTCOUNT.getBytes(), "0".getBytes());
        return Response.ok("ok".getBytes());
    }

    private boolean isAdmin(Context ctx) {
        String caller = ctx.caller();
        if (caller == null || caller.length() == 0) {
            return false;
        }
        return new String(ctx.getObject(ADMIN.getBytes())).equals(caller);
    }

    private String getIdFromNum(BigInteger num) {

        String numStr = num.toString();
        if (numStr.length() >= 20) {
            return numStr;
        }
        int count = 20 - numStr.length();
        return new String(new char[count]).replace('\0', '0') + numStr;
    }

    @ContractMethod
    public Response donate(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("you do not have permission to call this method");
        }
        if (ctx.args().get("donor") == null || ctx.args().get("donor").length == 0) {
            return Response.error("missing donor");
        }

        if (ctx.args().get("amount") == null || ctx.args().get("amount").length == 0) {
            return Response.error("missing amount");
        }

        if (ctx.args().get("timestamp") == null || ctx.args().get("timestamp").length == 0) {
            return Response.error("missing timestamp");
        }

        if (ctx.args().get("comments") == null || ctx.args().get("comments").length == 0) {
            return Response.error("missing comments");
        }

        String donor = new String(ctx.args().get("donor"));
        BigInteger amount = new BigInteger(new String(ctx.args().get("amount")));
        String timestamp = new String(ctx.args().get("timestamp"));
        String comments = new String(ctx.args().get("comments"));

        if (amount.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("amount must be positive");
        }
        BigInteger donateCount = new BigInteger(new String(ctx.getObject(DONATECOUNT.getBytes())));
        BigInteger totalDonates = new BigInteger(new String(ctx.getObject(TOTALDONATES.getBytes())));
        BigInteger balance = new BigInteger(new String(ctx.getObject(BALANCE.getBytes())));

        totalDonates = totalDonates.add(amount);
        balance = balance.add(amount);
        donateCount = donateCount.add(BigInteger.ONE);

        String donateId = getIdFromNum(donateCount);

        String userDonateKey = USERDONATE + donor + "%" + donateCount.toString();
        String allDonateKey = ALLDONATE + donateId;
        String donateDetail = "donor=" + donor + ",amount=" + amount + ",timestamp=" + timestamp + ",comments = "
                + comments;
        ctx.putObject(userDonateKey.getBytes(), donateDetail.getBytes());
        ctx.putObject(allDonateKey.getBytes(), donateDetail.getBytes());
        ctx.putObject(DONATECOUNT.getBytes(), donateCount.toString().getBytes());
        ctx.putObject(TOTALDONATES.getBytes(), totalDonates.toString().getBytes());
        ctx.putObject(BALANCE.getBytes(), balance.toString().getBytes());
        return Response.ok(donateId.getBytes());
    }

    @ContractMethod
    public Response cost(Context ctx) {
        if (!this.isAdmin(ctx)) {
            return Response.error("you do not have permission to call this method");
        }
        if (ctx.args().get("to") == null || ctx.args().get("to") == null) {
            return Response.error("missing to");
        }
        if (ctx.args().get("amount") == null || ctx.args().get("amount") == null) {
            return Response.error("missing amount");
        }
        if (ctx.args().get("timestamp") == null || ctx.args().get("timestamp") == null) {
            return Response.error("missing to");
        }
        if (ctx.args().get("comments") == null || ctx.args().get("comments") == null) {
            return Response.error("missing comments");
        }

        String to = new String(ctx.args().get("to"));
        BigInteger amount = new BigInteger(new String(ctx.args().get("amount")));
        String timestamp = new String(ctx.args().get("timestamp"));
        String comments = new String(ctx.args().get("comments"));

        if (amount.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("amount must greater than 0");
        }

        BigInteger totalCost = new BigInteger(new String(ctx.getObject(TOTALCOSTS.getBytes())));
        BigInteger balance = new BigInteger(new String(ctx.getObject(BALANCE.getBytes())));
        if (balance.compareTo(amount) < 0) {
            return Response.error("balance not enough");
        }
        balance = balance.subtract(amount);
        totalCost = totalCost.add(amount);

        BigInteger costCount = new BigInteger(new String(ctx.getObject(COSTCOUNT.getBytes())));
        costCount = costCount.add(BigInteger.ONE);

        String costId = getIdFromNum(costCount);
        String allCostKey = ALLCOST + costId;
        String costDetail = "to=" + to + ",amount=" + amount + ",timestamp=" + timestamp + ",comments=" + comments;

        ctx.putObject(allCostKey.getBytes(), costDetail.getBytes());
        ctx.putObject(COSTCOUNT.getBytes(), costCount.toString().getBytes());
        ctx.putObject(TOTALCOSTS.getBytes(), totalCost.toString().getBytes());
        ctx.putObject(BALANCE.getBytes(), balance.toString().getBytes());
        return Response.ok(costId.getBytes());
    }

    @ContractMethod
    public Response statistics(Context ctx) {
        String totalCost = new String(ctx.getObject(TOTALCOSTS.getBytes()));
        String balance = new String(ctx.getObject(BALANCE.getBytes()));
        String totalDonates = new String(ctx.getObject(TOTALDONATES.getBytes()));
        String result = "totalDonates=" + totalDonates + "," + "totalCosts=" + totalCost + "," + "fundBalance="
                + balance;
        return Response.ok(result.getBytes());
    }

    @ContractMethod
    public Response queryDonor(Context ctx) {
        if (ctx.args().get("donor") == null || ctx.args().get("donor").length == 0) {
            return Response.error("missing donor");
        }
        String donor = new String(ctx.args().get("donor"));
        String userDonateKey = USERDONATE + donor + "%";
        StringBuffer buf = new StringBuffer();
        AtomicInteger donateCount = new AtomicInteger();
        ctx.newIterator(userDonateKey.getBytes(), (userDonateKey + "~").getBytes()).forEachRemaining(elem -> {
            donateCount.getAndIncrement();
            String donateId = new String(elem.getKey()).substring(userDonateKey.length());
            String content = new String(elem.getValue());
            buf.append(("id=" + donateId + "," + "content=" + content + "\n"));
        });
        return Response.ok(buf.toString().getBytes());

    }

    @ContractMethod
    public Response queryDonates(Context ctx) {
        if (ctx.args().get("start") == null || ctx.args().get("start").length == 0) {
            return Response.error("missing start");
        }
        if (ctx.args().get("limit") == null || ctx.args().get("limit").length == 0) {
            return Response.error("missing limit");
        }
        String start = new String(ctx.args().get("start"));
        BigInteger limit = new BigInteger(new String(ctx.args().get("limit")));

        if (limit.compareTo(new BigInteger(MAX_LIMIT)) >= 0) {
            return Response.error("limit exceeded");
        }

        String startKey = ALLDONATE + start;
        String endKey = ALLDONATE + getIdFromNum(new BigInteger(start).add(limit));
        AtomicInteger selected = new AtomicInteger();
        StringBuffer buf = new StringBuffer();
        ctx.newIterator(startKey.getBytes(), (endKey).getBytes()).forEachRemaining(elem -> {
            if (selected.get() >= limit.intValue()) {
                return;
            }
            selected.incrementAndGet();
            String donateId = new String(elem.getKey()).substring(ALLDONATE.length());
            String content = new String(elem.getValue());
            buf.append("id=" + donateId + "," + content + "\n");
        });
        return Response.ok(buf.toString().getBytes());
    }

    @ContractMethod
    public Response queryCosts(Context ctx) {
        if (ctx.args().get("start") == null || ctx.args().get("start").length == 0) {
            return Response.error("start");
        }
        if (ctx.args().get("limit") == null || ctx.args().get("limit").length == 0) {
            return Response.error("limit");
        }
        String startId = new String(ctx.args().get("start"));
        BigInteger limit = new BigInteger(new String(ctx.args().get("limit")));

        if (limit.compareTo(BigInteger.ZERO) <= 0) {
            return Response.error("limit must be positive");
        }

        if (limit.compareTo(new BigInteger(MAX_LIMIT)) >= 0) {
            return Response.error("limit exceeded");
        }

        String startKey = ALLCOST + startId;
        String endKey = ALLCOST + getIdFromNum(new BigInteger(startId).add(limit));
        AtomicInteger selected = new AtomicInteger();
        StringBuffer buf = new StringBuffer();
        ctx.newIterator(startKey.getBytes(), endKey.getBytes()).forEachRemaining(elem -> {
            if (selected.get() >= limit.intValue()) {
                return;
            }
            selected.getAndIncrement();
            String donateId = new String(elem.getKey()).substring(ALLCOST.length());
            String content = new String(elem.getValue());
            buf.append("id=" + donateId + "," + content + "\n");
        });
        return Response.ok(buf.toString().getBytes());
    }

    public static void main(String[] args) {
        Driver.serve(new CharityDemo());
    }
}