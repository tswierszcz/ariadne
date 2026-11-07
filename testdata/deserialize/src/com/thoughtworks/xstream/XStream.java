package com.thoughtworks.xstream;

import com.thoughtworks.xstream.security.TypePermission;

/** Signature-only stub of XStream's configuration API (fixture). */
public class XStream {
    public Object fromXML(String xml) { return null; }
    public void allowTypes(Class[] types) {}
    public void allowTypes(String[] names) {}
    public void allowTypesByWildcard(String[] patterns) {}
    public void allowTypeHierarchy(Class type) {}
    public void denyTypes(Class[] types) {}
    public void addPermission(TypePermission permission) {}
    public static void setupDefaultSecurity(XStream xstream) {}
}
