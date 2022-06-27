// =============================================================================
// Auth: Alex Celani
// File: sentinel.go
// Revn: 06-22-2022  1.0
// Func: Receive message, play with it, send to a third place, receive
//       response, send off to first place
//
// TODO: comment
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-04-2022: init
// 05-05-2022: commented
// 05-10-2022: added argument 'n' to alter() for halving purposes
//             changed passed value to alter()/forward() a slice
// 06-20-2022: changed name to reflect first draft of sentinel.go
//*06-22-2022: bug fixing and commenting
//
// =============================================================================

package main

import (
    "net"       // ResolveTCPAddr, DialTCP, conn.Write,Read,Close
                // ListenTCP, listener.Accept
    "os"        // Exit, Stderr, Stdin, Args
    "fmt"       // Println, Fprintf
    "strings"   // ToLower, ToUpper
)


// quick error checking
func check( err error ){
    // if user checks an error, it better be nil
    if err != nil {
        // put error in string, print to stderr
        fmt.Fprintf( os.Stderr, "Fatal error: %s", err.Error() )
        os.Exit( 1 )        // quit
    }
}


// function to send message to third party and get response
func send( toSend string ) string {
    // ip:port
    // ip doesn't exist, implies localhost
    service := ":1202"      // capture ip address and host

    // "resolve" ip & host according to TCP rules
    tcpAddr, err := net.ResolveTCPAddr( "tcp", service )
    check( err )    // check error

    // "dial" ( establish connection ) to destination ip & port
    // according to TCP rules
    // laddr = nil -> local address is localhost
    conn, err := net.DialTCP( "tcp", nil, tcpAddr )
    check( err )    // check error

    defer conn.Close()  // barring any error, still close connection

    // convert user input to bytes and send over connection
    _, err = conn.Write( []byte( toSend ) )
    check( err )    // check error

    var buf [512]byte                   // init byte array
    // read from connection into byte array
    // amount of bytes read stored in n
    //n, err := conn.Read( buf[0:] )
    n, err := conn.Read( buf[0:] )
    check( err )    // check error

    return string( buf[:n] )        // return only written bytes
}


// remove the first word from the string
// actually mad inefficient tbh, but it works
func parse( recv string ) string {
    // split string over spaces
    keywords := strings.Split( recv, " " )
    var command string      // declare return variable
    // iterate over all but the first word
    for _, keyword := range keywords[1:] {
        command = command + keyword     // add word to returnable
        command = command + " "         // separate by spaces
    }
    // remove the trailing space
    command = command[:len( command ) - 1]
    return command                      // return
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

        // instantiate receipt variable
        var recv string = string( buf[:n] )

        // print recv'd message
        // string() only works on byte SLICES so [:] is required
        fmt.Println( "recv: ", recv )

        // parse the input in some way
        recv = parse( recv )
        fmt.Println( "parse: ", recv )  // print parse

        resp := send( recv )            // send message off
        fmt.Println( "node: ", resp )   // print response

        // write that response back to original client
        _, err = conn.Write( []byte( resp ) )
        if err != nil { // erroring on write will simply leave the
            return      // function so it can start again later
        }
    }
}


func main() {

    // ip:port
    // ip doesn't exist, implies localhost
    service := ":1201"      // capture ip address and host

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

