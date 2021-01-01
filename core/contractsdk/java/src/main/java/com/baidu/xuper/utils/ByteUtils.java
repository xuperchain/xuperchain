package com.baidu.xuper.utils;

import java.math.BigDecimal;

public class ByteUtils {
    public static byte[] stub(){
        return "not implemented".getBytes();
    }
    public static byte[] concat(byte[] a,byte[] b){
        byte[] c = new byte[a.length + b.length];
        System.arraycopy(a, 0, c, 0, a.length);
        System.arraycopy(b, 0, c, a.length, b.length);
        return c;
    }
    public static byte[]concat(String a,byte[]b){
        return concat(a.getBytes(),b);
    }
    public static byte[]concat(byte[] a,String b){
        return concat(a,b.getBytes());
    }
    public static byte[] concat(String a,String b){
        return concat(a.getBytes(),b.getBytes());
    }



//传递的数作为高精度数实现，
    public static byte[] add(byte[] a, byte[] b){
        return stub();
    }

    public static byte[] incrress(byte[] a){
        return stub();
    }

    public static byte[] sub(byte[] a,byte[] b){
        return stub();
    }


    public static boolean less(byte[] a,byte[] b){
        return false;

    }


    public static byte[] equal(byte[] a,byte[] b){
        return stub();
    }


    public static byte[] equal(byte[] a,int b){
        return stub();
    }

    public static byte[] le(byte[] a,int[]b){
        return stub();
    }
    public static byte[] ge(byte[] a,int[]b){
        return stub();
    }

    public static String toString(byte[]b){
        return "not implemented";
    }

    public static boolean equal(byte[]a,String b){
        return false;

    }
}
