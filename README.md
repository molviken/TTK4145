# Heisprosjekt - TTK4145
System of three elevators designed to complete all order effciently.

## Introduction
The master / slave network will happen over UDP, where all slaves communicate with the master elevator at the time.  The master elevator is chosen each time a remote order is requested, and is the elevator that first got the order. Because the master must communicate with many slaves at the same time and the master's function is transferable, we use UDP as it provides the most flexibility. We also do not feel the need for the properties we get with TCP vs. UDP in this project. We used a peer function between the lifts to keep track of lifts disappearing from the network / connecting to the network again. This was implemented between all lifts.

## Getting started
Write the following commands in your terminal, in order.
1. To download: "git clone https://www.github.com/molviken/TTK4145"
2. Change folder: "cd TTK4145".
3. Launch program: "./TTK4145 -id=x", where x is the id you want.

### Prerequisites

Git needs to be installed on your computer beforehand.

## Imported libraries
- "encoding/json", used for encoding our orders to bytes.
- "io/outil", used for storing the encoded bytes on a backup file.
- "container/list", all local orders were appended to a linked list for easy usage across all files.
- "fmt", used to print messages.
- "flag", used for keeping a different ID for each elevator in use from different local computers.
- "strconv", used for converting string to int, og int to string.
- "time", required when using a timer.-

## Created modules

Network.go – The task of this module is to handle communication between the elevators over UDP. In addition, it will maintain a peer function between all elevators such that the list of active lifts is kept up to date.

Elevator_io.go - This module controlls the hardware of the elevator.Functions such as start / stop of engine, change direction, turn on lights and listening to sensors are available.

ElevFunc.go – This module contains help functions for controlling the elevator.

Queue.go – The module handles the queue, removing/adding orders, checking for duplicates and calculating cost.

ElevAssigner.go –Remote orders are divided in the module and assigned to one of the elevators online.s

EventHandler.go – This module handles every event that occurs and maintains the flow of the program.
