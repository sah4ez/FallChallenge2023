package main

import "fmt"

// MOVE <x> <y> <light (1|0)> | WAIT <light (1|0)>
type Drone struct {
	ID        int
	X         int
	Y         int
	Emergency int
	Battery   int

	enabledLight bool
}

func (d *Drone) TurnLight() {
	if d.enabledLight {
		d.enabledLight = false
		return
	}
	d.enabledLight = true
}

func (d *Drone) Light() string {

	if d.enabledLight {
		return "1"
	}
	return "0"
}

func (d *Drone) Wait() {
	fmt.Printf("WAIT %s\n", d.Light())
}

func (d *Drone) Move(x, y int) {
	fmt.Printf("MOVE %d %d %s\n", x, y, d.Light())
}

func (d *Drone) Debug() {

}

func NewDrone() Drone {
	d := Drone{}
	fmt.Scan(&d.ID, &d.X, &d.Y, &d.Emergency, &d.Battery)
	return d
}
