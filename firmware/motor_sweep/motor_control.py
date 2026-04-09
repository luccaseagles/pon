#!/usr/bin/env python3
"""
Motor ramp controller — hold w/s to ramp, release to stop.
Uses raw terminal + key-repeat timing. No extra libs beyond pyserial.
Usage: python3 motor_control.py [port]
"""

import sys
import tty
import termios
import select
import time
import threading
import serial
import serial.tools.list_ports

# Tuning
RAMP_STEP   = 200    # speed added per tick while held
DECAY_STEP  = 300    # speed removed per tick on release
TICK_S      = 0.005  # 50 Hz
HELD_WINDOW = 0.08  # seconds — key counts as held if seen within this window
MAX_SPD     = 255


def find_port():
    if len(sys.argv) > 1:
        return sys.argv[1]
    for p in serial.tools.list_ports.comports():
        if "ACM" in p.device or "usbmodem" in p.device:
            return p.device
    ports = [p.device for p in serial.tools.list_ports.comports()]
    if not ports:
        sys.exit("No serial ports found.")
    if len(ports) == 1:
        return ports[0]
    for i, p in enumerate(ports):
        print(f"  {i}: {p}")
    return ports[int(input("Select port: "))]


last_seen = {"w": 0.0, "s": 0.0}
quit_flag = threading.Event()


def key_reader():
    """Reads raw keypresses and stamps their last-seen time."""
    fd = sys.stdin.fileno()
    old = termios.tcgetattr(fd)
    try:
        tty.setraw(fd)
        while not quit_flag.is_set():
            r, _, _ = select.select([sys.stdin], [], [], 0.05)
            if r:
                ch = sys.stdin.read(1)
                if ch in last_seen:
                    last_seen[ch] = time.monotonic()
                elif ch in ("q", "\x1b", "\x03"):  # q, Esc, Ctrl-C
                    quit_flag.set()
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old)


def held(key):
    return (time.monotonic() - last_seen[key]) < HELD_WINDOW


def main():
    port = find_port()
    ser = serial.Serial(port, 115200, timeout=0.1)
    print(f"Connected to {port}")
    print("Hold w = forward  |  Hold s = reverse  |  Release = stop  |  q/Esc = quit")

    threading.Thread(target=key_reader, daemon=True).start()

    time.sleep(0.5)
    while ser.in_waiting:
        print(ser.readline().decode(errors="replace"), end="")

    speed = 0.0
    last_sent = None

    while not quit_flag.is_set():
        w, s = held("w"), held("s")

        if w and not s:
            target = 255
        elif s and not w:
            target = -255
        else:
            target = 0

        speed = 0.05 * target + 0.95 * speed

        out = int(round(speed))
        if out != last_sent:
            ser.write(f"{out}\n".encode())
            last_sent = out
            print(f"\rspeed: {out:+4d}  ", end="", flush=True)

        while ser.in_waiting:
            print("\n" + ser.readline().decode(errors="replace"), end="")

        time.sleep(TICK_S)

    print("\nStopping...")
    ser.write(b"0\n")
    ser.close()


if __name__ == "__main__":
    main()
