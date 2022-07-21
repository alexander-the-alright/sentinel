// =============================================================================
// Auth: Alex Celani
// File: sentinel.go
// Revn: 07-14-2022  3.0
// Func: Receive message, determine recipient, forward, get response,
//       forward back
//
// TODO: add time flag
//       use tail command line args to send single message and quit
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
// 07-11-2022: added solo mode and flag
//             changed how list works
// 07-12-2022: comments
//*07-14-2022: changed input/output prompt to arrow
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
    "bufio"     // input()
)


// quick function to take user input
func input( ps1 string ) string {
    fmt.Print( ps1 )                            // print prompt
    scanner := bufio.NewScanner( os.Stdin )     // link to stdin
    scanner.Scan()                              // pull data
    return strings.ToLower( scanner.Text() )    // return lower text
}


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


// parse input, send to node, return response
func forward( recv string ) string {

    // break commands into words
    keywords := strings.Split( recv, " " )
    length := len( keywords )   // store length for later use

    // declare return variable
    var resp string

    if length == 1 {                // if length of input is only 1
        if keywords[0] == "list" {  // command better be /list/
            var i int = 0       // var to keep track of iterations
            // iterate over keys in map to get list of node names
            for name, _ := range nodes {
                if i != 0 {
                    // do not prepend spaces to first node
                    resp = resp + "   "
                }
                resp = resp + name + "\n"   // append node name
                i++             // keep track of iterations
            }
            resp = resp[:len( resp ) - 1]   // remove last newline
        } else {                    // command is not /list/, ergo
                                    // not supported
            resp = "command unrecognized"
        }
    } else if length > 1 {          // command is more than one word
        // user wants a comprehensive list of nodes and fields
        if keywords[0] == "list" && keywords[1] == "-v" {
            // iterate over all nodes
            for name, addr := range nodes {
                // add name, ip, newline, and justifying spaces
                resp = resp + name + " --> " + addr + "\n   "
                // send command /list/ to node
                resp = resp + send( "list " + name, addr )
                // add newline and justifying spaces
                resp = resp + "\n   "
            }
            // remove trailing newline and spaces
            resp = resp[:len( resp ) - 4]
        } else {        // user wants a different command
            ip := route( keywords )     // find appropriate ip
            if ip == "" {               // ip not found, node DNE
                resp = "node unrecognized"  // print error
            } else {    // node does exist
                // send message to node, get response
                resp = send( recv, ip )
            }
        }
    } else {        // length of command was 0
        resp = "command unrecognized"
    }

    return resp     // return response
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

        // parse input, send out to node, get response
        resp := forward( recv )

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
    solo *bool
)

// declare global variable to contain names and ips of endpoints
var nodes = make( map[string]string )


func main() {

    // get flags for verbose, self ip and destination ip
    verbose = flag.Bool( "v", false, "flag to print extra info" )
    self = flag.String( "self", ":1201", "port of self" )
    solo = flag.Bool( "s", false, "run sentinel in solo mode" )
    flag.Parse()        // parse

    // read /nodes/ file and populate nodes map
    initNodes()

    if *solo {      // if user requested stand-alone action
        for {       // iterate forever ( until user quits )
            in := input( "-> " )        // print prompt, take input
            resp := forward( in )       // forward directly to node
            fmt.Println( "<-", resp )   // print response
        }
    }

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

