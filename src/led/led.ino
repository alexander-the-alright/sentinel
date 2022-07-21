// ===================================================================
// Auth: Alex
// File: led.ino
// Revn: 07-20-2022  1.0
// Func: End node, to run on ESP8266. Operate as server, process
//       incoming messages, react to them, send responses
// 
// TODO: incorporate interrupt example for builtin led
// ===================================================================
// CHANGE LOG
// -------------------------------------------------------------------
// 07-07-2022: found online
// 07-12-2022: changed client.write() to client.println(), to fix
//               bug about sending Strings
//             investigated strcmp, pinmode, etc
//             cleaned up, commented setup()
// 07-13-2022: commented
//             investigated faster ways to take input; no dice
//             toggle builtin led on receipt
// 07-14-2022: commented loop()
//             wrote prse()
//             began added [list esp] message, lst()
//             wrote bones for refresh(), to update led
// 07-19-2022: added led to circuit, wrote refresh()
//             moved check for /list/ above check for second word
//               to fix /list/ not sending bug
//             added default error value for response variable
//             added call to refresh at the end of setup() to init
//               led state
//*07-20-2022: /tell/ commands set status to on
//             commented
//             changed name of file from democode to led.ino
//             changed NAME var from "esp" to "led"
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
#define SendKey 0  // button to send data Flash BTN on NodeMCU

int port = 1210;  // port of server
WiFiServer server(port);

IPAddress self( 192, 168, 1, 69 );  // ip address of self
IPAddress gate( 192, 168, 1, 1 );   // ip address of gateway
IPAddress subn( 255, 255, 0, 0 );   // subnet mask
IPAddress pDNS( 8, 8, 8, 8 );       // primary DNS server ip
IPAddress sDNS( 8, 8, 4, 4 );       // secondary DNS server ip

// credentials for WiFi Network
const char *ssid = "waystation";
const char *password = "homophobiaR0X";

// pin definitions
int pinR = 16;      // GPIO16 --> D0 on the board
int pinG = 5;       // GPIO5  --> D1 on the board
int pinB = 4;       // GPIO4  --> D2 on the board

//int pinBI = 2;      // GPIO2  --> D4 on the board, built-in LED

// state variables
bool statusRGB = true;    // init, true -> off
String color = "red";     // init color red

bool statusBI = true;     // init, true -> off

String NAME = "led";      // variable to semi-hardcode program name

// ===================================================================
//                    Power on setup
// ===================================================================
void setup() {
  pinMode( pinBI, OUTPUT );         // set built-in LED to output
  digitalWrite( pinBI, statusBI );  // initialize it off
  // this really dumb fucking board is active low, I am fucking weeping

  // init rgb led pins as output
  pinMode( pinR, OUTPUT );
  pinMode( pinG, OUTPUT );
  pinMode( pinB, OUTPUT );

  // init rgb led as off ( active low, the headache )
  digitalWrite( pinR, HIGH );
  digitalWrite( pinG, HIGH );
  digitalWrite( pinB, HIGH );
  
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

  refresh();      // update led on setup
}


void refresh() {
  // print led status
  Serial.print( "color: " );
  Serial.println( color );
  Serial.print( "status: " );
  Serial.println( statusRGB );

  // declare all pins off
  digitalWrite( pinR, HIGH );
  digitalWrite( pinG, HIGH );
  digitalWrite( pinB, HIGH );

  // if led should be on
  // if off, led pins are already off
  if( !statusRGB ) {
    if( color == "red" ) {            // if red
      digitalWrite( pinR, LOW );      // turn on only red
    } else if( color == "green" ) {   // if green
      digitalWrite( pinG, LOW );      // turn on only green
    } else if( color == "blue" ) {    // if blue
      digitalWrite( pinB, LOW );      // turn on only blue
    } else if( color == "cyan" ) {    // if 'cyan'
      digitalWrite( pinG, LOW );      // cyan = B + G
      digitalWrite( pinB, LOW );      
    } else if( color == "yellow" ) {  // if yellow
      digitalWrite( pinR, LOW );      // yellow = R + G
      digitalWrite( pinG, LOW );
    } else if( color == "magenta" ) { // if magenta
      digitalWrite( pinR, LOW );      // magenta = R + B
      digitalWrite( pinB, LOW );
    } else if( color == "white" ) {   // if white
      digitalWrite( pinR, LOW );      // all on for white
      digitalWrite( pinG, LOW );
      digitalWrite( pinB, LOW );
    }
  }
}


String lst() {
  //craft list string and return
  String resp = "";
  
  resp = resp + "color ^v\n";
  resp = resp + "       blue\n";
  resp = resp + "       cyan\n";
  resp = resp + "       green\n";
  resp = resp + "       magenta\n";
  resp = resp + "       red\n";
  resp = resp + "       white\n";
  resp = resp + "       yellow\n";
  resp = resp + "   status ^v\n";
  resp = resp + "       on\n";
  resp = resp + "       off\n";
  resp = resp + "   off ^\n";
  resp = resp + "   on ^\n";
  resp = resp + "   list v";
  
  return resp;
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
          // if user asks for color
          if( argument == "color" ) {
            response = color;     // set response to current color
            // if user wants status
          } else if( argument == "status" ) {
            // if status is true, led is off, else on
            response = statusRGB ? "off" : "on";
          } else {    // doesn't want status or color, arg unknown
            response = "field unrecognized";
          }
        // if user is telling ( some args can be one word )
        } else if( tell ) {
          // only args can be on and off
          if( argument == "off" || argument == "on" ) {
            // set status, and set return string
            statusRGB = argument == "off";
            response = argument;
          } else {    // if user didn't tell on or off
            // print argument
            Serial.print( "argument: " );
            Serial.println( argument );
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

                // if subcommand is color
                if( subcom == "color" ) {
                  // and subarg matches a supported color
                  if( subarg == "red" || subarg == "magenta"
                   || subarg == "blue" || subarg == "cyan"
                   || subarg == "green" || subarg == "yellow"
                   || subarg == "white" ) {
                    // set color to the input color
                    color = subarg;
                    // set response to the same color
                    response = color;
                    // turn on led
                    statusRGB = false;
                  } else {    // if color isn't supported
                    // say so, don't change anything
                    response = "argument unrecognized";
                  }
                // if subcommand is for status
                } else if( subcom == "status" ) {
                  // if subarg is for either on or off
                  if( subarg == "off" || subarg == "on" ) {
                    // set status to input
                    statusRGB = subarg == "off";
                    // if status is true, led is off, else on
                    response = statusRGB ? "off" : "on";
                  } else {    // if subarg isn't on or off
                    // say so, don't change anything
                    response = "argument unrecognized";
                  }
                } else {  // command is neither color nor status
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
          }
        } else {    // if user is neither asking nor telling
                    // also not list
          // say so, don't change anything
          response = "command unrecognized";
        }
      }
    }
  }
  
  refresh();    // update the led
  // print response
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
