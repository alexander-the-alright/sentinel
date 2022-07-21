// =============================================================================
// Auth: Alex Celani
// File: led.go
// Revn: 07-11-2022  4.0
// Func: receive data as an endpoint, send response back to middleman
//       ONLY A SIMULATION FOR THE REAL THING
//
// TODO: implement rainbow
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
//             began work on wrapper functions for led
// 06-28-2022: rewrote wrapper function for color, tell()
// 06-29-2022: wrote wrapper function for status, ask()
//             added /list/ to options in ask()
//*07-04-2022: rewrote /color/ in tell() just a little bit, to work
//                  common CATHODE ( so mad ) RGB leds
//             added support for white, cyan, yellow, and magenta
//             finished comments for tell() and ask() and even simple
//                  parsing stuff in handleClient()
//*07-05-2022: added support for kill messages
//*07-11-2022: added /list/ as field that calls ask()
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
        case "color":   // if user asks for color, get color
            response = color
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

    // initiate all pins as off
    pinR.High()
    pinG.High()
    pinB.High()

    // declare response variable
    var response string

    // XXX this could be done with a map instead
    // what is the first word?
    switch comm[0] {
        case "color":       // if user trying to change color
            color = comm[1]     // set current color to input
                                // FIXME on failure, this is wrong
                                // is that ok?
            response = color    // preemptively set response to color
            status = "on"       // preemptively set status
            switch comm[1] {
                case "red":     // user wants red, turn red on
                    pinR.Low()
                case "green":   // user wants green, turn green on
                    pinG.Low()
                case "blue":    // user wants blue, turn blue on
                    pinB.Low()
                case "magenta": // user wants blue and red
                    pinR.Low()
                    pinB.Low()
                case "yellow":  // user wants red and green
                    pinR.Low()
                    pinG.Low()
                case "cyan":    // user wants blue and green
                    pinB.Low()
                    pinG.Low()
                case "white":   // user wants all leds
                    pinR.Low()
                    pinG.Low()
                    pinB.Low()
                default:        // user asks for else, bail out
                    response = "color unknown"
                    status = "off"
            }
        case "status":      // if user trying to change status
            status = comm[1]    // preemptively set status
            switch comm[1] {
                case "on":      // if trying to turn led on
                                // re enter function as if trying to
                                // set current color
                    response = tell( []string{ "color", color } )
                case "off":     // leave leds off, set response to off
                    response = comm[1]
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

        if command[1] != "piled" {  // confirm message belongs here
            os.Exit( 2 )            // if not, exit
        }

        // initiate response variable
        var resp string

        switch command[0] {     // is user getting or setting?
            case "ask":         // if getting
                // get, store response in variable to be sent back
                resp = ask( command[2:] )
            case "list":        // is user getting list?
                // get, store response in variable to be sent back
                resp = ask( command )
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
var pinG rpio.Pin           // declare green led pin
var pinB rpio.Pin           // declare blue led pin
var color string = "red"    // declare color variable
var status string = "off"   // declare status variable

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
    ip = flag.String( "ip", ":1212", "ip and port of self" )
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

    // Initialize output to "off"
    pinR.High()
    pinG.High()
    pinB.High()

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

