package com.thoughtworks.xstream.security;

public interface TypePermission {
    boolean allows(Class type);
}
