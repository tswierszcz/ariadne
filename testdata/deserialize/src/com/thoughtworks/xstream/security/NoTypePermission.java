package com.thoughtworks.xstream.security;

public class NoTypePermission implements TypePermission {
    public static final TypePermission NONE = new NoTypePermission();
    public boolean allows(Class type) { return false; }
}
