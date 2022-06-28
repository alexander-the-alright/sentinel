// =============================================================================
// Auth: Alex Celani
// File: led.go
// Revn: 06-27-2022  1.3
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
// 06-23-2022: updated to produce three different colors
// 06-26-2022: commented
// 06-27-2022: added flag package
//
// =============================================================================

package main

import (
    "net"       // ResolveTCPAddr, conn.Write,Read,Close
                // ListenTCP, listener.Accept
    "os"        // Args, Stderr, Exit
    "fmt"       // Fprintf, Println
    "strings"   // ToLower
    "github.com/stianeikeland/go-rpio"      // talk to RasPi pins
    "flag"      // Parse, String, Int, Bool
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

    var buf [512]byte   // declare large byte array, store messages

    // iterate forever to always read over connection
    for {
        // read n bytes from connection into buffer
        n, err := conn.Read( buf[0:] )
        if err != nil { // erroring on read will simply leave the 
            return      // function so it can start again later
        }

        if *verbose {
            // print recv'd message
            // string() only works on byte SLICES so [:] is required
            fmt.Println( "recv: ", string( buf[:n] ) )
        }

        // split command into words for ease of parsing
        command := strings.Split( string( buf[:n] ), " " )

        if command[1] != "led" {    // confirm message belongs here
            os.Exit( 2 )            // if not, exit
        }

        var resp string             // declare response variable


        // TODO convert to switches?
        // len is 3 for "asK" commands
        if len( command ) == 3 {
            if command[2] == "color" {
                // if user wants color, return color
                resp = color
            } else if command[2] == "status" {
                // if user wants status, return status
                resp = status
            } else {
                // there are only those two fields
                resp = "usage unknown"
            }
        // len is 4 for tell commands
        } else if len( command ) == 4 {
            // if user wants to set color
            if command[2] == "color" {
                color = command[3]      // set current color
                resp = color            // return current color
                switch color {
                    case "blue":        // does user want blue?
                        // turn blue on, turn rest off
                        pinB.High()
                        pinR.Low()
                        pinG.Low()
                    case "red":         // does user want red?
                        // turn red on, turn rest off
                        pinB.Low()
                        pinR.High()
                        pinG.Low()
                    case "green":       // does user want green?
                        // turn green on, turn rest off
                        pinB.Low()
                        pinR.Low()
                        pinG.High()
                    default:            // color unknown
                        // turn off
                        pinB.Low()
                        pinR.Low()
                        pinG.Low()
                }
            // if user wants to set status
            } else if command[2] == "status" {
                status = command[3]     // set current status
                resp = status           // return current status
                switch status {
                    case "on":          // does user want on?
                        // default to red
                        pinB.Low()
                        pinR.High()
                        pinG.Low()
                    default:            // anything else, like off
                        // default to off
                        pinB.Low()
                        pinR.Low()
                        pinG.Low()
                }
            // user entered something else entirely
            } else {
                resp = "usage unknown"
            }
        // no other lengths are valid
        } else {
            os.Exit( 2 )        // exit in this case
        }

        if *verbose {
            fmt.Println( "answer: ", resp )     // print response
        }

        // write that response back to original client
        _, err = conn.Write( []byte( resp ) )
        if err != nil { // erroring on write will simply leave the
            return      // function so it can start again later
        }

        if *verbose {
            fmt.Println( "sent: ", resp )   // print response
        }
    }
}


//var params = make( map[string]string )
//var states = make( map[string]string )
var pinR rpio.Pin           // declare red led pin
var pinG rpio.Pin           // declare green led pin
var pinB rpio.Pin           // declare blue led pin
var color string            // declare color variable
var status string           // declare status variable

var (       // declare flag variables
    lr *int
    lg *int
    lb *int
    verbose *bool
    ip *string
)


func main() {

    // declare command line input variables
    lr = flag.Int( "lr", 11, "GPIO pin for red LED" )
    lg = flag.Int( "lg", 9, "GPIO pin for green LED" )
    lb = flag.Int( "lb", 25, "GPIO pin for blue LED" )
    verbose = flag.Bool( "v", false, "verbose printing" )
    ip = flag.String( "ip", ":1202", "ip and port of self" )
    flag.Parse()

    // initialize pins to GPIO 11, 9, and 25
    pinR = rpio.Pin( *lr )
    pinG = rpio.Pin( *lg )
    pinB = rpio.Pin( *lb )

	if err := rpio.Open(); err != nil {
        // on error, print error and exit
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pins to output mode
	pinR.Output()
	pinG.Output()
	pinB.Output()

    // "resolve" ip & host according to TCP rules
    tcpAddr, err := net.ResolveTCPAddr( "tcp", *ip )
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

