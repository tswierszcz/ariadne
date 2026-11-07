package demo;

import javax.servlet.ServletRequest;

final class FakeRequest implements ServletRequest {
    private final String value;
    FakeRequest(String value) { this.value = value; }
    public String getParameter(String name) { return value; }
}
