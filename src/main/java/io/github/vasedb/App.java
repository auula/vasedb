package io.github.vasedb;

import io.github.vasedb.server.HttpServer;

public class App {
    public static void main( String[] args )
    {
        HttpServer.run(":2468");
    }
}
