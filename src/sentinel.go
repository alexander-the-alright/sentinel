// =============================================================================
// Auth: Alex Celani
// File: sentinel.go
// Revn: 07-06-2022  2.0
// Func: Receive message, determine recipient, forward, get response,
//       forward back
//
// TODO:
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-04-2022: init
// 05-05-2022: commented
// 05-10-2022: added argument 'n' to alter() for halving purposes
//             changed passed value to alter()/forward() a slice
// 06-20-2022: changed name to reflect first draft of sentinel.go
//*06-22-2022: bug fixing and commenting
// 06-23-2022: removed call to parse
// 06-26-2022: refined a new route() function to check for destination
//                  validity
// 06-27-2022: added flag package
// 07-05-2022: added support for kill messages
//*07-06-2022: added support for mulitplexing multiple endpoints
//                  so long as their name and ip is in /nodes/ file
//
// =============================================================================

package main

import (
    "net"       // ResolveTCPAddr, DialTCP, conn.Write,Read,Close
                // ListenTCP, listener.Accept
    "os"        // Exit, Stderr, Stdin, Args
    "fmt"       // Println, Fprintf
    "strings"   // ToLower, ToUpper
    "flag"      // Parse, String, Bool
    "io/ioutil" // ReadFile
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


// read the given message and route to the right endpoint
func route( recv []string ) string {

    ip := nodes[recv[1]]    // get correlating ip to given node name

    return ip

}


// function to send message to third party and get response
func send( toSend, ip string ) string {

    // "resolve" ip & host according to TCP rules
    tcpAddr, err := net.ResolveTCPAddr( "tcp", ip )
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

        if *verbose {        // print recv'd message
            // string() only works on byte SLICES so [:] is required
            fmt.Println( "recv: ", recv )
        }

        // split message into separate words
        keywords := strings.Split( recv, " " )

        // parse the input in some way
        var resp string             // declare response variable

        // if user wants list of endpoints
        if keywords[1] == "list" {
            // iterate over keys in maps ( endpoint names )
            for key, _ := range nodes {
                // collect keys, delimit with newline, in string
                resp = resp + key + "\n"
            }
            // last char is newline, remove that
            resp = resp[:len( resp ) - 1]
        } else {    // user wants to send message
            ip := route( keywords ) // find ip of node, if it exists
            if ip == "" {           // if it doesn't exist
                return
            }
            resp = send( recv, ip ) // send message, get response
        }

        if *verbose {
            fmt.Println( "node: ", resp )   // print response
        }

        // write that response back to original client
        _, err = conn.Write( []byte( resp ) )
        if err != nil { // erroring on write will simply leave the
            return      // function so it can start again later
        }

        if resp == "kill" {     // if received kill message
            os.Exit( 1 )        // kill process
        }
    }
}


// function to read in initialization data for nodes
func initNodes() {

    // open and read config file
    file, err := ioutil.ReadFile( "nodes" )
    check( err )    // error checking on read

    // cast file to string, split string over newlines
    node := strings.Split( string( file ), "\n" )

    for i, kvp := range node {      // iterate over lines
        // XXX for whatever reason, the EOF is seen as a newline?
        // so this only works for all but the "last line"
        if i != len( node ) - 1 {
            // kvp is delimited by a bar, split over bar
            // key is the name of the endpoint
            // ip is the ip of the same endpoint
            nameip := strings.Split( kvp, "|" )
            // insert kvp into map
            nodes[nameip[0]] = nameip[1]
        }
    }
}


var (           // declare global variables for flag
    self *string
    verbose *bool
)

// declare global variable to contain names and ips of endpoints
var nodes = make( map[string]string )


func main() {

    // get flags for verbose, self ip and destination ip
    verbose = flag.Bool( "v", false, "flag to print extra info" )
    self = flag.String( "self", ":1201", "port of self" )
    flag.Parse()        // parse

    // read /nodes/ file and populate nodes map
    initNodes()

    // "resolve" ip & host according to TCP rules
    tcpAddr, err := net.ResolveTCPAddr( "tcp", *self )
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

