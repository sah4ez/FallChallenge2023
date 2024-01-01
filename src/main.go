package main

import (
	"fmt"
	"os"
)

func main() {
	game := NewGame()

	game.LoadCreatures()
	game.DebugCreatures()

	for {
		s := game.LoadState()
		s.DebugRadar()
		s.DebugVisibleCreatures()

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

			newPoint, dist, cID := drone.FindNearCapture(game, s)

			if len(s.DroneScnas[drone.ID]) >= 1 && drone.NeedSurface(int(dist)) {
				game.AddResurface(drone.ID, Point{X: drone.X, Y: int(SurfaceDistance)})
			} else if len(s.DroneScnas[drone.ID]) >= 6 {
				game.AddResurface(drone.ID, Point{X: drone.X, Y: int(SurfaceDistance)})
			}

			if !newPoint.IsZero() {
				targetCaptureID, ok := game.DroneTarget[drone.ID]
				if dist <= AutoScanDistance && targetCaptureID == cID && ok {
					game.RemoveDroneTarget(drone.ID)
				}
				if dist <= AutoScanDistance {
					game.TouchCreature(drone.ID, cID)
				}
				fmt.Fprintln(os.Stderr, drone.ID, "move to nearest")
				drone.TurnLight(game)
				drone.Move(newPoint)
			} else {
				var hasRadar bool
				s.DebugRadarByDroneID(drone.ID)
				for _, r := range s.MapRadar[drone.ID] {
					if game.IsTouchedCreature(r.CreatureID) {
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
						drone.TurnLight(game)
						drone.MoveToRadar(r.Radar)
						fmt.Fprintln(os.Stderr, drone.ID, "target", cID)
						hasRadar = true
						continue
					} else {
						if !s.CheckCreatureID(drone.ID, cID) {
							game.RemoveDroneTarget(drone.ID)
							fmt.Fprintf(os.Stderr, "%d missing %d clear target\n", drone.ID, r.CreatureID)

						} else {
							fmt.Fprintf(os.Stderr, "%d c %d target map %v\n", drone.ID, r.CreatureID, game.DroneTarget)
						}
					}
				}
				if !hasRadar {
					drone.RandMove()
				}
			}
		}
	}
}
