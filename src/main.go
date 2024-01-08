package main

import (
	"fmt"
	"math"
	"os"
	"sync"
)

func main() {
	game := NewGame()

	game.LoadCreatures()
	game.DebugCreatures()

	var once sync.Once
	var onceDrone sync.Once
	startQueueLen := 0

	for {
		s := game.LoadState()
		// s.DebugRadar()
		s.DebugVisibleCreatures()

		once.Do(game.DebugLightCoutns)
		onceDrone.Do(func() {
			for i := range s.MyDrones {
				drone := s.MyDrones[i]
				posX := drone.X
				deltaX := 0
				if drone.X < MaxPosistionX/2 {
					if drone.ID%2 == 0 {
						deltaX = 2000
					} else {
						deltaX = 1000
					}
				} else {
					if drone.ID%2 == 0 {
						deltaX = -1000
					} else {
						deltaX = -2000
					}
				}
				// if drone.X < MaxPosistionX/2 {
				// posX = 2000
				// } else {
				// posX = 8000
				// }
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
				game.MoveDrone(drone, Point{X: posX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX + deltaX, Y: MaxDepthInPath})
				game.MoveDrone(drone, Point{X: posX, Y: ResurfaceDistance})
			}
			startQueueLen = len(game.DroneQueue[0])
		})

		leftDrone, rightDrone := s.MyDrones[0].ID, s.MyDrones[1].ID
		if s.MyDrones[0].X > s.MyDrones[1].X {
			leftDrone, rightDrone = s.MyDrones[1].ID, s.MyDrones[0].ID
		}
		hashTarget := map[int]struct{}{}

		game.DroneTarget2 = make(map[int][]int)
		for _, r := range s.Radar {
			creature := game.GetCreature(r.CreatureID)
			drone := s.GetDrone(r.DroneID)

			if _, ok := s.MyCreatures[r.CreatureID]; ok {
				continue
			}
			var scanned bool
			for _, k := range s.MyDrones {
				if len(game.DroneQueue[k.ID]) == startQueueLen {
					continue
				}
				if k.DetectMode() == ModeType0 {
					delete(game.DroneTarget2, drone.ID)
					continue
				}
				if k.IsEmergency() {
					continue
				}
				if _, ok := s.DroneScnas[k.ID][r.CreatureID]; ok {
					for _, drone := range s.MyDrones {
						if vv, ok := game.DroneTarget2[drone.ID]; ok {
							if len(vv) == 1 && vv[0] == r.CreatureID {
								delete(game.DroneTarget, drone.ID)
							} else {
								for i := range vv {
									if vv[i] == r.CreatureID {
										vv = append(vv[:i], vv[i+1:]...)
										game.DroneTarget2[drone.ID] = vv
										break
									}
								}
							}
						}
					}
					scanned = true
					break
				}
			}
			for _, v := range game.DroneTarget2 {
				for _, vv := range v {
					if r.CreatureID == vv {
						scanned = true
						break
					}
				}
			}
			if _, ok := hashTarget[r.CreatureID]; ok {
				continue
			}
			if game.OnDroneDepth(creature, drone) {
				if scanned {
					fmt.Fprintln(os.Stderr, "skip target", creature.ID, drone.ID)
					continue
				}
				switch r.Radar {
				case RadarTL:
					if v, ok := s.DroneCreatureRadar[rightDrone]; ok {
						if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
							firstPoint := game.FirstCommand(*drone)
							// to resurface
							if firstPoint.Y < MaxPosistionY/2 {
								d := leftDrone
								if v, ok := game.DroneTarget2[d]; !ok {
									game.DroneTarget2[d] = []int{r.CreatureID}
								} else {
									v = append(v, r.CreatureID)
									game.DroneTarget2[d] = v
								}
								hashTarget[r.CreatureID] = struct{}{}
							}
						}
					}
				case RadarTR:
					if v, ok := s.DroneCreatureRadar[leftDrone]; ok {
						if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
							firstPoint := game.FirstCommand(*drone)
							// to resurface
							if firstPoint.Y < MaxPosistionY/2 {
								d := rightDrone
								if v, ok := game.DroneTarget2[d]; !ok {
									game.DroneTarget2[d] = []int{r.CreatureID}
								} else {
									v = append(v, r.CreatureID)
									game.DroneTarget2[d] = v
								}
								hashTarget[r.CreatureID] = struct{}{}
							}
						}
					}
				case RadarBL:
					if v, ok := s.DroneCreatureRadar[rightDrone]; ok {
						if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
							firstPoint := game.FirstCommand(*drone)
							// to bottom
							if firstPoint.Y > MaxPosistionY/2 {
								d := leftDrone
								if v, ok := game.DroneTarget2[d]; !ok {
									game.DroneTarget2[d] = []int{r.CreatureID}
								} else {
									v = append(v, r.CreatureID)
									game.DroneTarget2[d] = v
								}
								hashTarget[r.CreatureID] = struct{}{}
							}
						}
					}
				case RadarBR:
					if v, ok := s.DroneCreatureRadar[leftDrone]; ok {
						if radar, ok := v[r.CreatureID]; ok && radar == r.Radar {
							firstPoint := game.FirstCommand(*drone)
							// to bottom
							if firstPoint.Y > MaxPosistionY/2 {
								d := rightDrone
								if v, ok := game.DroneTarget2[d]; !ok {
									game.DroneTarget2[d] = []int{r.CreatureID}
								} else {
									v = append(v, r.CreatureID)
									game.DroneTarget2[d] = v
								}
								hashTarget[r.CreatureID] = struct{}{}
							}
						}
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "target: %v\nscanned:%d\n", game.DroneTarget2, s.DroneScnas)

		for i := range s.MyDrones {
			drone := s.MyDrones[i]
			if drone.IsEmergency() {
				game.RemoveDroneTarget(drone)
				fmt.Fprintln(os.Stderr, drone.ID, "emergency", s.DroneScnas[drone.ID])
				drone.Wait()
				continue
			}
			for _, s := range s.Creatures {
				gs := game.GetCreature(s.ID)
				if gs.Type < 0 && drone.DistanceToPoint(s.Point()) < AutoScanDistance {
					drone.NearMonster = true
					delete(game.DroneTarget2, drone.ID)
				}
			}
			for _, m := range game.PrevMonster {
				if drone.DistanceToPoint(m.MultNextPoint(2.0)) < AutoScanDistance {
					drone.NearMonster = true
					delete(game.DroneTarget2, drone.ID)
				}
			}
			for _, m := range game.PrevPrevMonster {
				if drone.DistanceToPoint(m.MultNextPoint(3.0)) < AutoScanDistance {
					drone.NearMonster = true
					delete(game.DroneTarget2, drone.ID)
				}
			}

			drone.TurnLight(game)
			drone.SolveRadarRadius(game, s.MapRadar[drone.ID])
			drone.DebugRadarRadius()

			resurface := ResurfaceByScore(game, s)
			newPoint := game.FirstCommand(drone)
			if len(game.DroneQueue[drone.ID]) == startQueueLen && math.Abs(float64(newPoint.Y-drone.Y)) < DroneSize {
				newPoint = game.PopCommand(drone)
				if drone.Y > MaxPosistionY/2 {
					game.DroneQueue[drone.ID][0].X = drone.X
					newPoint.X = drone.X
				}
			} else if resurface {
				newPoint.X = drone.X
				newPoint.Y = ResurfaceDistance
				game.DroneQueue[drone.ID][0].Y = ResurfaceDistance
				fmt.Fprintln(os.Stderr, "resurface by score drone", drone.ID)
				delete(game.DroneTarget2, drone.ID)
			} else if drone.DistanceToPoint(newPoint) < SurfaceDistance {
				newPoint = game.PopCommand(drone)
				if drone.Y > MaxPosistionY/2 {
					game.DroneQueue[drone.ID][0].X = drone.X
					newPoint.X = drone.X
				}
			}
			m := drone.Solve(game, s, s.MapRadar[drone.ID], newPoint)
			if drone.ID == 0 || drone.ID == 3 {
				// DebugLocation(m, drone.ID, newPoint)
			}
			//v := drone.SolveToGraph(game, s, m, newPoint)
			//BFS(v, func(v *Vertex) {
			//	// fmt.Fprintf(os.Stderr, ">>(%d:%d:%d)\n", v.ID.X, v.ID.Y, len(v.Vertices))
			//	for k := range v.Vertices {
			//		v.Vertices[k].Node.Score += v.Node.Score
			//	}
			//})
			path := drone.SolveFillLocation(m, nil)
			if drone.ID == 0 || drone.ID == 3 {
				// DebugLocation(m, drone.ID, newPoint)
			}
			// DebugVertex(v)
			DebugPath(path)

			// drone.MoveByVertex(v)
			// m = drone.MoveByLocation(m, nil)
			p := drone.MoveByLocation2(path)
			drone.Move(p.Point)

			// DebugLocation(m, drone.ID, newPoint)
		}

		for d, t := range game.DroneTarget2 {
			newTarget := []int{}
			for _, targetID := range t {
				if _, ok := s.DroneScnas[d][targetID]; !ok {
					newTarget = append(newTarget, targetID)
				}
			}
			game.DroneTarget2[d] = newTarget
		}

		if s.CreaturesAllScannedOrSaved() {
			for _, drone := range s.MyDrones {
				newPoint := game.FirstCommand(drone)
				if newPoint.Y != ResurfaceDistance && len(game.DroneQueue[drone.ID]) >= 1 {
					game.DroneQueue[drone.ID][0] = Point{X: drone.X, Y: ResurfaceDistance}
				}
			}
		}
		if len(game.PrevMonster) > 0 {
			game.PrevPrevMonster = make([]Creature, 0)
			for _, p := range game.PrevMonster {
				game.PrevPrevMonster = append(game.PrevPrevMonster, p)
			}

		}
		game.PrevMonster = make([]Creature, 0)
		for _, s := range s.Creatures {
			gs := game.GetCreature(s.ID)
			if gs.Type < 0 {
				game.PrevMonster = append(game.PrevMonster, s)
			}
		}
	}
}
