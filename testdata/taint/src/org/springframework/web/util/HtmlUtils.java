package org.springframework.web.util;

/** Stub of Spring's HtmlUtils (for the demo only). */
public final class HtmlUtils {
    public static String htmlEscape(String s) {
        return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;").replace("\"", "&quot;");
    }
}
