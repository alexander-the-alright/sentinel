// ===================================================================
// Auth: Alex
// File: led.ino
// Revn: 07-27-2022  1.0
// Func: End node, to run on ESP8266. Operate as server, process
//       incoming messages, react to them, send responses
// 
// TODO: incorporate interrupt example for builtin led
//       update to match lst()
// ===================================================================
// CHANGE LOG
// -------------------------------------------------------------------
// 07-25-2022: pulled from led.ino
// 07-26-2022: added /fill/ and units to /fill/ and /dist/
//             made distance calc values into constants
//             comments
//             rewrote lst() a little bit
// 07-27-2022: made quick getDepth() function to fix bad unit bug
//             filled in /tell/
//             commented
//
// ===================================================================

// ESP8266 version of Wifi.h
// API Documentation located at:
// 
// Pinout located at:
// https://microcontrollerslab.com/led-blinking-using-esp8266-nodemcu/#ESP8266_Pinout_in_Arduino_IDE
#include <ESP8266WiFi.h>
// documentation at
// https://arduino.cc/reference/en/libraries/esp8266timerinterrupt/
#include <ESP8266TimerInterrupt.h>

// real shit, not certain why this is important
#define SendKey 0   // button to send data Flash BTN on NodeMCU
#define Vs 0.034    // speed of sound [cm/s]
#define CMtoIN 0.393701   // convert cm to inches [in/cm]

int port = 1211;  // port of server
WiFiServer server(port);

IPAddress self( 192, 168, 1, 30 );  // ip address of self
IPAddress gate( 192, 168, 1, 1 );   // ip address of gateway
IPAddress subn( 255, 255, 0, 0 );   // subnet mask
IPAddress pDNS( 8, 8, 8, 8 );       // primary DNS server ip
IPAddress sDNS( 8, 8, 4, 4 );       // secondary DNS server ip

// credentials for WiFi Network
const char *ssid = "waystation";
const char *password = "homophobiaR0X";

// pin definitions
int trig = 10;      // GPIO10 --> SDD3 on the board
int echo = 9;       // GPIO9  --> SDD2 on the board

// state variables
// permanently set to [cm]
int depth = 30;             // depth of cat food container [cm]
bool inch = false;          // keep track of units
String NAME = "range";      // variable to semi-hardcode program name

// ===================================================================
//                    Power on setup
// ===================================================================
void setup() {
  // this really dumb fucking board is active low, I am fucking weeping

  // init rgb led pins as output
  pinMode( trig, OUTPUT );
  pinMode( echo, INPUT );
  
  Serial.begin( 115200 );           // board works at 115,200 [baud]
  pinMode( SendKey, INPUT_PULLUP ); // Btn to send data (NMC NMM)

  WiFi.mode( WIFI_STA );            // set mode to station
  // set static ip address ( only way it would let me do it )
  WiFi.config( self, gate, subn, pDNS, sDNS );
  WiFi.begin( ssid, password );     // connect to wifi
  
  // Serial.prints all considered verbose, because you can only see
  // them if you're trying to
  // wait for connection  
  Serial.println( "Connecting to Wifi" );
  // iterate until connected
  while( WiFi.status() != WL_CONNECTED ) {
    delay( 1000 );
    Serial.print( "." );
  }

  // print connection status
  Serial.println();
  Serial.print( "Connected to " );
  Serial.println( ssid );

  server.begin();       // bring server online
  // print server address
  Serial.print( "On at: " );
  Serial.print( WiFi.localIP() );  
  Serial.print( ":" );
  Serial.println( port );

}


String lst() {
  //craft list string and return
  String resp = "";
  
  resp = resp + "dist v\n";
  resp = resp + "   fill v\n";
  resp = resp + "   depth ^v\n";
  resp = resp + "       /int/\n";
  resp = resp + "   unit ^v\n";
  resp = resp + "       in[ch]\n";
  resp = resp + "       cm";
  
  return resp;
}


int getDist() {
  long duration, distance;
  
  digitalWrite( trig, LOW );      // init pin low
  delayMicroseconds( 2 );         // probably not exactly necessary
  digitalWrite( trig, HIGH );     // start trigger
  delayMicroseconds( 10 );        // hold a little bit
  digitalWrite( trig, LOW );      // end trigger

  // get time until pulse in?
  duration = pulseIn( echo, HIGH );
  Serial.print( "duration (raw): " );
  Serial.println( duration );
  
  Serial.print( "half TOF: " );
  Serial.println( duration / 2 );
  // calculate distance from time of flight (TOF)
  //           duration / 2   -> TOF is there and back, get there
  //                          * 0.034  -> cm / us
  distance = ( duration / 2 ) * Vs;

  Serial.print( "dist: " );
  Serial.println( distance );
  
  Serial.print( "string: " );
  Serial.println( String( distance ) );

  return distance;
}


String getDepth() {
  // if units are inches, convert, else scale by 1
  int deepness = inch ? depth * CMtoIN : depth;
  // convert to string
  String response = String( deepness );
  // append correct units
  response += inch ? " [in]" : " [cm]";
  
  return response;
}


