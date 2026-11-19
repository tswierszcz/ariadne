package fx;

import java.io.InputStream;
import java.io.OutputStream;
import java.net.ServerSocket;
import java.net.Socket;
import java.nio.charset.StandardCharsets;
import org.eclipse.jetty.client.HttpClient;
import org.eclipse.jetty.client.api.ContentResponse;

/**
 * LIBRARY_MODE LB4 oracle for jetty-client 9.4.36: a user of the library
 * calls only HttpClient#start, #GET and #stop. A local socket answers with
 * a "+"-prefixed Content-Length (CVE-2023-40167); the request URI goes
 * through HttpURI (CVE-2022-2047). Run by scripts/library-lb4-oracle.sh.
 */
public class JettyGet {
    public static void main(String[] args) throws Exception {
        try (ServerSocket ss = new ServerSocket(0)) {
            Thread server = new Thread(() -> {
                try (Socket s = ss.accept()) {
                    InputStream in = s.getInputStream();
                    byte[] buf = new byte[4096];
                    int n, total = 0;
                    String req = "";
                    while (!req.contains("\r\n\r\n") && (n = in.read(buf)) > 0) {
                        req += new String(buf, 0, n, StandardCharsets.ISO_8859_1);
                        total += n;
                    }
                    OutputStream out = s.getOutputStream();
                    out.write(("HTTP/1.1 200 OK\r\nContent-Length: +5\r\nConnection: close\r\n\r\nhello")
                            .getBytes(StandardCharsets.ISO_8859_1));
                    out.flush();
                } catch (Exception e) {
                    e.printStackTrace();
                }
            });
            server.start();
            HttpClient client = new HttpClient();
            client.start();
            try {
                ContentResponse r = client.GET("http://localhost:" + ss.getLocalPort() + "/");
                System.out.println("status=" + r.getStatus() + " body=" + r.getContentAsString());
            } finally {
                client.stop();
            }
            server.join();
        }
    }
}
