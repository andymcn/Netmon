/* Netmon LED drive board.

We display a grid of 20 columns of 4 rows of LEDs.

Commands are received over a serial link to turn on and off individual LEDs. Each byte is a complete command. The most
significant bits tells us to turn off (0) or on (1) one LED. The remaining 7 bits specify the LED.

At start of day we turn on all LEDs for a second or two to check they function. Then all LEDs are turned off until
we're told otherwise.

*/

#include "global.h"
#include <avr/interrupt.h>
#include <avr/wdt.h>


#define COLUMN_COUNT    20

// Pinout map for column drive.
// Columns connect to LED cathodes, so the values here are active low.
static struct {
    uint8_t a;
    uint8_t c;
    uint8_t d;
} _columnMap[COLUMN_COUNT] = {
    { 0xFF, 0xFF, 0xE0 },   { 0xFF, 0xFF, 0xD0 },   { 0xFF, 0xFF, 0xB0 },   { 0xFF, 0xFF, 0x70 },   // 0..3.
    { 0xFE, 0xFF, 0xF0 },   { 0xFD, 0xFF, 0xF0 },   { 0xFB, 0xFF, 0xF0 },   { 0xF7, 0xFF, 0xF0 },   // 4..7.
    { 0xEF, 0xFF, 0xF0 },   { 0xDF, 0xFF, 0xF0 },   { 0xBF, 0xFF, 0xF0 },   { 0x7F, 0xFF, 0xF0 },   // 8..11.
    { 0xFF, 0xFE, 0xF0 },   { 0xFF, 0xFD, 0xF0 },   { 0xFF, 0xFB, 0xF0 },   { 0xFF, 0xF7, 0xF0 },   // 12..15.
    { 0xFF, 0xEF, 0xF0 },   { 0xFF, 0xDF, 0xF0 },   { 0xFF, 0xBF, 0xF0 },   { 0xFF, 0x7F, 0xF0 }    // 16..19.
};


// State.
static uint8_t _rows[COLUMN_COUNT]; // LED states. Given as values to drive rows for each column.
static uint8_t _column; // The current column we're driving.
static uint16_t _start_check;   // Remaining ticks for initial all LEDs on.


// Disable watchdog timer on startup so it doesn't continually reset us after a reboot.
void wdt_init(void) __attribute__((naked)) __attribute__((section(".init3")));

void wdt_init(void)
{
    MCUSR = 0;
    wdt_disable();
}


void init(void)
{
    // Initialise hardware and state.

    // All GPIOs used are outputs.
    DDRA = 0xFF;  // A0..7 drive columns 4..11.
    DDRB = 0xFF;  // B0..3 drive rows. B4..7 unused.
    DDRC = 0xFF;  // C0..7 drive columns 12..19.
    DDRD = 0xFF;  // D0..1 are serial port. D2..3 unused. D4..7 drive columns 0..3.

    // Set all LED drive pins to inactive.
    PORTB = 0;
    PORTA = 0xFF;
    PORTC = 0xFF;
    PORTD = 0xF0;

    // Initialise UART.
    // 125kbaud, 8 bit data, even parity, 1 stop bit.
    // We only need RX interupts, we never transmit.
    UCSR0A = 0;
    UCSR0B = 0x98;
    UCSR0C = 0x26;
    UBRR0H = 0;
    UBRR0L = 3;

    // Set up 1kHz tick timer.
    TCCR0A = 2; // Set up CTC0.
    TCCR0B = 3; // ...
    TCNT0 = 0;  // ...
    OCR0A = 124; // ~1kHz tick rate (8MHz / 64 prescale / 125 (OCR0A + 1))
    TIMSK0 = 2; // Enable CTC0 interrupts.

    // Initialise LED state.
    for(uint8_t i = 0; i < COLUMN_COUNT; i++)
        _rows[i] = 0;

    // Start with column 0.
    _column = 0;

    // Turn on all LEDs for around a second.
    _start_check = 1000;
}


int main(void)
{
    init();

    // Enable interrupts
    sei();

    // Main program loop. Everything is run from interrupts.
    while(true);
}


// Tick interrupt.
// Called roughly every 1 millisecond.
ISR(TIMER0_COMPA_vect)
{
    // First stop driving the LEDs in the current column.
    PORTB = 0;

    // Move on to the next grid column.
    _column++;
    if(_column >= COLUMN_COUNT)  // Wrap to start of grid.
        _column = 0;

    PORTA = _columnMap[_column].a;
    PORTC = _columnMap[_column].c;
    PORTD = _columnMap[_column].d;

    // Now drive the LEDs for the new column.
    PORTB = _rows[_column];

    if(_start_check > 0) {
        // We're still displaying all LEDs at start of day.
        PORTB = 0xF;
        _start_check--;
    }
}


// Byte received interrupt.
ISR(USART0_RX_vect)
{
    // Get command and decode.
    uint8_t command = UDR0;
    bool on = (command & 0x80) != 0;
    uint8_t led = command & 0x7F;

    // LEDs are numbered from 0, down the column, then across the rows.
    uint8_t column = led >> 2;
    uint8_t row = led & 3;
    uint8_t row_value = 1 << row;

    // Change state of bit in row for specified LED. Row values are active high.
    if(on) {
        _rows[column] |= row_value;
    } else {
        _rows[column] &= ~row_value;
    }
}
