# Command line utility to send binary data to a serial port.

import argparse
import serial.tools.list_ports
import sys


def main():
    """ Main function. """

    # Handle command line arguments.
    parser = argparse.ArgumentParser('serialtest', description='Serial port test writer.')
    parser.add_argument('-l', '--list', action='store_true', required=False, help='List available serial port devices.')
    parser.add_argument('-s', '--serial', type=str, required=False, help='Serial port device.')

    args = parser.parse_args()

    if (hasattr(args, 'list') and args.list):
        # List available serial port devices and exit.
        print('Serial ports found:')
        ports = list(serial.tools.list_ports.comports())
        for p in ports:
            print(p)

        sys.exit(0)

    print('Serial port test.')

    # Default values.
    serial_device_name = 'COM7'
    
    if (hasattr(args, 'serial') and args.serial):   serial_device_name = args.serial

    # Open serial port.
    ser = serial.Serial(serial_device_name)
    ser.baudrate = 125000
    ser.parity = serial.PARITY_EVEN
    ser.bytesize = 8
    ser.stopbits = 1
    single_byte = bytearray(1)
    
    print('Enter hex bytes.')
    print('Type quit to exit.')

    # Main control loop.
    while True:
        # Get some input.
        line = input()

        if line == 'quit':
            sys.exit(0)

        value = int('0x' + line, base=16)
        single_byte[0] = value
        ser.write(single_byte)
    

if __name__ == '__main__':
    main()
