// =============================================================================
// Auth: Alex Celani
// File: speak.go
// Revn: 05-11-2022  0.1
// Func: initiate conversation
//
// TODO: create
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-11-2022: init
// 05-17-2022: changed an entry in the response table
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
        fmt.Println( "fatal: ", err.Error() )
        os.Exit( 1 )
    }
}


func handle( conn net.Conn ) {

    table := map[string]string{ "hello":"how are you",
                                "fine and yourself":"well enough for the weather",
                                "aye it scalds":"i lost my begonias" }
    var buffer [512]byte
    for {
        n, err := conn.Read( buffer[:] )
        check( err )

        if string( buffer[:n] ) == "exit" {
            break
        }

        fmt.Println( "\t\t\t", string( buffer[:n] ) )

        time.Sleep( 2 * time.Second )

        response, present := table[string( buffer[:n] )]
        if present {
            _, err = conn.Write( []byte( response ) )
            check( err )

            fmt.Println( response )
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

    conn, err := net.DialTCP( "tcp", nil, addr )
    check( err )

    _, err = conn.Write( []byte( "hello" ) )
    check( err )
    fmt.Println( "hello" )

    handle( conn )

    os.Exit( 0 )

}