// ===================================================================
String prse( String command ) {
  int amnt = 0;       // keep track of amount of spaces
  int len = command.length();   // capture length

  // quick variables to keep track of command
  bool ask = false;
  bool tell = false;
  bool list = false;

  // init failure string
  String response = "error: command unrecognized";

  // iterate over each letter of incoming message
  for( int i = 0; i < command.length(); i++ ) {
    // when letter is space...
    if( command[i] == ' ' ) {
      amnt++;   // increment count for number of spaces
      // on the first space
      if( amnt == 1 ) {
        // print first command
        Serial.print( "ask/tell: " );
        Serial.println( command.substring( 0, i ) );
        // snatch bools for various commands
        ask = command.substring( 0, i ) == "ask";
        tell = command.substring( 0, i ) == "tell";
        list = command.substring( 0, i ) == "list";

        // print bools
        Serial.println( ask );
        Serial.println( tell );
        Serial.println( list );

        // seemed hacky but it fuckin works
        // if command was list, get list and return it
        if( list ) {
          Serial.println( "LIST" );
          response = lst();
          Serial.println( response );
        }
      // on the second space
      } else if( amnt == 2 ) {
        // capture everything after third word
        String argument = command.substring( i + 1, len );
        // if user is asking ( should make argument one word )
        if( ask ) {
          // get distance before doing anything
          // not always necessary, but isn't resource intensive
          int distance = getDist();
          if( argument == "dist" ) {            // user wants dist
            // if units are inches, convert, else scale by 1
            distance *= inch ? CMtoIN : 1;
            response = String( distance );  // set return var
            // append correct units
            response += inch ? " [in]" : " [cm]";
          } else if( argument == "fill" ) {     // user wants fill
            // if distance is smaller than depth, distance / depth will
            // be 0, because int. scale up before taking fraction to find
            // percent empty space
            // sub from 100 to find the actual fill percent
            distance = 100 - ( distance * 100 ) / depth;
            response = String( distance ) + "%";  // append unit
          } else if( argument == "depth" ) {    // user wants depth
            response = getDepth();
          } else if( argument == "unit" ) {     // user wants units
            response = inch ? "[in]" : "[cm]";
          } else {    // doesn't want status or color, arg unknown
            response = "field unrecognized";
          }
        // node doesn't work with tell
        } else if( tell ) {
          response = "/tell/ is under construction";
          // capture number of letters in argument
          int arglen = argument.length();
          // bool to keep track of if a space was seen
          // if there was no space seen, arg is very incorrect
          bool space = false;
          // iterate over characters in argument
          for( int j = 0; j < arglen; j++ ) {
            // if letter is a space
            if( argument[j] == ' ' ) {
              // mark space as true
              space = true;
              // grab everything before the space, subcommand
              String subcom = argument.substring( 0, j );
              // grab everything after the space, subargument
              String subarg = argument.substring( j + 1, arglen );

              // user is redefining food depth
              if( subcom == "depth" ) {
                if( subarg.toInt() != 0 ) {
                  depth = subarg.toInt();
                  depth /= inch ? CMtoIN : 1;
                  response = getDepth();
                } else {
                  response = "depth argument must be an int";
                }
              // user is redefining units
              } else if( subcom == "unit" ) {
                // check if subarg is valid
                if( subarg == "cm" ) {
                  inch = false;   // false for cm
                  response = getDepth();
                } else if( subarg == "in" || subarg == "inch"
                        || subarg == "inches" ){
                  inch = true;    // true for inches
                  response = getDepth();
                } else {    // if input is invalid, don't change
                  // say so, don't change anything
                  response = "argument unrecognized";
                }
              } else {  // command is neither depth nor unit
                // say so, don't change anything
                response = "command unrecognized";
              }
              // break out once space is found
              // no need to keep iterating once space is found
              break;
            }
          }
          // if space was never found, argument was very wrong
          if( !space ) {
            // say so, don't change anything
            response = "/" + argument + "/ invalid";
          }
        } else {    // if user is neither asking nor telling
                    // also not list
          // say so, don't change anything
          response = "command unrecognized";
        }
      }
    }
  }
  
  Serial.print( "resp IN:" );
  Serial.println( response );
  return response;    // return response
}


// ===================================================================
//                    Loop
// ===================================================================
void loop() {
  // returns client one exists
  WiFiClient client = server.available();
  
  if( client ) {      // if server.available() finds client
    if( client.connected() ) {    // client has connected
      Serial.println( "Client Connected" );
    }

    String response = "";   // init response variable

    // for as long as client stays connected
    while( client.connected() ) {
      // for as long as there are bytes to be read, read em
      while( client.available() > 0 ) {
        // read data from the connected client
        String command = client.readStringUntil( '\n' );
        //digitalWrite( pinBI, statusBI = !statusBI );
        Serial.println( command );    // print to serial monitor

        int count = 0;
        int previous = 0;   // keep track of spaces
        // iterate over string
        for( int i = 0; i <= command.length(); i++ ) {
          // if char is space or end of string
          if( command[i] == ' ' || command[i] == '\0') {
            // print each word in command (except last one)
            Serial.println( command.substring( previous, i ) );
            if( count == 1 ) {    // check second word for name
              // if command was meant for me...
              if( command.substring( previous, i ) == NAME ) {
                // parse command
                response = prse( command );
                Serial.print( "resp OUT:" );
                Serial.println( response );
                //response = "not just yet";
              } else {  // tell user message doesn't belong here
                // this shouldn't be possible, but
                response = "node mismatch";
              }
              // once response is set, leave loop and send
              break;
            }
            // set previous to AFTER space, to capture ONLY the word
            previous = i + 1;
            count++;    // increment count of spaces
          }
        }
        // print rest of message to monitor
        Serial.println( command.substring( previous, command.length() ) );
        client.println( response );     // send entire command to client
      }
    }
    client.stop();    // stop client when done reading
    Serial.println( "Client disconnected" );    // print
  }
}
//=======================================================================
