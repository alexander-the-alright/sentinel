// =============================================================================
// Auth: Alex Celani
// File: colonel.go
// Revn: 06-20-2022  1.0
// Func: Send message to another machine and then receive a single
//       response. Extremely tight bounds, not robust at all
//
// TODO: comment
// =============================================================================
// CHANGE LOG
// -----------------------------------------------------------------------------
// 05-04-2022: init
//*06-20-2022: changed name to reflect first draft of colonel.go
//
// =============================================================================

package main

import (
    "net"               // ResolveTCPAddr, DialTCP, conn.Write,Read
    "os"                // Exit, Stdin, Stderr, Args
    "fmt"               // Println, Print, Fprintf
    "bufio"             // NewScanner, NewScanner.Scan,Text
    "strings"           // ToLower
    "time"              // Now, Sub
)


// quick function to take user input
func input( ps1 string ) string {
    fmt.Print( ps1 )                            // print prompt
    scanner := bufio.NewScanner( os.Stdin )     // link to stdin
    scanner.Scan()                              // pull data
    return strings.ToLower( scanner.Text() )    // return lower text
}


// quick error checking
func check( err error ) {
    // if user checks an error, it better be nil
    if err != nil {
        // put error in string, print in stderr
        fmt.Fprintf( os.Stderr, "Fatal error: %s", err.Error() )
        os.Exit( 2 )        // quit
    }
}


func main() {

    // if len is not 2 ( ./name host:port ), print usage and quit
    if len( os.Args ) != 2 {
        fmt.Fprintf( os.Stderr, "Usage: %s host:port", os.Args[0] )
        os.Exit( 1 )
    }

    service := os.Args[1]   // capture ip address and host

    // "resolve" ip & host according to TCP rules
    tcpAddr, err := net.ResolveTCPAddr( "tcp", service )
    check( err )    // check error

    // "dial" ( establish connection ) to destination ip & port
    // according to TCP rules
    // laddr = nil -> local address is localhost
    conn, err := net.DialTCP( "tcp", nil, tcpAddr )
    check( err )    // check error

    var in string = input( ">> " )      // get user input

    start := time.Now()                 // start round trip timer

    // convert user input to bytes and send over connection
    _, err = conn.Write( []byte( in ) )
    check( err )    // check error

    var buf [512]byte                   // init byte array
    // read from connection into byte array
    // amount of bytes read stored in n
    //n, err := conn.Read( buf[0:] )
    n, err := conn.Read( buf[:] )
    check( err )    // check error

    t := time.Now()                     // stop round trip timer

    // convert read bytes into string and print
    fmt.Println( string( buf[:n] ) )

    fmt.Println()       // newline to clean up a little
    fmt.Println( "rtt: ", t.Sub( start ) )      // print elapsed time


    os.Exit( 0 )        // exeunt

}

