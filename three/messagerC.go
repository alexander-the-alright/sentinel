// =============================================================================
// Auth: Alex Celani
// File: messagerC.go
// Revn: 05-10-2022  0.3
// Func:
//
// TODO: create
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-04-2022: init
// 05-09-2022: began commenting
// 05-10-2022: finished commenting
//             added alter() and forward()
//             made passed value to alter()/forward() a slice
//             changed refs to recv byte array to slice of n bytes
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


// function to change received message before passing it along
func alter( recv []byte ) string {
    var send string = string( recv[:] )
    send = send + " " + send
    send = strings.ToLower( send )
    return send
}


// function to convert message to string before forwarding
func forward( recv []byte ) string {
    return string( recv[:] )
}


// function to handle incoming connections
func handleClient( conn net.Conn ) {
    defer conn.Close()  // barring any error, still close connection

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

        send := alter( buf[:n] )        // send message ot get changed
//        send := forward( buf[:] )

        // write that response back to original client
        _, err = conn.Write( []byte( send ) )
        if err != nil { // erroring on write will simply leave the
            return      // function so it can start again later
        }

        fmt.Println( "sent: ", send )   // print response
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



