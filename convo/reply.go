// =============================================================================
// Auth: Alex Celani
// File: reply.go
// Revn: 05-13-2022  0.1
// Func: host connection, reply to speak.go
//
// TODO: create
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-13-2022: init
//
// =============================================================================

package main

import (
    "net"
    "os"
    "fmt"
    "time"
)


func check( err error ) {
    if err != nil {
        fmt.Println( "Fatal: ", err.Error() )
        os.Exit( 1 )
    }
}


func handle( conn net.Conn ) {

    table := map[string]string{ "hello":"hello",
                                "how are you":"fine and yourself",
                                "well enough for the weather":"aye it scalds",
                                "i lost my begonias":"exit" }
    var buffer [512]byte
    for {
        n, err := conn.Read( buffer[:] )
        check( err )

        fmt.Println( "\t\t\t", string( buffer[:n] ) )

        time.Sleep( 2 * time.Second )

        response, present := table[string( buffer[:n] )]
        if present {
            _, err = conn.Write( []byte( response ) )
            check( err )


            if response == "exit" {
                break
            } else {
                fmt.Println( response )
            }
        } else {
            fmt.Println( "excuse me good sir" )
            os.Exit( 2 )
        }
    }
}


func main() {
    
    service := ":1200"

    addr, err := net.ResolveTCPAddr( "tcp", service )
    check( err )

    listener, err := net.ListenTCP( "tcp", addr )

    conn, err := listener.Accept()
    check( err )

    handle( conn )

    os.Exit( 0 )

}

