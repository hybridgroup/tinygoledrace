# Racer

This is the basic logic of the racer.

## Game States

The PyPortal should attempt to connect to the MQTT server right away at startup.

### Looking

Display an attract screen. Sends heartbeat mqtt messages.

### Available

Display menu to choose to start playing?

### Ready

Display message "Let's race"

### Starting

Display message "The race is beginning now"

### Countdown

Display countdown message "3", "2", "1", "GO"

### Start

Display racing panel with controls

### Over

Display message "The race is over"

### Winner

Display message "You won!", or else "You lost"
