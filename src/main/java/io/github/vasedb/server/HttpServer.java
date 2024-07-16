package io.github.vasedb.server;

public class HttpServer {

    private String port;

    public HttpServer(String port) {
        this.port = port;
    }

    public static void run(String port) {
        var server = new HttpServer(port);
    }

    public String getPort() {
        return port;
    }

    public void setPort(String port) {
        this.port = port;
    }
}
