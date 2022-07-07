// =============================================================================
// Auth: Alex Celani
// File: led2.go
// Revn: 07-05-2022  1.0
// Func: receive data as an endpoint, send response back to middleman
//
// TODO:
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
//*07-05-2022: copied from led.go
//
// =============================================================================

package main

import (
    "net"           // ResolveTCPAddr, conn.Write,Read,Close
                    // ListenTCP, listener.Accept
    "os"            // Args, Stderr, Exit
    "fmt"           // Fprintf, Println
    "strings"       // ToLower
    "github.com/stianeikeland/go-rpio"      // talk to RasPi pins
    "flag"          // Parse, String, Int, Bool
    "io/ioutil"     // ReadFile
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


// quick function to return LED fields
func ask( comm []string ) string {

    // declare response variable
    var response string

    // XXX this could be done with a map instead
    // what is the first word?
    switch comm[0] {
        case "status":  // if user asks for status, get status
            response = status
        case "list":    // if user wants a list of fields
            // open file containing fields
            // XXX considering using this file to populate future map
            file, err := ioutil.ReadFile( os.Args[0] + "list" )
            check( err )    // check error
            // cast entire file to string and make returnable
            response = string( file )
        default:        // if user asks for something unknown
            response = "field unknown"
    }

    // return
    return response

}


// quick function to change LED colors
func tell( comm []string ) string {

    // declare response variable
    var response string

    // XXX this could be done with a map instead
    // what is the first word?
    switch comm[0] {
        case "status":      // if user trying to change status
            status = comm[1]    // preemptively set status
            response = status
            switch comm[1] {
                case "on":      // if trying to turn led on
                    pinR.High()
                case "off":     // leave leds off, set response to off
                    pinR.Low()
                default:        // user asks for else, bail out
                    response = "status unknown"
            }
        // user wants to stop running
        case "kill", "quit", "end", "die", "stop":
            response = "kill"
        default:            // user enters incorrect command
            response = "command unknown"
    }

    return response

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

        if command[1] != "led2" {   // confirm message belongs here
            os.Exit( 2 )            // if not, exit
        }

        // TODO differentiate ask commands from tell commands

        // initiate response variable
        var resp string

        switch command[0] {     // is user getting or setting?
            case "ask":         // if getting
                // get, store response in variable to be sent back
                resp = ask( command[2:] )
            case "tell":        // if setting
                // set, store response in variable to be sent back
                resp = tell( command[2:] )
            default:            // if neither, bail out
                resp = "command form unknown"
        }

        if *verbose {   // XXX verbose print
            fmt.Println( "answer: ", resp )     // print response
        }

        // write that response back to original client
        _, err = conn.Write( []byte( resp ) )
        if err != nil { // erroring on write will simply leave the
            return      // function so it can start again later
        }

        if *verbose {   // XXX identical verbose print
            fmt.Println( "sent: ", resp )   // print response
        }

        if resp == "kill" {     // if user sent back kill
            os.Exit( 1 )        // end process
        }
    }
}


//var params = make( map[string]string )
//var states = make( map[string]string )
var pinR rpio.Pin           // declare red led pin
var status string = "off"   // declare status variable

var (       // declare flag variables
    led *int
    verbose *bool
    ip *string
)


func main() {

    // declare command line input variables
    led = flag.Int( "led", 24, "GPIO pin for red LED" )
    verbose = flag.Bool( "v", false, "verbose printing" )
    ip = flag.String( "ip", ":1203", "ip and port of self" )
    flag.Parse()

    // initialize pins to GPIO 24
    pinR = rpio.Pin( *led )

	if err := rpio.Open(); err != nil {
        // on error, print error and exit
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pins to output mode
	pinR.Output()

    // Initialize output to "off"
    pinR.High()

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

