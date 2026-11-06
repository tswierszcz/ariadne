package com.thoughtworks.xstream.security;

public class AnyTypePermission implements TypePermission {
    public static final TypePermission ANY = new AnyTypePermission();
    public boolean allows(Class type) { return true; }
}
