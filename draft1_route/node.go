// =============================================================================
// Auth: Alex Celani
// File: node.go
// Revn: 06-22-2022  1.0
// Func: receive data as an endpoint, send response back to middleman
//
// TODO:
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-04-2022: init
// 05-09-2022: began commenting
// 05-10-2022: finished commenting
//             added alter() and forward()
//             made passed value to alter()/forward() a slice
//             changed refs to recv byte array to slice of n bytes
// 06-20-2022: changed name to reflect first draft of node.go
//*06-22-2022: bug squashing and commenting
//
// =============================================================================

package main

import (
    "net"       // ResolveTCPAddr, conn.Write,Read,Close
                // ListenTCP, listener.Accept
    "os"        // Args, Stderr, Exit
    "fmt"       // Fprintf, Println
    "strings"   // ToLower
)


// quick error checking
func check( err error ) {
    // if user checks an error, it better be nil
    if err != nil {
        // put error in string, print in stderr
        fmt.Fprintf( os.Stderr, "Fatal error: %s", err.Error() )
        os.Exit( 2 )        // quit
    }
}


// parse incoming message, keep only second word
func parse( recv string ) string {
    // split message over spaces
    keywords := strings.Split( recv, " " )
    return keywords[1]      // return the second word
}


// function to handle incoming connections
func handleClient( conn net.Conn ) {
    defer conn.Close()  // barring any error, still close connection

    // init map of fields to their values
    reply := map[string]string{ "soc" : "30",
                                "batt" : "30",
                                "temp" : "65",
                                "humid" : "35",
                                "sun" : "10",
                                "list" : "soc batt temp humid sun list" }

    var buf [512]byte   // declare large byte array, store messages

    // iterate forever to always read over connection
    for {
        // read n bytes from connection into buffer
        n, err := conn.Read( buf[0:] )
        if err != nil { // erroring on read will simply leave the 
            return      // function so it can start again later
        }

        // print recv'd message
        // string() only works on byte SLICES so [:] is required
        fmt.Println( "recv: ", string( buf[:n] ) )

        // send input to parse
        recv := parse( string( buf[:n] ) )
        fmt.Println( "parse: ", recv )      // print

        // declare variables to grab values from field map
        var resp string
        var prs bool
        // get values from map, prs is false if key (recv) DNE in map
        resp, prs = reply[recv]
        if !prs {                           // if key was not present
            resp = "field does not exist"   // response is DNE
        }
        fmt.Println( "answer: ", resp )     // print response

        // write that response back to original client
        _, err = conn.Write( []byte( resp ) )
        if err != nil { // erroring on write will simply leave the
            return      // function so it can start again later
        }

        fmt.Println( "sent: ", resp )   // print response
    }
}


func main() {

    // ip:port
    // ip doesn't exist, implies localhost
    service := ":1202"      // capture ip address and host

    // "resolve" ip & host according to TCP rules
    tcpAddr, err := net.ResolveTCPAddr( "tcp", service )
    check( err )    // check error

    // bind and "listen" to ip and port, according to tcp rules
    listener, err := net.ListenTCP( "tcp", tcpAddr )
    check( err )    // check error

    // iterate forever
    // TODO i mean i can totally make this more user friendly
    for {
        // accept a connection that makes its way to bound port
        conn, err := listener.Accept()
        if err != nil {     // if connection fails...
            continue        // don't quit program, not fatal error
        }

        // asynchronous function to handle connection to client
        go handleClient( conn )
    }
}

