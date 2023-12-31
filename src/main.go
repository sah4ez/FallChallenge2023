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
		// s.DebugRadar()

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

			if len(s.DroneScnas[drone.ID]) >= 2 {
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
				var hasRadar bool
				s.DebugRadar()
				for _, r := range s.Radar {
					if r.DroneID != drone.ID {
						continue
					}
					if hasRadar {
						break
					}
					var found bool
					for _, cs := range game.CreaturesTouched {
						_, found = cs[r.CreatureID]
					}
					if found {
						continue
					}
					cID, ok := game.DroneTarget[drone.ID]
					if r.DroneID == drone.ID && !ok {
						if added := game.AddDroneTarget(drone.ID, r.CreatureID); !added {
							fmt.Fprintln(os.Stderr, "not added target", len(game.TargetCreatures))
							continue
						}

						drone.MoveToRadar(r.Radar)
						fmt.Fprintln(os.Stderr, drone.ID, "no target")
						hasRadar = true
						continue
					} else if r.DroneID == drone.ID && ok && r.CreatureID == cID {
						drone.TurnLight()
						drone.MoveToRadar(r.Radar)
						fmt.Fprintln(os.Stderr, drone.ID, "target", cID)
						hasRadar = true
						continue
					} else {
						fmt.Fprintf(os.Stderr, "%v\n", game.DroneTarget)
					}
				}
				if !hasRadar {
					drone.Wait()
				}
			}
		}
	}
}
