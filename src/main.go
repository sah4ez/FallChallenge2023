package main

import (
	"fmt"
	"os"
)

func main() {
	game := NewGame()

	game.LoadCreatures()

	for {
		s := game.LoadState()
		s.DebugRadar()

		for i := range s.MyDrones {
			drone := s.MyDrones[i]

			if p, ok := game.Resurface[drone.ID]; ok {
				if !drone.IsSurfaced() {
					drone.Move(p)
					fmt.Fprintln(os.Stderr, drone.ID, "surfaced", drone.X, drone.Y)
					continue
				}
				game.RemoveResurface(drone.ID)
			}

			if len(s.DroneScnas[drone.ID]) >= 1 {
				game.AddResurface(drone.ID, Point{X: drone.X, Y: int(SurfaceDistance)})
			}
			newPoint, dist, cID := drone.FindNearCapture(game, s)
			if !newPoint.IsZero() {
				targetCaptureID, ok := game.DroneTarget[drone.ID]
				if dist <= AutoScanDistance && targetCaptureID == cID && ok {
					game.RemoveDroneTarget(drone.ID)
				}
				if dist <= AutoScanDistance {
					game.TouchCreature(drone.ID, cID)
				}
				fmt.Fprintln(os.Stderr, drone.ID, "move to nearest")
				drone.TurnLight()
				drone.Move(newPoint)
			} else {
				for _, r := range s.Radar {
					var found bool
					for _, cs := range game.CreaturesTouched {
						_, found = cs[r.CreatureID]
					}
					if found {
						continue
					}
					cID, ok := game.DroneTarget[drone.ID]
					if r.DroneID == drone.ID && !ok {
						drone.MoveToRadar(r.Radar)
						game.AddDroneTarget(drone.ID, r.CreatureID)
						fmt.Fprintln(os.Stderr, drone.ID, "no target")
						continue
					} else if r.DroneID == drone.ID && ok && r.CreatureID == cID {
						drone.TurnLight()
						drone.MoveToRadar(r.Radar)
						fmt.Fprintln(os.Stderr, drone.ID, "target", cID)
						continue
					}
				}
				// drone.Wait()
				// s.DebugRadar()
			}
		}
	}
}
