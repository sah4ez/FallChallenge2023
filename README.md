
This challenge is based on a system of leagues.
For this challenge, multiple leagues for the same game will be available. Once you have proven your worth against the first Boss, you will access the higher league and unlock new opponents.

  Goal
Win more points than your opponent by scanning the most fish.

To protect marine life, it is crucial to understand it. Explore the ocean floor using your to scan as many fish as possible to better understand them!
  Rules

The game is played turn by turn. Each turn, each player gives an action for their drone to perform.
The Map

The map is a square of 10,000 units on each side. Length units will be denoted as "u" in the rest of the statement. The coordinate (0, 0) is located at the top left corner of the map.
Drones

Each player has to explore the ocean floor and scan the fish. Each turn, the player can decide to move their drone in a direction or not activate its motors.

Your drone continuously emits light around it. If a fish is within this light radius, it is automatically scanned. You can increase the power of your light (and thus your scan radius), but this will drain your battery.
Fish

On the map, different fish are present. Each fish has a specific type and color. In addition to the points earned if you scan a fish and bring the scan back to the surface, bonuses will be awarded if you scan all the fish of the same type or same color, or if you are the first to do so.
Unit Details
Drones

Drones move towards the given point, with a maximum distance per turn of 600u. If the motors are not activated in a turn, the drone will sink by 300u.

At the end of the turn, fish within a radius of 800u will be automatically scanned.

If you have increased the power of your light, this radius becomes 2000u, but the battery drains by 5 points. If the powerful light is not activated, the battery recharges by 1. The battery has a capacity of 30 and is fully charged at the beginning of the game.
Score Details

Points are awarded for each scan depending on the type of scanned fish. Being the first to perform a scan or a combination allows you to earn double the points.
Scan 	Points 	Points if first to scan
Type 0 	1 	2
Type 1 	2 	4
Type 2 	3 	6
All fish of one color 	3 	6
All fish of one type 	4 	8
Victory Conditions

    The game reaches 200 turns
    A player has earned enough points that their opponent cannot catch up
    Both players have saved the scans of all remaining fish on the map

Defeat Conditions

    Your program does not respond within the given time or provides an unrecognized command.


🐞 Debugging Tips

    Hover over an entity to see more information about it.
    Add text at the end of an instruction to display that text above your drone.
    Click on the gear icon to display additional visual options.
    Use the keyboard to control actions: space for play/pause, arrows for step-by-step forward movement.

  Game Protocol
Initialization Input
First line: creatureCount an integer for the number of creatures in the game zone. Will always be 12.
Next creatureCount lines: 3 integers describing each creature:

    creatureId for this creature's unique id.
    color (0 to 3) and type (0 to 2).

Input for One Game Turn
myScore for you current score.
foeScore for you opponent's score.

myScanCount for your amount of scans
Next myScanCount lines: creatureId for each scan.

foeScanCount for your opponent's amount of scans.
Next foeScanCount lines: creatureId for each scan of your opponent.

For your drone:

    droneId: this drone's unique id.
    droneX and droneY: this drone's position.
    battery: this drone's current battery level. 

Next, for your opponent's drone:

    droneId: this drone's unique id.
    droneX and droneY: this drone's position.
    battery: this drone's current battery level. 


For every fish:

    creatureId: this creature's unique id.
    creatureX and creatureY: this creature's position.
    creatureVx and creatureVy: this creature's current speed.

The rest of the variables can be ignored and will be used in later leagues.
Output
One line: one valid instruction for your drone:

    MOVE x y light: makes the drone move towards (x,y), engines on.
    WAIT light. Switches engines off. The drone will sink but can still use light to scan nearby creatures.

Set light to 1 to activate the powerful light, 0 otherwise.
Constraints
creatureCount = 12 in this league
myDroneCount = 1 in this league

Response time per turn ≤ 50ms
Response time for the first turn ≤ 1000ms

