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
//*06-22-2022: bug fixing and comments
//
// =============================================================================

package main

import (
    "net"       // ResolveTCPAddr, DialTCP, conn.Write,Read,Close
                // ListenTCP, listener.Accept
    "os"        // Exit, Stderr, Stdin, Args
    "fmt"       // Println, Fprintf
    "github.com/stianeikeland/go-rpio"      // talk to RasPi pins
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

        // toggle pin on successful receipt
        pin.Toggle()

    }
}


// declare pin
var (
	// Use mcu pin 10, corresponds to physical pin 19 on the pi
	pin = rpio.Pin(10)
)


func main() {

	// Open and map memory to access gpio, check for errors
    // inline open gpio pin and check the error var
	if err := rpio.Open(); err != nil {
        // on error, print error and exit
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()

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

